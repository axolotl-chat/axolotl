use crate::error::ApplicationError;
use crate::manager_thread::ManagerThread;
use crate::requests::{
    AxolotlMessage, AxolotlRequest, AxolotlResponse, GetMessagesRequest,
    SendMessageRequest,
};
extern crate dirs;
use futures::channel::oneshot::Receiver;
use futures::executor::block_on;
use futures::lock::Mutex;
use futures::stream::SplitSink;
use futures::{SinkExt, StreamExt};
use presage::prelude::{ContentBody, DataMessage};
use serde::{Serialize, Serializer};
use serde_json::error::Error as SerdeError;
use std::cell::Cell;
use std::io::Write;
use std::time::UNIX_EPOCH;
use std::{sync::Arc, thread};
use url::Url;

use warp::filters::ws::Message;
use warp::ws::WebSocket;

const MESSAGE_BOUND: usize = 10;

use futures::channel::oneshot;

use presage::MigrationConflictStrategy;
use presage::{SledStore, Thread};
use tokio::sync::mpsc;

/// Handles a client connection
pub async fn handle_ws_client(websocket: warp::ws::WebSocket) {
    match start_manager(websocket).await {
        Ok(_) => log::info!("Manager started"),
        Err(e) => log::error!("Error starting the manager: {}", e),
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

        let (send_content, _receive_content) = mpsc::unbounded_channel();
        let (send_error, mut receive_error) = mpsc::channel(MESSAGE_BOUND);

        let shared_sender_mutex = Arc::clone(&shared_sender);
        thread::spawn(move || {
            block_on(handle_provisoning(
                provisioning_link_rx,
                shared_sender_mutex,
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
        let mut unlocked_sender = mutex_sender.lock().await;

        match axolotl_request.request.as_str() {
            "getContacts" => {
                log::info!("Getting contacts");
                manager.update_cotacts_from_profile().await.ok();
                let contacts = manager.get_contacts().await.ok().unwrap();
                let mut out = Vec::new();
                write_as_json(&mut out, contacts)?;
                let response = AxolotlResponse {
                    response_type: "contact_list".to_string(),
                    data: String::from_utf8(out).unwrap(),
                };
                unlocked_sender
                    .send(Message::text(serde_json::to_string(&response).unwrap()))
                    .await
                    .unwrap();
            }
            "getChatList" => {
                log::debug!("Getting chats");
                let conversations = manager.get_conversations().await.ok().unwrap();
                let mut out = Vec::new();
                write_as_json(&mut out, conversations)?;
                let response = AxolotlResponse {
                    response_type: "chat_list".to_string(),
                    data: String::from_utf8(out).unwrap(),
                };
                unlocked_sender
                    .send(Message::text(serde_json::to_string(&response).unwrap()))
                    .await
                    .unwrap();
            }
            "getMessageList" => {
                log::debug!("Getting messageList");
                if axolotl_request.data.is_some() {
                    log::info!("Getting messages");
                    let data = axolotl_request.data.unwrap();
                    if let Ok::<GetMessagesRequest, SerdeError>(messages_request) =
                        serde_json::from_str(&data)
                    {
                        let thread: Thread = Thread::try_from(&messages_request.id).unwrap();
                        let messages = manager.get_messages(thread, None).await.ok().unwrap();
                        let mut axolotl_messages: Vec<AxolotlMessage> = Vec::new();
                        for message in messages {
                            axolotl_messages.push(AxolotlMessage::from_message(message));
                        }
                        let mut out = Vec::new();
                        log::info!("Writing messages {:?}", axolotl_messages);
                        write_as_json(&mut out, axolotl_messages)?;
                        let response = AxolotlResponse {
                            response_type: "message_list".to_string(),
                            data: String::from_utf8(out).unwrap(),
                        };
                        unlocked_sender
                            .send(Message::text(serde_json::to_string(&response).unwrap()))
                            .await
                            .unwrap();
                    }
                } else {
                    log::info!("No id for getMessageList {:?}", axolotl_request.data);
                }
            }
            "ping" => {
                let ping = AxolotlResponse {
                    response_type: "pong".to_string(),
                    data: "".to_string(),
                };
                unlocked_sender
                    .send(Message::text(serde_json::to_string(&ping).unwrap()))
                    .await
                    .unwrap();
            }
            "sendMessage" => {
                log::debug!("Sending a message");
                if axolotl_request.data.is_some() {
                    let data = axolotl_request.data.unwrap();
                    if let Ok::<SendMessageRequest, SerdeError>(send_message_request) =
                        serde_json::from_str(&data)
                    {
                        let timestamp = std::time::SystemTime::now()
                            .duration_since(UNIX_EPOCH)
                            .expect("Time went backwards")
                            .as_millis() as u64;

                        let message = ContentBody::DataMessage(DataMessage {
                            body: Some(send_message_request.text),
                            timestamp: Some(timestamp),
                            ..Default::default()
                        });
                        let result = manager
                            .send_message(send_message_request.recipient, message, timestamp)
                            .await;

                        if let Err(e) = result {
                            log::error!("Error while sending the message. {:?}", e);
                        }
                    }
                }
            }
            _ => {
                log::error!("Unhandled axolotl request {}", axolotl_request.request);
            }
        }
        std::mem::drop(unlocked_sender);
    } else {
        // Error or unhandled request
        log::error!("Unhandled request {}", msg);
    }
    Ok(())
    //sender.send(Message::text("working")).await.unwrap();
}
