use std::sync::{Arc, Mutex};

use futures::{select, FutureExt, StreamExt};
use presage::libsignal_service::{groups_v2::Group, sender::AttachmentUploadError};
use presage::{
    prelude::{content::*, AttachmentSpec, Contact, ContentBody, DataMessage, ServiceAddress, *},
    Error, Manager, MessageStore, Registered, Store,
};
use tokio::sync::{mpsc, oneshot};

use crate::error::ApplicationError;

const MESSAGE_BOUND: usize = 10;

enum Command {
    RequestContactsSync(oneshot::Sender<Result<(), Error>>),
    Uuid(oneshot::Sender<Uuid>),
    GetContacts(oneshot::Sender<Result<Vec<Contact>, Error>>),
    GetGroupV2(GroupMasterKey, oneshot::Sender<Result<Group, Error>>),
    SendMessage(
        ServiceAddress,
        Box<ContentBody>,
        u64,
        oneshot::Sender<Result<(), Error>>,
    ),
    SendMessageToGroup(
        Vec<ServiceAddress>,
        Box<DataMessage>,
        u64,
        oneshot::Sender<Result<(), Error>>,
    ),
    GetAttachment(AttachmentPointer, oneshot::Sender<Result<Vec<u8>, Error>>),
    UploadAttachments(
        Vec<(AttachmentSpec, Vec<u8>)>,
        oneshot::Sender<Result<Vec<Result<AttachmentPointer, AttachmentUploadError>>, Error>>,
    ),
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
}

impl Clone for ManagerThread {
    fn clone(&self) -> Self {
        Self {
            command_sender: self.command_sender.clone(),
            uuid: self.uuid,
            contacts: self.contacts.clone(),
        }
    }
}

impl ManagerThread {
    pub async fn new<C>(
        config_store: C,
        device_name: String,
        link_callback: futures::channel::oneshot::Sender<url::Url>,
        error_callback: futures::channel::oneshot::Sender<Error>,
        content: mpsc::UnboundedSender<Content>,
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
                        let setup = setup_manager(config_store, device_name, link_callback).await;
                        if let Ok(mut manager) = setup {
                            println!("Starting command loop");
                            drop(error_callback);
                            command_loop(&mut manager, receiver, content, error).await;
                        } else {
                            let e = setup.err().unwrap();
                            println!("Got error: {}", e);
                            error_callback.send(e).expect("Failed to send error")
                        }
                    });
            }));
            if let Err(_e) = panic {
                println!("Manager-thread paniced");
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
            println!("Could not load contacts");
        }
        Some(Self {
            command_sender: sender,
            uuid: uuid.unwrap(),
            contacts: Arc::new(Mutex::new(contacts.unwrap().unwrap_or_default())),
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
        let mut c = self.contacts.lock().expect("Poisoned mutex");
        *c = contacts;
        println!("Synced contacts. Got {} contacts.", c.len());
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

    pub fn get_contacts(&self) -> Result<impl Iterator<Item = Contact> + '_, Error> {
        let c = self.contacts.lock().expect("Poisoned mutex");
        // Very weird way to counteract "returning borrowed c".
        Ok(c.iter()
            .map(almost_clone_contact)
            .collect::<Vec<_>>()
            .into_iter())
    }

    pub fn get_contact_by_id(&self, id: Uuid) -> Result<Option<Contact>, Error> {
        Ok(self
            .contacts
            .lock()
            .expect("Poisoned mutex")
            .iter()
            .filter(|c| c.address.uuid == Some(id))
            .map(almost_clone_contact)
            .next())
    }

    pub async fn get_group_v2(&self, group_master_key: GroupMasterKey) -> Result<Group, Error> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::GetGroupV2(group_master_key, sender))
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
        recipients: impl IntoIterator<Item = ServiceAddress>,
        message: DataMessage,
        timestamp: u64,
    ) -> Result<(), Error> {
        let (sender, receiver) = oneshot::channel();
        self.command_sender
            .send(Command::SendMessageToGroup(
                recipients.into_iter().collect(),
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
}

async fn setup_manager<C>(
    config_store: C,
    name: String,
    link_callback: futures::channel::oneshot::Sender<url::Url>,
) -> Result<presage::Manager<C, presage::Registered>, Error>
where
    C: Store + 'static,
{
    println!("Loading the configuration store");
    // presage::Manager::load_registered(config_store.clone())

    if let Ok(manager) = presage::Manager::load_registered(config_store.clone()) {
        println!("The configuration store is already valid, loading a registered account");
        drop(link_callback);
        Ok(manager)
    } else {
        println!("The config store is not valid yet, linking with a secondary device");
        presage::Manager::link_secondary_device(
            config_store.clone(),
            presage::prelude::SignalServers::Production,
            name,
            link_callback,
        )
        .await
    }
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
                                if content.send(msg).is_err() {
                                    println!("Failed to send message to `Manager`, exiting");
                                    break 'outer;
                                }
                            } else {
                                println!("Message stream finished. Restarting command loop.");
                                break;
                            }
                        },
                        cmd = receiver.recv().fuse() => {
                            if let Some(cmd) = cmd {
                                handle_command(manager, cmd).await;
                            }
                        },
                        // _ = crate::utils::await_suspend_wakeup_online().fuse() => {
                        //     println!("Waking up from suspend. Restarting command loop.");
                        //     break;
                        // },
                        complete => {
                            println!("Command loop complete. Restarting command loop.");
                            break
                        },
                    }
                }
            }
            Err(e) => {
                println!("Got error receiving: {}, {:?}", e, e);
                let e = e.into();
                // Don't send no-internet errors, Flare is able to handle them automatically.
                // TODO: Think about maybe handling if the application is not in the background?
                if !matches!(e, ApplicationError::NoInternet) {
                    error.send(e).await.expect("Callback sending failed");
                }
                tokio::time::sleep(std::time::Duration::from_secs(15)).await;
            }
        }
        println!("Websocket closed, trying again");
    }
    println!("Exiting `ManagerThread::command_loop`");
}

async fn handle_command<C: Store + 'static>(
    manager: &mut Manager<C, Registered>,
    command: Command,
) {
    println!("Got command: {:?}", command);
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
                    .get_contacts()
                    .map(|c| c.filter_map(|o| o.ok()).collect()),
            )
            .expect("Callback sending failed"),
        Command::GetGroupV2(master_key, callback) => callback
            .send(manager.get_group_v2(master_key).await)
            .map_err(|_| ())
            .expect("Callback sending failed"),
        Command::SendMessage(recipient_address, message, timestamp, callback) => callback
            .send(
                manager
                    .send_message(recipient_address, *message, timestamp)
                    .await,
            )
            .expect("Callback sending failed"),
        Command::SendMessageToGroup(recipients, message, timestamp, callback) => callback
            .send(
                manager
                    .send_message_to_group(recipients, *message, timestamp)
                    .await,
            )
            .expect("Callback sending failed"),
        Command::GetAttachment(attachment, callback) => callback
            .send(manager.get_attachment(&attachment).await)
            .expect("Callback sending failed"),
        Command::UploadAttachments(attachments, callback) => callback
            .send(manager.upload_attachments(attachments).await)
            .expect("Callback sending failed"),
    }
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
