use crate::error::ApplicationError;
use crate::manager_thread::ManagerThread;
use crate::requests::{
    AxolotlConfig, AxolotlMessage, AxolotlRequest, AxolotlResponse, GetMessagesRequest,
    SendMessageRequest, SendMessageResponse,
};
extern crate dirs;
use futures::channel::oneshot::Receiver;
use futures::executor::block_on;
use futures::lock::Mutex;
use futures::stream::SplitSink;
use futures::{SinkExt, StreamExt};
use libsignal_service::prelude::GroupMasterKey;
use presage::prelude::Content;
use presage::prelude::{ContentBody, DataMessage, GroupContextV2, ServiceAddress};
use serde::{Serialize, Serializer};
use serde_json::error::Error as SerdeError;
use std::cell::Cell;
use std::io::Write;
use std::process::exit;
use std::time::UNIX_EPOCH;
use std::{sync::Arc, thread};
use url::Url;

use warp::filters::ws::Message;
use warp::ws::WebSocket;

const MESSAGE_BOUND: usize = 10;

use futures::channel::oneshot;

use presage::MigrationConflictStrategy;
use presage::{SledStore, Thread};
use tokio::sync::mpsc::{self, UnboundedReceiver};

/// Handles a client connection
pub async fn handle_ws_client(websocket: warp::ws::WebSocket) {
    match start_manager(websocket).await {
        Ok(_) => log::info!("Manager started"),
        Err(e) => {
            log::error!("Error starting the manager: {}", e);
            exit(0);
        }
    }
}

async fn start_manager(websocket: warp::ws::WebSocket) -> Result<(), ApplicationError> {
    let mut count = 0u32;

    let (sender, mut receiver) = websocket.split();
    let shared_sender = Arc::new(Mutex::new(sender));

    let config_path = dirs::config_dir()
        .unwrap()
        .into_os_string()
        .into_string()
        .unwrap();

    loop {
        if count == 5 {
            log::error!("Too many errors, exiting");

            // Exit this loop
            break;
        }
        count += 1;
        log::info!("Setting up the manager");
        let (provisioning_link_tx, provisioning_link_rx) = oneshot::channel();
        let (error_tx, error_rx) = oneshot::channel();

        let (send_content, receive_content) = mpsc::unbounded_channel();
        let (send_error, mut receive_error) = mpsc::channel(MESSAGE_BOUND);

        let shared_sender_mutex = Arc::clone(&shared_sender);
        let shared_sender_mutex2 = Arc::clone(&shared_sender);
        let current_chat: Option<Thread> = None;
        let current_chat_mutex = Arc::new(Mutex::new(current_chat));
        thread::spawn(move || {
            block_on(handle_provisoning(
                provisioning_link_rx,
                shared_sender_mutex,
            ))
        });
        thread::spawn(move || {
            block_on(handle_received_message(
                receive_content,
                shared_sender_mutex2,
            ))
        });
        let db_path = format!("{config_path}/textsecure.nanuc");

        log::info!("Opening the database at {}", db_path);
        let config_store = match SledStore::open_with_passphrase(
            db_path,
            None::<&str>,
            MigrationConflictStrategy::BackupAndDrop,
        )
        .ok()
        {
            Some(store) => store,
            None => {
                log::info!("Failed to open the database");
                continue;
            }
        };
        let manager = ManagerThread::new(
            config_store,
            "axolotl".to_string(),
            provisioning_link_tx,
            error_tx,
            send_content,
            current_chat_mutex,
            send_error,
        )
        .await;

        if manager.is_none() {
            if let Some(error_opt) = receive_error.recv().await {
                log::error!("Got error after linking device: {}", error_opt);
                continue;
            }
        }
        log::info!("Awaiting for error linking");
        match error_rx.await {
            Ok(err) => {
                log::error!("Got error linking device: {}", err);
                continue;
            }
            Err(_e) => log::info!("Manager setup successfull"),
        }

        // Anounce to client, that registartion is done
        let shared_sender_mutex = Arc::clone(&shared_sender);
        let mut mut_sender = shared_sender_mutex.lock().await;
        match mut_sender
            .send(Message::text("{\"Type\":\"registrationDone\"}"))
            .await
        {
            Ok(_) => log::info!("Sent registration done message"),
            Err(e) => log::error!("Error sending registration done message: {}", e),
        };
        std::mem::drop(mut_sender);
        let m = manager.unwrap();
        log::info!("Manager created");
        // While messages come, handle them
        while let Some(body) = receiver.next().await {
            let message = match body {
                Ok(msg) => msg,
                Err(_) => {
                    continue;
                }
            };
            let shared_sender_mutex = Arc::clone(&shared_sender);
            log::debug!("Got websocket message from axolotl-web: {:?}", message);
            match handle_websocket_message(message, shared_sender_mutex, &m).await {
                Ok(_) => log::info!("Message handled"),
                Err(e) => log::error!("Error handling message: {}", e),
            };
        }
    }
    let shared_sender_mutex = Arc::clone(&shared_sender);

    shared_sender_mutex.lock().await.close().await?;
    Err(ApplicationError::WebSocketHandleMessageError(
        "Too many errors, exiting".to_string(),
    ))
}
async fn handle_provisoning(
    provisioning_link_rx: Receiver<Url>,
    sender: Arc<futures::lock::Mutex<SplitSink<WebSocket, warp::ws::Message>>>,
) {
    log::info!("Awaiting for provisioning link");
    match provisioning_link_rx.await {
        Ok(url) => {
            log::info!("Manager wants to show QR code, emitting signal");
            let qr_code = format!("{{\"response_type\":\"qr_code\",\"data\":\"{}\"}}", url);
            let mut ws_sender = sender.lock().await;
            ws_sender.send(Message::text(qr_code)).await.unwrap();
        }
        Err(_e) => log::trace!("Manager is already linked"),
    }
}
async fn handle_received_message(
    // manager: &ManagerThread,
    mut receive: UnboundedReceiver<Content>,
    sender: Arc<futures::lock::Mutex<SplitSink<WebSocket, warp::ws::Message>>>,
) {
    log::info!("Awaiting for received message");
    loop {
        match receive.recv().await {
            Some(content) => {
                log::info!("Got message from manager");
                let thread = Thread::try_from(&content).unwrap();
                let mut axolotl_message = AxolotlMessage::from_message(content);
                axolotl_message.thread_id = Some(thread.to_string());
                let axolotl_message_json = serde_json::to_string(&axolotl_message).unwrap();
                let response_type = "message_received".to_string();
                let response =  AxolotlResponse {
                    response_type,
                    data: axolotl_message_json,
                };
                let response = serde_json::to_string(&response).unwrap();

                let mut ws_sender = sender.lock().await;
                ws_sender.send(Message::text(response)).await.unwrap();
                std::mem::drop(ws_sender);
            }
            None => {
                log::error!("Error receiving message");
                break;
            }
        };
    }
}
pub fn write_as_json<I, P, W>(out: &mut W, groups: I) -> serde_json::Result<()>
where
    I: IntoIterator<Item = P>,
    P: Serialize,
    W: Write,
{
    struct Wrapper<T>(Cell<Option<T>>);

    impl<I, P> Serialize for Wrapper<I>
    where
        I: IntoIterator<Item = P>,
        P: Serialize,
    {
        fn serialize<S: Serializer>(&self, s: S) -> Result<S::Ok, S::Error> {
            s.collect_seq(self.0.take().unwrap())
        }
    }

    serde_json::to_writer_pretty(out, &Wrapper(Cell::new(Some(groups))))
}
async fn handle_get_contacts(manager: &ManagerThread) -> Result<AxolotlResponse, ApplicationError> {
    log::info!("Getting contacts");
    manager.update_cotacts_from_profile().await.ok();
    let contacts = manager.get_contacts().await.ok().unwrap();
    let mut out = Vec::new();
    write_as_json(&mut out, contacts)?;
    let response = AxolotlResponse {
        response_type: "contact_list".to_string(),
        data: String::from_utf8(out).unwrap(),
    };
    Ok(response)
}
async fn handle_chat_list(manager: &ManagerThread) -> Result<AxolotlResponse, ApplicationError> {
    log::info!("Getting chat list");
    let conversations = manager.get_conversations().await.ok().unwrap();
    let mut out = Vec::new();
    write_as_json(&mut out, conversations)?;
    let response = AxolotlResponse {
        response_type: "chat_list".to_string(),
        data: String::from_utf8(out).unwrap(),
    };
    Ok(response)
}
async fn handle_get_message_list(
    manager: &ManagerThread,
    data: Option<String>,
) -> Result<AxolotlResponse, ApplicationError> {
    log::info!("Getting message list");
    let data = data.ok_or(ApplicationError::InvalidRequest)?;
    if let Ok::<GetMessagesRequest, SerdeError>(messages_request) = serde_json::from_str(&data) {
        let thread: Thread = Thread::try_from(&messages_request.id).unwrap();
        match thread {
            Thread::Contact(_) => {
                manager.update_cotacts_from_profile().await.ok().unwrap();
            }
            _ => {}
        }
        let messages = manager.get_messages(thread, None).await.ok().unwrap();
        let mut axolotl_messages: Vec<AxolotlMessage> = Vec::new();
        for message in messages {
            axolotl_messages.push(AxolotlMessage::from_message(message));
        }
        let mut out = Vec::new();
        write_as_json(&mut out, axolotl_messages)?;
        let response = AxolotlResponse {
            response_type: "message_list".to_string(),
            data: String::from_utf8(out).unwrap(),
        };
        Ok(response)
    } else {
        Err(ApplicationError::InvalidRequest)
    }
}

fn handle_ping() -> Result<AxolotlResponse, ApplicationError> {
    log::info!("Got ping");
    let response = AxolotlResponse {
        response_type: "pong".to_string(),
        data: "".to_string(),
    };
    Ok(response)
}
async fn handle_send_message(
    manager: &ManagerThread,
    data: Option<String>,
) -> Result<AxolotlResponse, ApplicationError> {
    log::info!("Sending message");
    let data = data.ok_or(ApplicationError::InvalidRequest)?;
    match serde_json::from_str(&data) {
        Ok::<SendMessageRequest, SerdeError>(send_message_request) => {
            let timestamp = std::time::SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .expect("Time went backwards")
                .as_millis() as u64;
            let data_message = DataMessage {
                body: Some(send_message_request.text),
                timestamp: Some(timestamp),
                ..Default::default()
            };
            let thread = match Thread::try_from(&send_message_request.recipient) {
                Ok(t) => t,
                Err(e) => {
                    log::error!("Error while parsing the request. {:?}", e);
                    return Err(ApplicationError::InvalidRequest);
                }
            };
            let result = match thread {
                Thread::Contact(contact) => {
                    let message = ContentBody::DataMessage(data_message.clone());
                    manager
                        .send_message(contact, message.clone(), timestamp)
                        .await
                }
                Thread::Group(group) => {
                    let group_master_key = GroupMasterKey::new(group.clone());
                    let group_from_store =
                        manager.get_group_v2(group_master_key).await.ok().unwrap();
                    let group_members = group_from_store.members.iter();
                    let mut group_members_service_addresses: Vec<ServiceAddress> = Vec::new();

                    for member in group_members {
                        group_members_service_addresses.push(ServiceAddress {
                            uuid: Some(member.uuid.clone()),
                            phonenumber: None,
                            relay: None,
                        });
                    }
                    let mut group_data_message = data_message.clone();
                    group_data_message.group_v2 = Some(GroupContextV2 {
                        master_key: Some(group.to_vec()),
                        group_change: None,
                        revision: Some(group_from_store.version),
                    });
                    manager
                        .send_message_to_group(
                            group_members_service_addresses,
                            group_data_message,
                            timestamp,
                        )
                        .await
                }
            };
            let is_failed = result.is_err();
            if is_failed {
                log::error!("Error while sending the message. {:?}", result.err());
            }
            let mut message = AxolotlMessage::from_data_message(data_message);
            message.thread_id = Some(thread.to_string());
            message.sender = Some(manager.uuid());
            let response_data = SendMessageResponse { message, is_failed };
            let response_data_json = serde_json::to_string(&response_data).unwrap();
            let response = AxolotlResponse {
                response_type: "message_sent".to_string(),
                data: response_data_json,
            };
            Ok(response)
        }
        Err(e) => {
            log::error!("Error while parsing the request. {:?}", e);
            Err(ApplicationError::InvalidRequest)
        }
    }
}
async fn handle_get_config(manager: &ManagerThread) -> Result<AxolotlResponse, ApplicationError> {
    log::info!("Getting config");
    let my_uuid = manager.uuid();
    let platform = "linux".to_string();
    #[cfg(target_os = "windows")]
    {
        let platform = "windows".to_string();
    }
    #[cfg(target_os = "macos")]
    {
        let platform = "macos".to_string();
    }
    #[cfg(target_os = "android")]
    {
        let platform = "android".to_string();
    }
    #[cfg(target_os = "ios")]
    {
        let platform = "ios".to_string();
    }
    let feature = "".to_string();
    #[cfg(feature = "tauri")]
    {
        let feature = "tauri".to_string();
    }
    #[cfg(feature = "ut")]
    {
        let feature = "ut".to_string();
    }

    let config = AxolotlConfig {
        uuid: Some(my_uuid.to_string()),
        e164: None,
        platform: Some(platform),
        feature: Some(feature),
    };
    let response = AxolotlResponse {
        response_type: "config".to_string(),
        data: serde_json::to_string(&config).unwrap(),
    };
    Ok(response)
}
/// Handles a websocket message
async fn handle_websocket_message(
    message: Message,
    mutex_sender: Arc<futures::lock::Mutex<SplitSink<WebSocket, warp::ws::Message>>>,
    manager: &ManagerThread,
) -> Result<(), ApplicationError> {
    // Skip any non-Text messages...
    let msg = if let Ok(s) = message.to_str() {
        s
    } else {
        "Invalid message"
    };
    log::debug!("Got message: {}", msg);
    struct Wrapper<T>(Cell<Option<T>>);

    impl<I, P> Serialize for Wrapper<I>
    where
        I: IntoIterator<Item = P>,
        P: Serialize,
    {
        fn serialize<S: Serializer>(&self, s: S) -> Result<S::Ok, S::Error> {
            s.collect_seq(self.0.take().unwrap())
        }
    }
    // Check the type of request
    if let Ok::<AxolotlRequest, SerdeError>(axolotl_request) = serde_json::from_str(msg) {
        // Axolotl request
        log::info!("Axolotl request: {}", axolotl_request.request.as_str());

        let response = match axolotl_request.request.as_str() {
            "getContacts" => handle_get_contacts(manager).await,
            "getChatList" => handle_chat_list(manager).await,
            "getMessageList" => handle_get_message_list(manager, axolotl_request.data).await,
            "ping" => handle_ping(),
            "sendMessage" => handle_send_message(manager, axolotl_request.data).await,
            "getConfig" => handle_get_config(manager).await,
            _ => {
                log::error!("Unhandled axolotl request {}", axolotl_request.request);
                Err(ApplicationError::InvalidRequest)
            }
        };
        match response {
            Ok(response) => {
                let mut unlocked_sender = mutex_sender.lock().await;
                unlocked_sender
                    .send(Message::text(serde_json::to_string(&response).unwrap()))
                    .await
                    .unwrap();

                std::mem::drop(unlocked_sender);
            }
            Err(e) => {
                log::error!("Error while handling request. {:?}", e);
            }
        }
    } else {
        // Error or unhandled request
        log::error!("Unhandled request {}", msg);
    }
    Ok(())
    //sender.send(Message::text("working")).await.unwrap();
}
