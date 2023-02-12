use std::sync::Arc;
use tokio::sync::Mutex;
use futures::{select, FutureExt, StreamExt};
use presage::libsignal_service::{groups_v2::Group, sender::AttachmentUploadError};
use presage::{Thread, GroupMasterKeyBytes};
use presage::{
    manager::Session,
    prelude::{content::*, AttachmentSpec, Contact, ContentBody, DataMessage, ServiceAddress, *},
    Error, Manager, MessageStore, Registered, Store,
};
use serde::{Deserialize, Serialize};
use tokio::sync::{mpsc, oneshot};

use crate::error::ApplicationError;

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
    RequestContactsSync(oneshot::Sender<Result<(), Error>>),
    Uuid(oneshot::Sender<Uuid>),
    GetContacts(oneshot::Sender<Result<Vec<Contact>, Error>>),
    GetGroup(GroupMasterKeyBytes, oneshot::Sender<Result<Option<Group>, Error>>),
    SendMessage(
        ServiceAddress,
        Box<ContentBody>,
        u64,
        oneshot::Sender<Result<(), Error>>,
    ),
    SendMessageToGroup(
        GroupMasterKeyBytes,
        Box<DataMessage>,
        u64,
        oneshot::Sender<Result<(), Error>>,
    ),
    GetConversations(oneshot::Sender<Result<Vec<Session>, Error>>),
    GetAttachment(AttachmentPointer, oneshot::Sender<Result<Vec<u8>, Error>>),
    UploadAttachments(
        Vec<(AttachmentSpec, Vec<u8>)>,
        oneshot::Sender<Result<Vec<Result<AttachmentPointer, AttachmentUploadError>>, Error>>,
    ),
    GetMessages(
        Thread,
        Option<u64>,
        oneshot::Sender<Result<Vec<Content>, Error>>,
    ),
    RequestContactsUpdateFromProfile(oneshot::Sender<Result<(), Error>>),
}

impl std::fmt::Debug for Command {
    fn fmt(&self, _f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        Ok(())
    }
}

pub struct ManagerThread {
    command_sender: mpsc::Sender<Command>,
    uuid: Uuid,
    contacts: Arc<Mutex<Vec<Contact>>>,
    sessions: Arc<Mutex<Vec<AxolotlSession>>>,
    current_chat: Arc<Mutex<Option<Thread>>>,
}
#[derive(Serialize, Deserialize, Debug, Clone)]

pub struct AxolotlSession {
    pub id: String,
    pub last_message: Option<String>,
    pub unread_messages_count: usize,
    pub is_group: bool,
    pub title: Option<String>,
    pub last_message_timestamp: u64,
}
impl TryFrom<Session> for AxolotlSession {
    type Error = Error;

    fn try_from(session: Session) -> Result<Self, Error> {
        let id: String = session.thread.to_string();
        let mut timestamp: u64 = 0;
        let message: Option<String> = match session.last_message {
            Some(message) => Some(match message.body {
                ContentBody::DataMessage(msg) => {
                    timestamp = msg.timestamp.unwrap_or_default();
                    msg.body.unwrap_or("".to_string())
                }
                ContentBody::SynchronizeMessage(msg) => {
                    let sent_message = msg.sent.unwrap_or_default().message.unwrap_or_default();
                    timestamp = sent_message.timestamp.unwrap_or_default();
                    if sent_message.body.is_some() {
                        sent_message.body.unwrap_or("".to_string())
                    } else if sent_message.attachments.len()>0 {
                        "Attachment".to_string()
                    } else if sent_message.reaction.is_some() {
                        sent_message.reaction.unwrap_or_default().emoji.unwrap_or_default()
                    } else {
                       "Unknown sync message".to_string()
                    }
                }
                _ => "Unknown message type".to_string(),
            }),
            None => None,
        };
        let is_group: bool = match session.thread {
            Thread::Group(_group) => true,
            _ => false,
        };

        Ok(Self {
            id: id,
            last_message: message,
            unread_messages_count: session.unread_messages_count,
            is_group: is_group,
            title: session.title,
            last_message_timestamp: timestamp,
        })
    }
}

impl Clone for ManagerThread {
    fn clone(&self) -> Self {
        Self {
            command_sender: self.command_sender.clone(),
            uuid: self.uuid,
            contacts: self.contacts.clone(),
            sessions: self.sessions.clone(),
            current_chat: self.current_chat.clone(),
        }
    }
}

impl ManagerThread {
    pub async fn new<C>(
        config_store: C,
        device_name: String,
        link_callback:futures::channel::oneshot::Sender<url::Url>,
        error_callback: futures::channel::oneshot::Sender<Error>,
        content: mpsc::UnboundedSender<Content>,
        current_chat: Arc<Mutex<Option<Thread>>>,
        error: mpsc::Sender<ApplicationError>,
    ) -> Option<Self>
    where
        C: presage::Store + std::marker::Send + std::marker::Sync + 'static + presage::MessageStore,
    {
        let (sender, receiver) = mpsc::channel(MESSAGE_BOUND);
        std::thread::spawn(move || {
            let error_clone = error.clone();
            let panic = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                tokio::runtime::Runtime::new()
                    .expect("Failed to setup runtime")
                    .block_on(async move {
                        let setup = setup_manager(config_store.clone(), device_name, link_callback).await;
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

        let (sender_uuid, receiver_uuid) = oneshot::channel();
        if sender.send(Command::Uuid(sender_uuid)).await.is_err() {
            return None;
        }
        let uuid = receiver_uuid.await;

        let (sender_contacts, receiver_contacts) = oneshot::channel();
        if sender
            .send(Command::GetContacts(sender_contacts))
            .await
            .is_err()
        {
            return None;
        }
        let contacts = receiver_contacts.await;

        if uuid.is_err() || contacts.is_err() {
            return None;
        }

        if let Err(_e) = &contacts.as_ref().unwrap() {
            // TODO: Error handling
            log::info!("Could not load contacts");
        }
        Some(Self {
            command_sender: sender,
            uuid: uuid.unwrap(),
            contacts: Arc::new(Mutex::new(contacts.unwrap().unwrap_or_default())),
            sessions: Arc::new(Mutex::new(Vec::new())),
            current_chat: current_chat,
        })
    }
}

impl ManagerThread {
    pub async fn sync_contacts(&mut self) -> Result<(), Error> {
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
    pub async fn request_contacts_sync(&self) -> Result<(), Error> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::RequestContactsSync(sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }

    pub fn uuid(&self) -> Uuid {
        self.uuid
    }
    pub async fn current_chat(&self) -> Option<Thread> {
        self.current_chat.lock().await.clone()
    }

    pub async fn get_contacts(&self) -> Result<impl Iterator<Item = Contact> + '_, Error> {
        let c = self.contacts.lock().await;
        // Very weird way to counteract "returning borrowed c".
        Ok(c.iter()
            .map(almost_clone_contact)
            .collect::<Vec<_>>()
            .into_iter())
    }
    pub async fn update_cotacts_from_profile(&self) -> Result<(), Error> {
        log::debug!("Updating contacts from profile");
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::RequestContactsUpdateFromProfile(sender))
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
                let axolotl_sessions = sessions
                    .into_iter()
                    .map(|s| AxolotlSession::try_from(s).unwrap())
                    .collect::<Vec<_>>();
                Ok(axolotl_sessions.into_iter())
            }
            Ok(Err(e)) => {
                log::error!("Loading coversations failed: {}", e);
                Err(ApplicationError::ManagerThreadPanic)
            }
            Err(_) => Err(ApplicationError::ManagerThreadPanic),
        }
    }
    pub async fn get_contact_by_id(&self, id: Uuid) -> Result<Option<Contact>, Error> {
        Ok(self
            .contacts
            .lock()
            .await
            .iter()
            .filter(|c| c.address.uuid == id)
            .map(almost_clone_contact)
            .next())
    }

    pub async fn get_group(&self, group_master_key: GroupMasterKeyBytes) -> Result<Option<Group>, Error> {
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
    ) -> Result<(), Error> {
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
    ) -> Result<(), Error> {
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
    ) -> Result<Vec<u8>, Error> {
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
    ) -> Result<Vec<Result<AttachmentPointer, AttachmentUploadError>>, Error> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::UploadAttachments(attachments, sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }
    pub async fn get_messages(
        &self,
        thread: Thread,
        count: Option<u64>,
    ) -> Result<Vec<Content>, Error> {
        let (sender, receiver) = oneshot::channel();
        // self.current_chat = Some(thread.clone());
        self.command_sender
            .send(Command::GetMessages(thread, count, sender))
            .await
            .expect("Command sending failed");
        receiver.await.expect("Callback receiving failed")
    }
    // open_chat saves the current chat to avoid getting notifications
    pub async fn open_chat(&self, thread: Thread) -> Result<(), Error> {
        let mut current_chat = self.current_chat.lock().await;
        *current_chat = Some(thread.clone());
        Ok(())
    }
    pub async fn close_chat(&self) -> Result<(), Error> {
        let mut current_chat = self.current_chat.lock().await;
        *current_chat = None;
        Ok(())
    }

}

async fn setup_manager<C>(
    config_store: C,
    name: String,
    link_callback: futures::channel::oneshot::Sender<url::Url>,
) -> Result<presage::Manager<C, presage::Registered>, Error>
where
    C: Store + 'static,
{
    log::info!("Loading the configuration store");
    // presage::Manager::load_registered(config_store.clone())

    if let Ok(manager) = presage::Manager::load_registered(config_store.clone()) {
        log::info!("The configuration store is already valid, loading a registered account");
        drop(link_callback);
        Ok(manager)
    } else {
        log::info!("The config store is not valid yet, not registered yet");
        match presage::Manager::link_secondary_device(
            config_store.clone(),
            presage::prelude::SignalServers::Production,
            name,
            link_callback,
        )
        .await{
            Ok(manager) => {
                log::info!("Successfully registered");
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

async fn command_loop<C: Store + 'static + MessageStore>(
    manager: &mut Manager<C, Registered>,
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
                                        log::info!("Received message: {:?}", &data);

                                        let body = data.body.as_ref().unwrap_or(&String::from("")).to_string();
                                        let thread = Thread::try_from(&msg).unwrap();
                                        let title = manager.get_title_for_thread(&thread).await.unwrap_or("".to_string());
                                        let sender = msg.metadata.sender.uuid.clone();
                                        let is_group = match thread {
                                            Thread::Group(_)=> true,
                                            _ => false,
                                        };
                                        let mut notification = Notification{
                                            sender: title.clone(),
                                            message: body,
                                            group: None,
                                            thread: thread,
                                        };
                                        if is_group {
                                            notification.group = Some(title);
                                            let contact_thread = Thread::Contact(sender);
                                            let contact_title = manager.get_title_for_thread(&contact_thread).await.unwrap_or("".to_string());
                                            notification.sender = contact_title;
                                        }
                                        if data.reaction.is_some(){
                                            notification.message = data.reaction.unwrap().emoji.unwrap();
                                        }
                                        if notification.message=="".to_string() {
                                           continue;
                                        }
                                        notify_message(&notification).await;

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
                .summary(&group)
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

async fn handle_command<C: Store + 'static>(
    manager: &mut Manager<C, Registered>,
    command: Command,
) {
    match command {
        Command::RequestContactsSync(callback) => callback
            .send(manager.request_contacts_sync().await)
            .expect("Callback sending failed"),
        Command::Uuid(callback) => callback
            .send(manager.uuid())
            .expect("Callback sending failed"),
        Command::GetContacts(callback) => callback
            .send(
                manager
                    .contacts()
                    .map(|c| c.filter_map(|o| o.ok()).collect()),
            )
            .expect("Callback sending failed"),
        Command::GetConversations(callback) => callback
            .send(manager.load_conversations().await)
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
        Command::GetMessages(thread, count, callback) => callback
            .send(manager.get_messages(&thread, count))
            .expect("Callback sending failed"),
        Command::RequestContactsUpdateFromProfile(callback) => callback
            .send(manager.request_contacts_update_from_profile().await)
            .expect("Callback sending failed"),
    };
}

// TODO: Clone attachment
fn almost_clone_contact(contact: &Contact) -> Contact {
    Contact {
        address: contact.address.clone(),
        name: contact.name.clone(),
        color: contact.color.clone(),
        verified: contact.verified.clone(),
        profile_key: contact.profile_key.clone(),
        blocked: contact.blocked,
        expire_timer: contact.expire_timer,
        inbox_position: contact.inbox_position,
        archived: contact.archived,
        avatar: None,
    }
}
