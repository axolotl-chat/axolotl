use futures::{select, FutureExt, StreamExt};
use presage::libsignal_service::prelude::AttachmentIdentifier;
use presage::libsignal_service::{groups_v2::Group, sender::AttachmentUploadError};
use presage::{
    prelude::{content::*, AttachmentSpec, Contact, ContentBody, DataMessage, ServiceAddress, *},
    Error, Manager, Registered,
};
use presage::{GroupMasterKeyBytes, Thread, ThreadMetadata, ThreadMetadataMessageContent};
use std::ops::Bound;
use std::sync::Arc;
use tokio::sync::Mutex;
type PresageError = presage::Error<presage_store_sled::SledStoreError>;
use presage::Store;
use presage_store_sled::SledStore;

use serde::{Deserialize, Serialize};
use tokio::sync::{mpsc, oneshot};

use crate::error::ApplicationError;
use crate::handlers;

#[cfg(feature = "ut")]
use dbus::arg::{AppendAll, ReadAll};
#[cfg(feature = "ut")]
use dbus::blocking::Connection;
#[cfg(feature = "ut")]
use serde_json::{json, Value};
#[cfg(feature = "ut")]
use std::time::Duration;

const MESSAGE_BOUND: usize = 10;

enum Command {
    RequestContactsSync(oneshot::Sender<Result<(), PresageError>>),
    GetContacts(oneshot::Sender<Result<Vec<Contact>, PresageError>>),
    GetGroup(
        GroupMasterKeyBytes,
        oneshot::Sender<Result<Option<Group>, PresageError>>,
    ),
    SendMessage(
        ServiceAddress,
        Box<ContentBody>,
        u64,
        oneshot::Sender<Result<(), PresageError>>,
    ),
    SendMessageToGroup(
        GroupMasterKeyBytes,
        Box<DataMessage>,
        u64,
        oneshot::Sender<Result<(), PresageError>>,
    ),
    GetConversations(oneshot::Sender<Result<Vec<ThreadMetadata>, PresageError>>),
    GetAttachment(
        AttachmentPointer,
        oneshot::Sender<Result<Vec<u8>, PresageError>>,
    ),
    UploadAttachments(
        Vec<(AttachmentSpec, Vec<u8>)>,
        oneshot::Sender<
            Result<Vec<Result<AttachmentPointer, AttachmentUploadError>>, PresageError>,
        >,
    ),
    GetMessages(
        Thread,
        (Bound<u64>, Bound<u64>),
        oneshot::Sender<
            Result<<presage_store_sled::SledStore as presage::Store>::MessagesIter, PresageError>,
        >,
    ),
    ThreadMetadata(
        Thread,
        oneshot::Sender<Result<Option<ThreadMetadata>, PresageError>>,
    ),
    SaveThreadMetadata(ThreadMetadata, oneshot::Sender<Result<(), PresageError>>),
    RequestContactsUpdateFromProfile(oneshot::Sender<Result<(), PresageError>>),
    RequestContactUpdateFromProfile(Uuid, oneshot::Sender<Result<Contact, PresageError>>),
}

impl std::fmt::Debug for Command {
    fn fmt(&self, _f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        Ok(())
    }
}

pub struct ManagerThread {
    command_sender: mpsc::Sender<Command>,
    contacts: Arc<Mutex<Vec<Contact>>>,
    sessions: Arc<Mutex<Vec<AxolotlSession>>>,
    current_chat: Arc<Mutex<Option<Thread>>>,
}
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct AxolotlSession {
    pub id: Thread,
    pub last_message: Option<String>,
    pub unread_messages_count: usize,
    pub is_group: bool,
    pub title: Option<String>,
    pub last_message_timestamp: u64,
    pub muted: bool,
}
impl TryFrom<ThreadMetadata> for AxolotlSession {
    type Error = PresageError;

    fn try_from(session: ThreadMetadata) -> Result<Self, PresageError> {
        let timestamp: u64 = match &session.last_message {
            Some(message) => message.timestamp,
            None => 0,
        };
        let message: Option<String> = session.last_message.as_ref().map(|message| match &message.message {
                Some(message) => message.to_string(),
                None => String::new(),
            });
        let is_group: bool = match session.thread {
            Thread::Group(_group) => true,
            _ => false,
        };

        Ok(Self {
            id: session.thread,
            last_message: message,
            unread_messages_count: session.unread_messages_count,
            is_group,
            title: session.title,
            last_message_timestamp: timestamp,
            muted: session.muted,
        })
    }
}

impl Clone for ManagerThread {
    fn clone(&self) -> Self {
        Self {
            command_sender: self.command_sender.clone(),
            contacts: self.contacts.clone(),
            sessions: self.sessions.clone(),
            current_chat: self.current_chat.clone(),
        }
    }
}

impl ManagerThread {
    pub async fn new(
        config_store: presage_store_sled::SledStore,
        device_name: String,
        link_callback: futures::channel::oneshot::Sender<url::Url>,
        error_callback: futures::channel::oneshot::Sender<PresageError>,
        content: mpsc::UnboundedSender<Content>,
        current_chat: Arc<Mutex<Option<Thread>>>,
        error: mpsc::Sender<ApplicationError>,
    ) -> Option<Self> {
        let (sender, receiver) = mpsc::channel(MESSAGE_BOUND);
        std::thread::spawn(move || {
            let error_clone = error.clone();
            let panic = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                tokio::runtime::Runtime::new()
                    .expect("Failed to setup runtime")
                    .block_on(async move {
                        let setup =
                            setup_manager(config_store.clone(), device_name, link_callback).await;
                        if let Ok(mut manager) = setup {
                            log::info!("Starting command loop");
                            drop(error_callback);
                            command_loop(&mut manager, receiver, content, error).await;
                        } else {
                            let e = setup.err().unwrap();
                            log::info!("Got error: {}", e);
                            error_callback.send(e).expect("Failed to send error")
                        }
                    });
            }));
            if let Err(_e) = panic {
                log::info!("Manager-thread paniced");
                tokio::runtime::Runtime::new()
                    .expect("Failed to setup runtime")
                    .block_on(async move {
                        error_clone
                            .send(ApplicationError::ManagerThreadPanic)
                            .await
                            .expect("Failed to send error");
                    });
            }
        });

        let (sender_contacts, receiver_contacts) = oneshot::channel();
        if sender
            .send(Command::GetContacts(sender_contacts))
            .await
            .is_err()
        {
            return None;
        }
        let contacts = receiver_contacts.await;

        if contacts.is_err() {
            return None;
        }

        if let Err(_e) = &contacts.as_ref().unwrap() {
            // TODO: Error handling
            log::info!("Could not load contacts");
        }
        Some(Self {
            command_sender: sender,
            contacts: Arc::new(Mutex::new(contacts.unwrap().unwrap_or_default())),
            sessions: Arc::new(Mutex::new(Vec::new())),
            current_chat,
        })
    }
}

impl ManagerThread {
    pub async fn sync_contacts(&mut self) -> Result<(), PresageError> {
        let (sender_contacts, receiver_contacts) = oneshot::channel();
        self.command_sender
            .send(Command::GetContacts(sender_contacts))
            .await
            .expect("Command sending failed");
        let contacts = receiver_contacts
            .await
            .expect("Callback receiving failed")?;
        let mut c = self.contacts.lock().await;
        *c = contacts;
        log::info!("Synced contacts. Got {} contacts.", c.len());
        Ok(())
    }
    pub async fn request_contacts_sync(&self) -> Result<(), PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::RequestContactsSync(sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }

    pub async fn current_chat(&self) -> Option<Thread> {
        self.current_chat.lock().await.clone()
    }

    pub async fn get_contacts(&self) -> Result<impl Iterator<Item = Contact> + '_, PresageError> {
        let c = self.contacts.lock().await;
        // Very weird way to counteract "returning borrowed c".
        log::info!("Contacts not implemented yet");
        Ok(c.iter()
            .map(almost_clone_contact)
            .collect::<Vec<_>>()
            .into_iter())
    }
    pub async fn update_contacts_from_profile(&self) -> Result<(), PresageError> {
        log::debug!("Updating contacts from profile -> todo");
        // let (sender, receiver) = oneshot::channel();
        // self.command_sender
        //     .send(Command::RequestContactsUpdateFromProfile(sender))
        //     .await
        //     .expect("Command sending failed");
        // receiver.await.expect("Callback receiving failed")
        Ok(())
    }
    pub async fn update_contact_from_profile(&self, id: Uuid) -> Result<Contact, PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::RequestContactUpdateFromProfile(id, sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }

    pub async fn get_conversations(
        &self,
    ) -> Result<impl Iterator<Item = AxolotlSession> + '_, ApplicationError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::GetConversations(sender))
            .await
            .expect("Command sending failed");
        match receiver.await {
            Ok(Ok(sessions)) => {
                log::info!("Got {} sessions", sessions.len());
                let mut axolotl_sessions = sessions
                    .into_iter()
                    .map(|s| AxolotlSession::try_from(s).unwrap())
                    .collect::<Vec<_>>();
                // sort by timestamp
                axolotl_sessions.sort_by_key(|f| f.last_message_timestamp);
                axolotl_sessions.reverse();
                Ok(axolotl_sessions.into_iter())
            }
            Ok(Err(e)) => {
                log::error!("Loading coversations failed: {}", e);
                Err(ApplicationError::ManagerThreadPanic)
            }
            Err(_) => Err(ApplicationError::ManagerThreadPanic),
        }
    }
    pub async fn thread_metadata(
        &self,
        thread: &Thread,
    ) -> Result<Option<ThreadMetadata>, PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::ThreadMetadata(thread.clone(), sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }
    pub async fn save_thread_metadata(&self, metadata: ThreadMetadata) -> Result<(), PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::SaveThreadMetadata(metadata, sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }
    pub async fn get_contact_by_id(&self, id: Uuid) -> Result<Option<Contact>, PresageError> {
        Ok(self
            .contacts
            .lock()
            .await
            .iter()
            .filter(|c| c.uuid == id)
            .map(almost_clone_contact)
            .next())
    }

    pub async fn get_group(
        &self,
        group_master_key: GroupMasterKeyBytes,
    ) -> Result<Option<Group>, PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::GetGroup(group_master_key, sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }

    pub async fn send_message(
        &self,
        recipient_addr: impl Into<ServiceAddress>,
        message: impl Into<ContentBody>,
        timestamp: u64,
    ) -> Result<(), PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::SendMessage(
                recipient_addr.into(),
                Box::new(message.into()),
                timestamp,
                sender,
            ))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }

    pub async fn send_message_to_group(
        &self,
        group_master_key: GroupMasterKeyBytes,
        message: DataMessage,
        timestamp: u64,
    ) -> Result<(), PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::SendMessageToGroup(
                group_master_key,
                Box::new(message),
                timestamp,
                sender,
            ))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }

    pub async fn get_attachment(
        &self,
        attachment_pointer: &AttachmentPointer,
    ) -> Result<Vec<u8>, PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::GetAttachment(attachment_pointer.clone(), sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }

    pub async fn upload_attachments(
        &self,
        attachments: Vec<(AttachmentSpec, Vec<u8>)>,
    ) -> Result<Vec<Result<AttachmentPointer, AttachmentUploadError>>, PresageError> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::UploadAttachments(attachments, sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }
    pub async fn messages(
        &self,
        thread: Thread,
        range: (Bound<u64>, Bound<u64>),
    ) -> Result<<presage_store_sled::SledStore as presage::Store>::MessagesIter, PresageError> {
        let (sender, receiver) = oneshot::channel();
        // TODO: self.current_chat = Some(thread.clone());
        self.command_sender
            .send(Command::GetMessages(thread, range, sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }
    // open_chat saves the current chat to avoid getting notifications
    pub async fn open_chat(&self, thread: Thread) -> Result<(), PresageError> {
        let mut current_chat = self.current_chat.lock().await;
        *current_chat = Some(thread.clone());
        Ok(())
    }
    pub async fn close_chat(&self) -> Result<(), PresageError> {
        let mut current_chat = self.current_chat.lock().await;
        *current_chat = None;
        Ok(())
    }
}

async fn setup_manager(
    config_store: SledStore,
    name: String,
    link_callback: futures::channel::oneshot::Sender<url::Url>,
) -> Result<presage::Manager<SledStore, presage::Registered>, PresageError> {
    log::info!("Loading the configuration store");
    // presage::Manager::load_registered(config_store.clone())
    if config_store.is_registered() {
        log::debug!("The config store is valid and reports registered, loading the manager");
        let manager = presage::Manager::load_registered(config_store.clone()).await?;
        log::info!("The configuration store is already valid, loading a registered account");
        drop(link_callback);
        Ok(manager)
    } else {
        log::debug!("The config store reports not registered, registering the manage");

        log::info!("The config store is not valid yet, not registered yet");
        match presage::Manager::link_secondary_device(
            config_store.clone(),
            presage::prelude::SignalServers::Production,
            name,
            link_callback,
        )
        .await
        {
            Ok(manager) => {
                log::info!("Successfully registered");
                //todo notify the user
                Ok(manager)
            }
            Err(e) => {
                log::error!("Failed to register: {}", e);
                Err(Error::NotYetRegisteredError)
            }
        }
    }
}

pub struct Notification {
    sender: String,
    message: String,
    group: Option<String>,
    thread: Thread,
}

async fn command_loop(
    manager: &mut Manager<SledStore, Registered>,
    mut receiver: mpsc::Receiver<Command>,
    content: mpsc::UnboundedSender<Content>,
    error: mpsc::Sender<ApplicationError>,
) {
    'outer: loop {
        let msgs = manager.receive_messages().await;
        match msgs {
            Ok(messages) => {
                futures::pin_mut!(messages);
                loop {
                    select! {
                        msg = messages.next().fuse() => {
                            if let Some(msg) = msg {
                                match msg.body.clone() {
                                    ContentBody::DataMessage(data) => {
                                        log::debug!("Received message data: {:?}", &data);

                                        let body = data.body.as_ref().unwrap_or(&String::from("")).to_string();
                                        let thread = Thread::try_from(&msg).unwrap();
                                        let title = manager.get_title_for_thread(&thread).await.unwrap_or("".to_string());
                                        let mut thread_metadata = manager.thread_metadata(&thread).await.unwrap().unwrap();
                                        if !body.is_empty() || !data.attachments.is_empty() {
                                            thread_metadata.title = Some(title.clone());
                                            thread_metadata.unread_messages_count += 1;
                                            let mut message = Some(body.clone());
                                            if body.is_empty() {
                                                message = Some("Attachment".to_string());
                                            }
                                            thread_metadata.last_message = Some(ThreadMetadataMessageContent{
                                                message,
                                                timestamp: msg.metadata.timestamp,
                                                sender: msg.metadata.sender.uuid,
                                            });
                                            let _ = manager.save_thread_metadata(thread_metadata.clone());

                                        }
                                        let sender = msg.metadata.sender.uuid;
                                        let is_group = match thread {
                                            Thread::Group(_)=> true,
                                            _ => false,
                                        };
                                        let mut notification = Notification{
                                            sender: title.clone(),
                                            message: body,
                                            group: None,
                                            thread,
                                        };
                                        if is_group {
                                            notification.group = Some(title);
                                            let contact_thread = Thread::Contact(sender);
                                            let contact_title = manager.get_title_for_thread(&contact_thread).await.unwrap_or("".to_string());
                                            notification.sender = contact_title;
                                        }
                                        // download attachments
                                        if !data.attachments.is_empty() {
                                            let attachments = data.attachments.clone();
                                            for attachment in attachments {
                                                let attachment_pointer = attachment.clone();
                                                match manager.get_attachment(&attachment_pointer).await{
                                                    Ok(attachment) => {
                                                        let identifier = match attachment_pointer.attachment_identifier.clone().unwrap() {
                                                            AttachmentIdentifier::CdnId(id) => id.to_string(),
                                                            AttachmentIdentifier::CdnKey(key) => key,
                                                            _ => {
                                                                log::debug!("Attachment: {:?} \n pointer {:?}", attachment.clone(), attachment_pointer.clone());
                                                                log::error!("The uploaded attachment has no identifier.");
                                                                "0".to_string()
                                                            }
                                                        };
                                                        handlers::save_attachment(&attachment, &identifier);
                                                    },
                                                    Err(e) => {
                                                        log::error!("Failed to download attachment: {}", e);
                                                    }
                                                }
                                            }
                                        }
                                        if data.reaction.is_some(){
                                            notification.message = data.reaction.unwrap().emoji.unwrap();
                                            //TODO: handle reactions
                                            continue;
                                        }

                                        if thread_metadata.muted {
                                            continue;
                                        }
                                        if notification.message==*"" {
                                            if !data.attachments.is_empty() {
                                                notification.message = "Attachment".to_string();
                                            }else{
                                                continue;
                                            }
                                         }
                                        notify_message(&notification).await;
                                    }
                                    ContentBody::SynchronizeMessage(sync_message) => {
                                      match sync_message.sent{
                                        Some(sm) => {
                                            match sm.message {
                                                Some(m) =>{
                                                    // download attachments
                                                    if !m.attachments.is_empty() {
                                                       let attachments = m.attachments.clone();
                                                       for attachment in attachments {
                                                           let attachment_pointer = attachment.clone();
                                                           match manager.get_attachment(&attachment_pointer).await{
                                                               Ok(attachment) => {
                                                                   let cdnid = match attachment_pointer.attachment_identifier.clone().unwrap() {
                                                                       AttachmentIdentifier::CdnId(id) => id,
                                                                       _ => {
                                                                           log::error!("The uploaded attachment has no identifier.");
                                                                           0
                                                                       }
                                                                   };
                                                                   handlers::save_attachment(&attachment, &cdnid.to_string());
                                                               },
                                                               Err(e) => {
                                                                   log::error!("Failed to download attachment: {}", e);
                                                               }
                                                           }
                                                       }
                                                    }
                                                    match m.body.clone() {

                                                        Some(_data) => {
                                                            let body = m.body.as_ref().unwrap_or(&String::from("")).to_string();
                                                            let thread = Thread::try_from(&msg).unwrap();
                                                            log::debug!("Received sync data message: {:?}", &thread);
                                                            let title = manager.get_title_for_thread(&thread).await.unwrap_or("".to_string());
                                                            let mut thread_metadata = manager.thread_metadata(&thread).await.unwrap().unwrap();
                                                            if !body.is_empty() {
                                                                thread_metadata.title = Some(title.clone());
                                                                thread_metadata.unread_messages_count += 1;
                                                                thread_metadata.last_message = Some(ThreadMetadataMessageContent{
                                                                    message: Some(body.clone()),
                                                                    timestamp: msg.metadata.timestamp,
                                                                    sender: msg.metadata.sender.uuid,
                                                                });
                                                                let _ = manager.save_thread_metadata(thread_metadata.clone());

                                                            }
                                                        },
                                                        None => {}
                                                    }
                                                },
                                                None => {}
                                            }
                                        },
                                        None => {}
                                      }
                                    }
                                    _ => {}
                                }
                                if content.send(msg).is_err() {
                                    log::info!("Failed to send message to `Manager`, exiting");
                                    break 'outer;
                                }
                            } else {
                                log::info!("Message stream finished. Restarting command loop.");
                                break;
                            }
                        },
                        cmd = receiver.recv().fuse() => {
                            if let Some(cmd) = cmd {
                                handle_command(manager, cmd).await;
                            }
                        },
                        // _ = crate::utils::await_suspend_wakeup_online().fuse() => {
                        //     log::info!("Waking up from suspend. Restarting command loop.");
                        //     break;
                        // },
                        complete => {
                            log::info!("Command loop complete. Restarting command loop.");
                            break
                        },
                    }
                }
            }
            Err(e) => {
                log::info!("Got error receiving: {}, {:?}", e, e);
                let e = e.into();
                // Don't send no-internet errors, Flare is able to handle them automatically.
                // TODO: Think about maybe handling if the application is not in the background?
                if !matches!(e, ApplicationError::NoInternet) {
                    error.send(e).await.expect("Callback sending failed");
                }
                tokio::time::sleep(std::time::Duration::from_secs(15)).await;
            }
        }
        log::info!("Websocket closed, trying again");
    }
    log::info!("Exiting `ManagerThread::command_loop`");
}

#[cfg(not(feature = "ut"))]
async fn notify_message(msg: &Notification) {
    use notify_rust::Notification;
    match &msg.group {
        Some(group) => {
            let body = format!("{}: {}", msg.sender, msg.message);
            Notification::new()
                .summary(group)
                .body(&body)
                .icon("signal")
                .timeout(5000)
                .show()
                .expect("Failed to send notification");
        }
        None => {
            Notification::new()
                .summary(&msg.sender)
                .body(&msg.message)
                .icon("signal")
                .timeout(5000)
                .show()
                .expect("Failed to send notification");
        }
    }
}

#[cfg(feature = "ut")]
const DBUS_NAME: &str = "com.lomiri.Postal";

#[cfg(feature = "ut")]
const DBUS_INTERFACE: &str = "com.lomiri.Postal";

#[cfg(feature = "ut")]
const DBUS_PATH_PART: &str = "/com/lomiri/Postal/";

#[cfg(feature = "ut")]
const DBUS_POST_METHOD: &str = "Post";

#[cfg(feature = "ut")]
const DBUS_CLEAR_METHOD: &str = "ClearPersistent";

#[cfg(feature = "ut")]
const DBUS_LIST_METHOD: &str = "ListPersistent";

#[cfg(feature = "ut")]
const APP_ID: &str = "textsecure.nanuc";

#[cfg(feature = "ut")]
const HOOK_NAME: &str = "textsecure";

#[cfg(feature = "ut")]
fn postal<R: ReadAll, A: AppendAll>(method: &str, args: A) -> Result<R, dbus::Error> {
    let conn = Connection::new_session()?;
    let proxy = conn.with_proxy(
        DBUS_NAME,
        format!(
            "{}{}",
            DBUS_PATH_PART,
            APP_ID.replace(".", "_2e").replace("-", "_2f")
        ),
        Duration::from_millis(5000),
    );
    proxy.method_call(DBUS_INTERFACE, method, args)
}

#[cfg(feature = "ut")]
fn postal_post(data: Value) -> Result<(), dbus::Error> {
    postal(
        DBUS_POST_METHOD,
        (format!("{}_{}", APP_ID, HOOK_NAME), data.to_string()),
    )?;
    Ok(())
}
#[cfg(feature = "ut")]
fn postal_clear_persistent(tag: &str) -> Result<(), dbus::Error> {
    postal(
        DBUS_CLEAR_METHOD,
        (format!("{}_{}", APP_ID, HOOK_NAME), tag),
    )?;
    Ok(())
}

#[cfg(feature = "ut")]
async fn notify_message(msg: &Notification) {
    let _yconn = match Connection::new_session() {
        Ok(c) => c,
        Err(e) => {
            log::error!("Failed to connect to dbus: {}", e);
            return;
        }
    };
    let tag = msg.thread.to_string();
    postal_clear_persistent(tag.as_str()).expect("Failed to clear persistent notification");
    let mut data = json!({
      "message": msg.sender,
      "notification": {
        "card": {
          "body": msg.message,
          "persist": true,
          "popup": true,
          "summary":  msg.sender
        },
        "sound": "buzz.mp3",
        "tag": tag,
        "vibrate": {
          "duration": 200,
          "pattern": [ 200, 100 ],
          "repeat": 2
        }
      }
    });

    match &msg.group {
        Some(group) => {
            let body = format!("{}: {}", msg.sender, msg.message);
            data["notification"]["card"]["body"] = json!(body);
            data["notification"]["card"]["summary"] = json!(group);
            data["notification"]["message"] = json!(group);
        }
        None => {}
    }
    postal_post(data).expect("Failed to send notification");
}

async fn handle_command(manager: &mut Manager<SledStore, Registered>, command: Command) {
    match command {
        Command::RequestContactsSync(callback) => callback
            .send(manager.request_contacts_sync().await)
            .expect("Callback sending failed"),
        Command::GetContacts(callback) => callback
            .send(
                manager
                    .contacts()
                    .map(|c| c.filter_map(|o| o.ok()).collect()),
            )
            .expect("Callback sending failed"),
        Command::GetConversations(callback) => callback
            .send(match manager.thread_metadatas().await {
                Ok(m) => Ok(m.filter_map(|r| r.ok()).collect()),
                Err(e) => {
                    log::error!("Failed to get thread metadatas: {}", e);
                    return ;
                }
            })
            .expect("Callback sending failed"),
        Command::GetGroup(master_key, callback) => callback
            .send(manager.group(&master_key))
            .map_err(|_| ())
            .expect("Callback sending failed"),
        Command::SendMessage(recipient_address, message, timestamp, callback) => callback
            .send(
                manager
                    .send_message(recipient_address, *message, timestamp)
                    .await,
            )
            .expect("Callback sending failed"),
        Command::SendMessageToGroup(group, message, timestamp, callback) => callback
            .send(
                manager
                    .send_message_to_group(&group, *message, timestamp)
                    .await,
            )
            .expect("Callback sending failed"),
        Command::GetAttachment(attachment, callback) => callback
            .send(manager.get_attachment(&attachment).await)
            .expect("Callback sending failed"),
        Command::UploadAttachments(attachments, callback) => callback
            .send(manager.upload_attachments(attachments).await)
            .expect("Callback sending failed"),

        Command::GetMessages(thread, range, callback) => {
            let _ = callback.send(manager.messages(&thread, range));
        }
        Command::ThreadMetadata(thread, callback) => callback
            .send(manager.thread_metadata(&thread).await)
            .expect("Callback sending failed"),
        Command::SaveThreadMetadata(metadata, callback) => callback
            .send(manager.save_thread_metadata(metadata))
            .expect("Callback sending failed"),
        Command::RequestContactsUpdateFromProfile(callback) => callback
            .send(manager.request_contacts_update_from_profile().await)
            .expect("Callback sending failed"),
        Command::RequestContactUpdateFromProfile(uuid, callback) => callback
            .send(manager.request_contact_update_from_profile(uuid).await)
            .expect("Callback sending failed"),
    };
}

// TODO: Clone attachment
fn almost_clone_contact(contact: &Contact) -> Contact {
    Contact {
        uuid: contact.uuid,
        name: contact.name.clone(),
        color: contact.color.clone(),
        verified: contact.verified.clone(),
        profile_key: contact.profile_key.clone(),
        phone_number: contact.phone_number.clone(),
        blocked: contact.blocked,
        expire_timer: contact.expire_timer,
        inbox_position: contact.inbox_position,
        archived: contact.archived,
        avatar: None,
    }
}