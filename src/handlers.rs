use crate::error::ApplicationError;
use crate::manager_thread::ManagerThread;
use crate::messages::send_message;
use crate::requests::{
    AxolotlConfig, AxolotlMessage, AxolotlRequest, AxolotlResponse,
    ChangeNotificationsForThreadRequest, GetMessagesRequest, ProfileRequest, SendMessageRequest,
    SendMessageResponse, UploadAttachmentRequest,
};
// #[cfg(feature = "ut")]
use crate::requests::UploadAttachmentUtRequest;
use data_url::DataUrl;
extern crate dirs;
use futures::channel::oneshot::Receiver;
use futures::executor::block_on;
use futures::stream::{SplitSink, SplitStream};
use futures::{SinkExt, StreamExt};
use libsignal_service::prelude::{phonenumber, Uuid};
use presage::libsignal_service::prelude::AttachmentIdentifier;
use presage::prelude::proto::AttachmentPointer;
use presage::prelude::AttachmentSpec;
use presage::prelude::{Content, PhoneNumber};
use presage::prelude::{ContentBody, DataMessage, GroupContextV2};
use presage_store_sled::SledStoreError;
use serde::{Serialize, Serializer};
use serde_json::error::Error as SerdeError;
use std::cell::Cell;
use std::fs::{self, File};
use std::io::{Read, Write};
use std::ops::Bound::Unbounded;
use std::process::exit;
use std::time::{self, UNIX_EPOCH};
use std::{sync::Arc, thread};
use tokio::runtime::Runtime;
use tokio::sync::Mutex;
use url::Url;
use warp::filters::ws::Message;
use warp::http::response;
use warp::ws::WebSocket;

const MESSAGE_BOUND: usize = 10;

use futures::channel::oneshot;

use presage::{Confirmation, Manager, RegistrationOptions, Store};
use presage::{Thread, ThreadMetadata};
use presage_store_sled::MigrationConflictStrategy;
use presage_store_sled::SledStore;

use tokio::sync::mpsc::{self, UnboundedReceiver};

pub struct Handler {
    pub provisioning_link_rx: Option<Arc<Mutex<Receiver<Url>>>>,
    pub provisioning_link: Option<Url>,
    pub error_rx: Option<Arc<Mutex<Receiver<presage::Error<SledStoreError>>>>>,
    pub receive_error: mpsc::Receiver<ApplicationError>,
    pub sender: Option<Arc<Mutex<SplitSink<WebSocket, Message>>>>,
    pub receiver: Option<Arc<Mutex<SplitStream<WebSocket>>>>,
    pub manager_thread: Arc<Mutex<Option<ManagerThread>>>,
    pub receive_content: Arc<Mutex<Option<UnboundedReceiver<Content>>>>,
    pub is_registered: Option<bool>,
    pub captcha: Option<String>,
    pub phone_number: Option<PhoneNumber>,
    pub registration_manager: Option<Arc<Mutex<Manager<SledStore, Confirmation>>>>,
}
impl Handler {
    pub async fn new() -> Result<Self, ApplicationError> {
        log::info!("Setting up the handler");
        let (provisioning_link_tx, provisioning_link_rx) = oneshot::channel();
        let (error_tx, error_rx) = oneshot::channel();

        let (send_content, receive_content) = mpsc::unbounded_channel();
        let (send_error, receive_error) = mpsc::channel(MESSAGE_BOUND);
        let current_chat: Option<Thread> = None;
        let current_chat_mutex = Arc::new(Mutex::new(current_chat));
        let config_store = Handler::get_config_store().await?;
        let manager_thread: Arc<Mutex<Option<ManagerThread>>> = Arc::new(Mutex::new(None));
        let thread = manager_thread.clone();
        let mut is_registered = Some(false);
        log::info!("Setting up the manager2");
        if config_store.is_registered() {
            log::info!("Registered, starting the manager");
            is_registered = Some(true);

            tokio::spawn(async move {
                let manager = ManagerThread::new(
                    config_store.clone(),
                    "axolotl".to_string(),
                    provisioning_link_tx,
                    error_tx,
                    send_content,
                    current_chat_mutex,
                    send_error,
                )
                .await;
                log::info!("Manager thread started, ready to receive messages from the client");
                let mut m = thread.lock().await;
                *m = manager;
                // free the lock
                drop(m);
            });
        } else {
            log::info!("Not yet registered.");
        }

        Ok(Self {
            provisioning_link_rx: Some(Arc::new(Mutex::new(provisioning_link_rx))),
            provisioning_link: None,
            error_rx: Some(Arc::new(Mutex::new(error_rx))),
            receive_error,
            sender: None,
            receiver: None,
            manager_thread: manager_thread,
            receive_content: Arc::new(Mutex::new(Some(receive_content))),
            is_registered: is_registered,
            captcha: None,
            phone_number: None,
            registration_manager: None,
        })
    }
    pub async fn get_config_store() -> Result<SledStore, ApplicationError> {
        let config_path = dirs::config_dir()
            .unwrap()
            .into_os_string()
            .into_string()
            .unwrap();
        // todo: check if a tmp folder exists, if so, copy the content to the new folder and delete the tmp folder

        let db_path = format!("{config_path}/textsecure.nanuc/sled");
        let config_store = match SledStore::open_with_passphrase(
            db_path,
            None::<&str>,
            MigrationConflictStrategy::BackupAndDrop,
        ) {
            Ok(store) => store,
            Err(e) => {
                log::info!(
                    "Failed to open the database: {}, retry with tmp database",
                    e
                );
                let db_path = format!("{config_path}/textsecure.nanuc/tmp");
                match SledStore::open_with_passphrase(
                    db_path,
                    None::<&str>,
                    MigrationConflictStrategy::BackupAndDrop,
                ) {
                    Ok(store) => store,
                    Err(e) => {
                        log::info!("Failed to open the database: {}", e);
                        exit(0);
                    }
                }
            }
        };
        Ok(config_store)
    }
    /// Handles a client connection
    pub async fn handle_ws_client(&mut self, websocket: warp::ws::WebSocket) {
        // start manager only the first time, else replace the sender and receiver
        log::debug!("Starting the manager and handling the client.");
        match self.start_manager(websocket).await {
            Ok(_) => log::info!("Manager started"),
            Err(e) => {
                log::error!("Error starting the manager: {}", e);
            }
        }
        log::info!("Manager started");
    }

    pub async fn start_manager(
        &mut self,
        websocket: warp::ws::WebSocket,
    ) -> Result<(), ApplicationError> {
        let mut count = 0u32;

        let (sender, receiver) = websocket.split();
        let shared_sender = Arc::new(Mutex::new(sender));
        self.sender = Some(shared_sender.clone());
        self.receiver = Some(Arc::new(Mutex::new(receiver)));
        loop {
            log::debug!("Starting the manager loop: {:?}", count);
            if count == 5 || self.sender.is_none() {
                log::error!("Too many errors, exiting or sender is none {:?}", count);
                // Exit this loop
                break;
            }
            count += 1;

            if self.is_registered.is_none() {
                match Handler::check_registration().await {
                    Ok(_) => {
                        self.is_registered = Some(true);
                    }
                    Err(e) => {
                        log::debug!("Error checking registration: {}", e);
                        self.is_registered = Some(false);
                    }
                }
            }
            log::debug!("Is registered: {:?}", self.is_registered);

            if self.is_registered.is_some() && !self.is_registered.unwrap() {
                log::info!("Starting registration process");

                match self.start_registration().await {
                    Err(_) => {
                        self.sender = None;
                    }
                    _ => (),
                }

                // check if client wants to register as primary or secondary device

                let receiver = self.receiver.clone();
                if let Some(r) = receiver {
                    let mut r = r.lock().await;
                    let mut is_closed = false;
                    while let Some(message) = r.next().await {
                        match message {
                            Ok(message) => {
                                if message.is_close() {
                                    log::info!("Got close message, exiting");
                                    is_closed = true;
                                    break;
                                } else if message.is_text() {
                                    let text = message.to_str().unwrap();
                                    if let Ok::<AxolotlRequest, SerdeError>(axolotl_request) =
                                        serde_json::from_str(text)
                                    {
                                        // Axolotl request
                                        let request_type: &str = axolotl_request.request.as_str();
                                        log::info!(
                                            "Axolotl registration request: {}",
                                            request_type
                                        );
                                        if request_type == "primaryDevice" {
                                            self.get_phone_number().await;
                                        } else if request_type == "registerSecondaryDevice" {
                                            loop {
                                                log::debug!("Registering secondary device");
                                                self.create_provisioning_link().await?;
                                                if self.is_registered.is_some()
                                                    && self.is_registered.unwrap()
                                                {
                                                    log::debug!("Device is already registered");
                                                    break;
                                                }
                                                log::debug!(
                                                    "Provisioning link created successfully"
                                                );
                                                self.handle_provisoning().await;
                                                log::debug!(
                                                    "Provisioning link handled successfully"
                                                );
                                                self.send_provisioning_link().await;
                                                log::debug!(
                                                    "Provisioning link sent successfully to client"
                                                );
                                                let error_reciever =
                                                    self.error_rx.as_mut().unwrap();
                                                let mut error_reciever =
                                                    error_reciever.lock().await;
                                                while let Ok(e) = error_reciever.try_recv() {
                                                    match e {
                                                        Some(u) => {
                                                            log::error!("Error registering secondary device: {}", u);
                                                            Some(u)
                                                        }
                                                        None => {
                                                            thread::sleep(
                                                                time::Duration::from_secs(1),
                                                            );
                                                            continue;
                                                        }
                                                    };
                                                }
                                                if error_reciever.try_recv().is_err() {
                                                    log::debug!("Break out of loop, because error channel is closed");
                                                    match Handler::check_registration().await {
                                                        Ok(_) => {
                                                            self.is_registered = Some(true);
                                                            break;
                                                        }
                                                        Err(e) => {
                                                            log::debug!(
                                                                "Error checking registration: {}",
                                                                e
                                                            );
                                                            self.is_registered = Some(false);
                                                        }
                                                    }
                                                }
                                            }
                                            self.send_registration_confirmation().await;
                                            log::debug!("Registration confirmation sent and done, now sleeping");
                                            if let Some(error_opt) = self.receive_error.recv().await
                                            {
                                                log::error!(
                                                    "Got error after linking device2: {}",
                                                    error_opt
                                                );
                                                continue;
                                            }
                                        }
                                        if request_type == "sendCaptchaToken" {
                                            log::debug!("Got captcha token");
                                            self.captcha = axolotl_request.data;
                                            log::debug!("Got captcha token: {:?}", self.captcha);
                                            log::debug!("Getting phone number");
                                            self.get_phone_number().await;
                                        } else if request_type == "requestCode" {
                                            self.phone_number = match axolotl_request.data {
                                                Some(data) => {
                                                    match phonenumber::parse(None, data) {
                                                        Ok(phone_number) => Some(phone_number),
                                                        Err(e) => {
                                                            log::error!(
                                                                "Error parsing phone number: {}",
                                                                e
                                                            );
                                                            None
                                                        }
                                                    }
                                                }
                                                None => None,
                                            };
                                            if self.phone_number.is_some() {
                                                log::debug!(
                                                    "Got phone number: {}",
                                                    self.phone_number.as_ref().unwrap()
                                                );
                                                let (code_tx, code_rx) =
                                                    mpsc::channel(MESSAGE_BOUND);
                                                self.get_phone_pin().await;
                                                match self.get_verification_code(code_rx).await {
                                                    Ok(_) => log::debug!(
                                                        "Success sending verification code"
                                                    ),
                                                    Err(e) => {
                                                        log::error!(
                                                            "Error getting verification code: {}",
                                                            e
                                                        );
                                                        self.get_phone_number().await;
                                                        break;
                                                    }
                                                }
                                                let code_message: Option<
                                                    Result<Message, warp::Error>,
                                                > = r.next().await;
                                                if let Some(code_message) = code_message {
                                                    match code_message {
                                                        Ok(code_message) => {
                                                            if code_message.is_close() {
                                                                log::info!(
                                                                    "Got close message, exiting"
                                                                );
                                                                is_closed = true;
                                                                break;
                                                            } else if code_message.is_text() {
                                                                let text =
                                                                    code_message.to_str().unwrap();
                                                                if let Ok::<
                                                                    AxolotlRequest,
                                                                    SerdeError,
                                                                >(
                                                                    axolotl_request
                                                                ) = serde_json::from_str(text)
                                                                {
                                                                    // Axolotl request
                                                                    let request_type: &str =
                                                                        axolotl_request
                                                                            .request
                                                                            .as_str();
                                                                    log::info!(
                                                                        "Axolotl registration request code message: {}",
                                                                        request_type
                                                                    );
                                                                    if request_type == "sendCode" {
                                                                        let code =
                                                                            axolotl_request.data;
                                                                        if code.is_some() {
                                                                            let code =
                                                                                code.unwrap();
                                                                            log::info!(
                                                                                "Got code: {}",
                                                                                code
                                                                            );
                                                                            match code_tx
                                                                                .send(code.into_boxed_str())
                                                                                .await
                                                                            {
                                                                                Ok(_) => (),
                                                                                Err(e) => {
                                                                                    log::error!("Error sending code to registration manager: {}", e);
                                                                                    break;
                                                                                }
                                                                            };
                                                                        } else {
                                                                            log::error!("No valid code provided");
                                                                            self.get_phone_pin()
                                                                                .await;
                                                                        }
                                                                    }
                                                                }
                                                            }
                                                        }
                                                        Err(e) => {
                                                            log::error!(
                                                                "Error getting message: {}",
                                                                e
                                                            );
                                                            break;
                                                        }
                                                    }
                                                }
                                            } else {
                                                log::error!("No valid phone number provided");
                                                self.get_phone_number().await;
                                            }
                                        }
                                    }
                                    log::info!("Got text message: {}", text);
                                }
                            }
                            Err(e) => {
                                log::error!("Error getting message: {}", e);
                                break;
                            }
                        }
                    }
                    if is_closed {
                        log::info!("Got close message, exiting2");
                        break;
                    }
                }
            } else {
                log::info!("Already registered, lets start the manager1");
                self.send_registration_confirmation().await;
                log::debug!("Registration confirmation sent and done");
            }
            log::debug!("Creating manager");
            let manager = self.manager_thread.lock().await;
            if manager.is_none() {
                log::debug!("Manager is none, creating {:?}", self.receive_error);
                // todo for errors
                if let Some(error_opt) = self.receive_error.recv().await {
                    log::error!("Got error after linking device: {}", error_opt);
                    continue;
                }
            }
            std::mem::drop(manager);

            log::info!("Manager created");

            // While messages come, handle them
            if self.is_registered.is_some() && self.is_registered.unwrap() {
                self.send_registration_confirmation().await;
                let shared_sender_mutex = Arc::clone(&shared_sender);
                let r = self.receive_content.clone();

                thread::spawn(move || {
                    block_on(Handler::handle_received_message(r, shared_sender_mutex))
                });
                self.handle_receiving_messages().await;
            }
        }
        log::debug!("Exiting loop");
        Err(ApplicationError::WebSocketHandleMessageError(
            "Too many errors, exiting".to_string(),
        ))
    }

    async fn check_registration() -> Result<(), ApplicationError> {
        // Check if we are already registered

        // wait 3 seconds for the manager to be initialized

        thread::sleep(time::Duration::from_secs(2));
        let config_store = match Handler::get_config_store().await {
            Ok(c) => c,
            Err(e) => {
                log::error!("Error getting config store: {}", e);
                return Err(ApplicationError::RegistrationError(
                    "Error getting config store".to_string(),
                ));
            }
        };
        if config_store.is_registered() {
            log::info!("Already registered, lets start the manager2");
        } else {
            log::info!("Not registered, lets start the registration");
            return Err(ApplicationError::RegistrationError(
                "Not yet registered".to_string(),
            ));
        }
        Ok(())
    }
    async fn create_provisioning_link(&mut self) -> Result<(), ApplicationError> {
        log::debug!("Creating provisioning link");
        let config_store = match Handler::get_config_store().await {
            Ok(c) => c,
            Err(e) => {
                log::error!("Error getting config store: {}", e);
                return Err(ApplicationError::RegistrationError(
                    "Error getting config store".to_string(),
                ));
            }
        };
        if config_store.is_registered() {
            log::info!("Already registered, lets start the manager4");
            return Ok(());
        }

        let (provisioning_link_tx, provisioning_link_rx) = oneshot::channel();
        self.provisioning_link_rx = Some(Arc::new(Mutex::new(provisioning_link_rx)));
        let (error_tx, error_rx) = oneshot::channel();
        self.error_rx = Some(Arc::new(Mutex::new(error_rx)));
        let (send_content, receive_content) = mpsc::unbounded_channel();
        self.receive_content = Arc::new(Mutex::new(Some(receive_content)));
        let current_chat: Option<Thread> = None;
        let current_chat_mutex = Arc::new(Mutex::new(current_chat));
        let (send_error, receive_error) = mpsc::channel(MESSAGE_BOUND);
        self.receive_error = receive_error;
        log::debug!("Creating runtime");
        thread::spawn(|| {
            Runtime::new().unwrap().spawn(async move {
                log::debug!("Spawning manager thread");
                ManagerThread::new(
                    config_store.clone(),
                    "axolotl".to_string(),
                    provisioning_link_tx,
                    error_tx,
                    send_content,
                    current_chat_mutex,
                    send_error,
                )
                .await;
                log::info!("Manager thread started, ready to receive messages from the client2");
            });
        });
        log::debug!("runtime created");

        Ok(())
    }
    async fn send_provisioning_link(&self) {
        log::debug!("Sending provisioning link");
        if self.provisioning_link.is_none() {
            log::error!("No provisioning link provided");
            return;
        }
        let qr_code = format!(
            "{{\"response_type\":\"qr_code\",\"data\":\"{}\"}}",
            self.provisioning_link.clone().unwrap()
        );
        let mut ws_sender = self.sender.as_ref().unwrap().lock().await;
        match ws_sender.send(Message::text(qr_code)).await {
            Ok(_) => (),
            Err(e) => {
                log::error!("Error sending provisioning link to client: {}", e);
            }
        }
    }

    async fn get_verification_code(
        &mut self,
        mut code_rx: tokio::sync::mpsc::Receiver<Box<str>>,
    ) -> Result<(), ApplicationError> {
        log::debug!("Getting verification code");
        if self.phone_number.is_none() {
            log::error!("No phone number provided");
            return Err(ApplicationError::RegistrationError(
                "No phone number provided".to_string(),
            ));
        }
        if self.captcha.is_none() {
            log::error!("No captcha provided");
            return Err(ApplicationError::RegistrationError(
                "No captcha provided".to_string(),
            ));
        }
        let p = self.phone_number.clone().unwrap();
        let c = self.captcha.as_mut().unwrap().clone();
        // create new error receiver
        let (error_tx, error_rx) = oneshot::channel();

        std::thread::spawn(move || {
            let panic = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
                tokio::runtime::Runtime::new()
                    .expect("Failed to setup runtime")
                    .block_on(async move {
                        let config_store = match Handler::get_config_store().await {
                            Ok(c) => c,
                            Err(e) => {
                                log::error!("Error getting config store: {}", e);
                                return Err(ApplicationError::RegistrationError(
                                    "Error getting config store".to_string(),
                                ));
                            }
                        };
                        log::debug!("Creating manager for registration");
                        let manager = match Manager::register(
                            config_store,
                            RegistrationOptions {
                                signal_servers: presage::prelude::SignalServers::Production,
                                phone_number: p,
                                use_voice_call: false,
                                captcha: Some(c.as_str()),
                                force: true,
                            },
                        )
                        .await
                        {
                            Ok(m) => m,
                            Err(e) => {
                                log::error!("Error requesting pin: {}", e);
                                // todo send error to client
                                error_tx.send(Some(e)).unwrap();
                                return Err(ApplicationError::RegistrationError(
                                    "Error requesting pin".to_string(),
                                ));
                            }
                        };
                        log::debug!("check code_rx for code {:?}", code_rx);
                        let code = match code_rx.recv().await {
                            Some(c) => c,
                            None => {
                                log::error!("No code provided");
                                error_tx.send(None).unwrap();
                                return Err(ApplicationError::RegistrationError(
                                    "No code provided".to_string(),
                                ));
                            }
                        };
                        match manager.confirm_verification_code(code).await {
                            Ok(_) => (),
                            Err(e) => {
                                log::error!("Error confirming pin: {}", e);
                                error_tx.send(Some(e)).unwrap();
                                return Err(ApplicationError::RegistrationError(
                                    "Error confirming pin".to_string(),
                                ));
                            }
                        }
                        drop(error_tx);
                        Ok(())
                    })
                    .unwrap();
            }));
            if panic.is_err() {
                log::error!("Error/panic getting verification code: {:?}", panic);
            }
        });
        log::debug!("Awaiting for error");
        let error = error_rx.await.unwrap();
        if error.is_some() {
            log::debug!("Got error: {:?}", error);
            self.send_error(ApplicationError::Presage(error.unwrap()))
                .await;
            return Err(ApplicationError::RegistrationError(
                "Error getting verification code".to_string(),
            ));
        }
        log::debug!("Getting verification code done");

        Ok(())
    }

    async fn get_phone_number(&self) {
        let message = format!("{{\"response_type\":\"phone_number\",\"data\":\"\"}}",);
        let mut ws_sender = self.sender.as_ref().unwrap().lock().await;
        match ws_sender.send(Message::text(message)).await {
            Ok(_) => (),
            Err(e) => {
                log::error!("Error sending phone number request to client: {}", e);
            }
        }
        std::mem::drop(ws_sender);
    }
    async fn send_error(&self, error: ApplicationError) {
        let mut error_string = error.to_string();
        error_string.pop();
        let message = format!(
            "{{\"response_type\":\"registration_error\",\"data\":\"{}\"}}",
            error_string
        );
        log::debug!("Sending error to client: {}", message);
        let mut ws_sender = self.sender.as_ref().unwrap().lock().await;
        match ws_sender.send(Message::text(message)).await {
            Ok(_) => (),
            Err(e) => {
                log::error!("Error sending error to client: {}", e);
            }
        }
        std::mem::drop(ws_sender);
    }
    async fn get_phone_pin(&self) {
        let message = format!("{{\"response_type\":\"pin\",\"data\":\"\"}}",);
        let mut ws_sender = self.sender.as_ref().unwrap().lock().await;
        match ws_sender.send(Message::text(message)).await {
            Ok(_) => (),
            Err(e) => {
                log::error!("Error sending pin request to client: {}", e);
            }
        }
        std::mem::drop(ws_sender);
    }
    async fn send_registration_confirmation(&self) {
        let qr_code = format!("{{\"response_type\":\"registration_done\",\"data\":\"\"}}");
        let mut ws_sender = self.sender.as_ref().unwrap().lock().await;
        match ws_sender.send(Message::text(qr_code)).await {
            Ok(_) => (),
            Err(e) => {
                log::error!("Error sending registration status done to client: {}", e);
            }
        }
        std::mem::drop(ws_sender);
    }
    async fn handle_receiving_messages(&self) {
        log::info!("Awaiting for received messages");
        if self.receiver.is_none() {
            log::error!("Receiver not initialized");
            return;
        }
        let mut receiver = self.receiver.as_ref().unwrap().lock().await;
        self.send_registration_confirmation().await;

        while let Some(body) = receiver.next().await {
            log::debug!(
                "Got message from axolotl: {:?}, awaitng manager thread lock",
                body
            );
            let manager = self.manager_thread.lock().await;
            log::debug!("Got manager thread lock");

            if manager.is_none() {
                log::error!("Manager not initialized");
                return;
            }
            let message = match body {
                Ok(msg) => msg,
                Err(_) => {
                    continue;
                }
            };
            log::debug!("Asking sender lock");
            let sender = self.sender.clone().unwrap();
            log::debug!("Got sender lock");

            log::debug!("Got websocket message from axolotl-web: {:?}", message);
            match self
                .handle_websocket_message(message, sender, &manager.clone().unwrap())
                .await
            {
                Ok(_) => log::info!("Message handled"),
                Err(e) => log::error!("Error handling message: {}", e),
            };
        }
    }
    async fn handle_provisoning(&mut self) {
        log::info!("Awaiting for provisioning link");
        if self.provisioning_link_rx.is_none() {
            log::error!("Provisioning link receiver not initialized");
            return;
        }
        let mut p = self.provisioning_link_rx.as_ref().unwrap().lock().await;

        while let Ok(mut url) = p.try_recv() {
            url = match url {
                Some(u) => Some(u),
                None => {
                    thread::sleep(time::Duration::from_secs(1));
                    match p.try_recv() {
                        Ok(u) => u,
                        Err(_) => {
                            log::error!("Error getting provisioning link");
                            return;
                        }
                    }
                }
            };

            log::debug!("Got provisioning link: {:?}", url);
            self.provisioning_link = url;
        }
    }
    async fn start_registration(&self) -> Result<(), ApplicationError> {
        log::info!("Starting registration");
        // wait for a sender to be available
        if self.sender.is_none() {
            log::info!("Sender not initialized, waiting for it");

            loop {
                if self.sender.is_some() {
                    break;
                }
                std::thread::sleep(std::time::Duration::from_millis(100));
            }
        }
        let mut mut_sender = self.sender.as_ref().unwrap().lock().await;
        match mut_sender
            .send(Message::text("{\"response_type\":\"registration_start\"}"))
            .await
        {
            Ok(_) => log::info!("Sent registration start message to client"),
            Err(e) => {
                log::error!("Error sending registration start message: {}", e)
            }
        };
        Ok(())
    }
    async fn thread_metadata(
        &self,
        thread: &Thread,
    ) -> Result<Option<ThreadMetadata>, ApplicationError> {
        let mut manager = self.manager_thread.lock().await;
        let manager = manager.as_mut().unwrap();
        match manager.thread_metadata(thread).await {
            Ok(metadata) => Ok(metadata),
            Err(e) => Err(ApplicationError::from(e)),
        }
    }
    async fn create_thread_metadata(&self, thread: &Thread) -> Result<(), ApplicationError> {
        let metadata = ThreadMetadata {
            thread: thread.clone(),
            unread_messages_count: 0,
            last_message: None,
            title: None,
            archived: false,
            muted: false,
        };

        let mut manager = self.manager_thread.lock().await;
        let manager = manager.as_mut().unwrap();
        match manager.save_thread_metadata(metadata).await {
            Ok(_) => Ok(()),
            Err(e) => Err(ApplicationError::from(e)),
        }
    }
    async fn handle_received_message(
        receive: Arc<Mutex<Option<UnboundedReceiver<Content>>>>,
        sender: Arc<Mutex<SplitSink<WebSocket, warp::ws::Message>>>,
    ) {
        log::info!("Awaiting for received message");
        let mut receive = match receive.lock().await.take() {
            Some(r) => r,
            None => {
                log::error!("receiver is not initalised");
                // TODO: reinitialise receiver or use a receiver that doesn't need a take
                return;
            }
        };
        log::debug!("Got receive lock");

        loop {
            match receive.recv().await {
                Some(content) => {
                    let thread = Thread::try_from(&content).unwrap();
                    let mut axolotl_message = AxolotlMessage::from_message(content);
                    axolotl_message.thread_id = Some(thread);
                    let axolotl_message_json = serde_json::to_string(&axolotl_message).unwrap();
                    let response_type = "message_received".to_string();
                    let response = AxolotlResponse {
                        response_type,
                        data: axolotl_message_json,
                    };
                    let response = serde_json::to_string(&response).unwrap();

                    let mut ws_sender = sender.lock().await;
                    match ws_sender.send(Message::text(response)).await {
                        Ok(_) => (),
                        Err(e) => {
                            log::error!("Error sending message to client: {}", e);
                        }
                    }
                    std::mem::drop(ws_sender);
                }
                None => {
                    log::error!("Error receiving message");
                    break;
                }
            };
        }
    }
    pub fn write_as_json<I, P, W>(&self, out: &mut W, groups: I) -> serde_json::Result<()>
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
    async fn handle_get_contacts(
        &self,
        manager: &ManagerThread,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Getting contacts");
        // Todo: update contacts from profile
        match manager.update_contacts_from_profile().await {
            Ok(_) => (),
            Err(e) => {
                log::error!(
                    "handle_get_contacts: Error updating contacts from profile: {}",
                    e
                );
            }
        }
        let contacts = manager.get_contacts().await.ok().unwrap();
        let mut out = Vec::new();
        self.write_as_json(&mut out, contacts)?;
        let response = AxolotlResponse {
            response_type: "contact_list".to_string(),
            data: String::from_utf8(out).unwrap(),
        };
        Ok(Some(response))
    }
    async fn handle_chat_list(
        &self,
        manager: &ManagerThread,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Getting chat list");
        let conversations = manager.get_conversations().await.ok().unwrap();
        let mut out = Vec::new();
        self.write_as_json(&mut out, conversations)?;
        let response = AxolotlResponse {
            response_type: "chat_list".to_string(),
            data: String::from_utf8(out).unwrap(),
        };
        Ok(Some(response))
    }
    fn string_to_thread(&self, thread_id: &String) -> Result<Thread, ApplicationError> {
        let thread = thread_id
            .to_string()
            .replace(&['{', '}', '\"', '[', ']', ' '][..], "");
        let thread = thread.split(":").collect::<Vec<&str>>();
        log::debug!("Thread: {:?}", thread);
        let thread = match thread[0] {
            "Contact" => Thread::Contact(Uuid::parse_str(thread[1]).unwrap()),
            "Group" => {
                let decoded: Vec<u8> = thread[1]
                    .split(",")
                    .map(|s| s.parse().expect("parse error"))
                    .collect();
                // transform decoded to [u8; 32]
                let mut res = [0u8; 32];
                res.copy_from_slice(&decoded);

                Thread::Group(res)
            }
            _ => return Err(ApplicationError::InvalidRequest),
        };
        Ok(thread)
    }

    async fn handle_upload_attachment(
        &self,
        manager: &ManagerThread,

        data: Option<String>,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Uploading attachment.");
        let data = data.ok_or(ApplicationError::InvalidRequest)?;
        match serde_json::from_str(&data) {
            Ok::<UploadAttachmentRequest, SerdeError>(upload_attachment_request) => {
                log::debug!("Attachment request parsed.");
                let thread = self.string_to_thread(&upload_attachment_request.recipient)?;

                let data_attachment = match DataUrl::process(&upload_attachment_request.attachment)
                {
                    Ok(d) => d,
                    Err(e) => {
                        log::error!("Error while reading data URL. {:?}", e);
                        return Err(ApplicationError::InvalidRequest);
                    }
                };
                let (body, _fragment) = data_attachment.decode_to_vec().unwrap();
                let decoded_attachment: Vec<u8> = body;
                let decoded_tosave_attachment = decoded_attachment.clone();

                let attachment_spec = AttachmentSpec {
                    content_type: data_attachment.mime_type().to_string(),
                    length: decoded_attachment.len(),
                    file_name: None,
                    preview: None,
                    voice_note: None,
                    borderless: None,
                    width: None,
                    height: None,
                    caption: None,
                    blur_hash: None,
                };

                let attachments: Vec<(AttachmentSpec, Vec<u8>)> =
                    vec![(attachment_spec, decoded_attachment)];
                let upload_response = manager.upload_attachments(attachments).await;

                match upload_response {
                    Ok(attachments_pointers) => {
                        let mut pointers: Vec<AttachmentPointer> = Vec::new();
                        for attachment_pointer in attachments_pointers {
                            if let Ok(p) = attachment_pointer {
                                pointers.push(p);
                            }
                        }

                        // We send one attachment at a time
                        // Use its CdnId as filename
                        if pointers.len() > 0 {
                            let cdnid = match pointers[0].attachment_identifier.clone().unwrap() {
                                AttachmentIdentifier::CdnId(id) => id,
                                _ => {
                                    log::debug!(
                                        "Attachment identifier: {:?}",
                                        pointers[0].attachment_identifier.clone().unwrap()
                                    );
                                    log::error!("The uploaded attachment has no identifier.");
                                    return Ok(Some(AxolotlResponse {
                                        response_type: "attachment_not_sent".to_string(),
                                        data: "{\"success: false\"}".to_string(),
                                    }));
                                }
                            };
                            save_attachment(&decoded_tosave_attachment, &cdnid.to_string());
                            send_message(thread, None, Some(pointers), &manager, "attachment_sent")
                                .await?;
                        } else {
                            log::error!("Error while sending attachment.");
                            return Ok(Some(AxolotlResponse {
                                response_type: "attachment_not_sent".to_string(),
                                data: "{\"success: false\"}".to_string(),
                            }));
                        }
                    }
                    Err(e) => {
                        log::error!("Error while uploading attachment. {:?}", e);
                        return Ok(Some(AxolotlResponse {
                            response_type: "attachment_not_sent".to_string(),
                            data: "{\"success: false\"}".to_string(),
                        }));
                    }
                };
                Ok(None)
            }
            Err(e) => {
                log::error!("Error while parsing the attachment request. {:?}", e);
                Err(ApplicationError::InvalidRequest)
            }
        }
    }
    // #[cfg(feature = "ut")]
    async fn handle_upload_attachment_ut(
        &self,
        manager: &ManagerThread,

        data: Option<String>,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Uploading attachment.");
        let data = data.ok_or(ApplicationError::InvalidRequest)?;
        match serde_json::from_str(&data) {
            Ok::<UploadAttachmentUtRequest, SerdeError>(upload_attachment_request) => {
                log::debug!("Attachment request parsed.");
                let thread = self.string_to_thread(&upload_attachment_request.recipient)?;

                let data_attachment = read_a_file(upload_attachment_request.path).unwrap();
                let decoded_attachment: Vec<u8> = data_attachment;
                let decoded_tosave_attachment = decoded_attachment.clone();

                let attachment_spec = AttachmentSpec {
                    content_type: upload_attachment_request.mimetype,
                    length: decoded_attachment.len(),
                    file_name: None,
                    preview: None,
                    voice_note: None,
                    borderless: None,
                    width: None,
                    height: None,
                    caption: None,
                    blur_hash: None,
                };

                let attachments: Vec<(AttachmentSpec, Vec<u8>)> =
                    vec![(attachment_spec, decoded_attachment)];
                let upload_response = manager.upload_attachments(attachments).await;

                match upload_response {
                    Ok(attachments_pointers) => {
                        let mut pointers: Vec<AttachmentPointer> = Vec::new();
                        for attachment_pointer in attachments_pointers {
                            if let Ok(p) = attachment_pointer {
                                pointers.push(p);
                            }
                        }

                        // We send one attachment at a time
                        // Use its CdnId as filename
                        if pointers.len() > 0 {
                            let cdnid = match pointers[0].attachment_identifier.clone().unwrap() {
                                AttachmentIdentifier::CdnId(id) => id,
                                _ => {
                                    log::error!("The uploaded attachment has no identifier.");
                                    return Ok(Some(AxolotlResponse {
                                        response_type: "attachment_not_sent".to_string(),
                                        data: "{\"success: false\"}".to_string(),
                                    }));
                                }
                            };
                            save_attachment(&decoded_tosave_attachment, &cdnid.to_string());
                            send_message(thread, None, Some(pointers), &manager, "attachment_sent")
                                .await?;
                        } else {
                            log::error!("Error while sending attachment.");
                            return Ok(Some(AxolotlResponse {
                                response_type: "attachment_not_sent".to_string(),
                                data: "{\"success: false\"}".to_string(),
                            }));
                        }
                    }
                    Err(e) => {
                        log::error!("Error while uploading attachment. {:?}", e);
                        return Ok(Some(AxolotlResponse {
                            response_type: "attachment_not_sent".to_string(),
                            data: "{\"success: false\"}".to_string(),
                        }));
                    }
                };
                Ok(None)
            }
            Err(e) => {
                log::error!("Error while parsing the attachment request. {:?}", e);
                Err(ApplicationError::InvalidRequest)
            }
        }
    }
    async fn handle_get_message_list(
        &self,
        manager: &ManagerThread,
        data: Option<String>,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Getting message list");
        let data = data.ok_or(ApplicationError::InvalidRequest)?;
        if let Ok::<GetMessagesRequest, SerdeError>(messages_request) = serde_json::from_str(&data)
        {
            let thread: Thread = self.string_to_thread(&messages_request.id)?;
            // match thread {
            //     Thread::Contact(_) => {
            //         manager.update_cotacts_from_profile().await.ok().unwrap();
            //     }
            //     _ => {}
            // }
            let thread_metadata = manager.thread_metadata(&thread).await.ok().unwrap();
            if thread_metadata.is_none() {
                self.create_thread_metadata(&thread).await.ok().unwrap();
            } else {
                let mut thread_metadata = thread_metadata.unwrap();
                match thread_metadata.title.clone() {
                    Some(title) => {
                        // check if title is a valid uuid
                        if Uuid::parse_str(&title).is_ok() {
                            let uuid = Uuid::parse_str(&title).unwrap();
                            match manager.update_contact_from_profile(uuid).await {
                                Ok(_) => {
                                    thread_metadata = manager
                                        .thread_metadata(&thread)
                                        .await
                                        .ok()
                                        .unwrap()
                                        .unwrap();
                                }
                                Err(e) => {
                                    log::error!("handle_get_message_list: Error updating contacts from profile: {}", e);
                                }
                            }
                        }
                    }
                    None => match manager.update_contacts_from_profile().await {
                        Ok(_) => (),
                        Err(e) => {
                            log::error!("handle_get_message_list_2: Error updating contacts from profile: {}", e);
                        }
                    },
                }
                thread_metadata.unread_messages_count = 0;
                manager
                    .save_thread_metadata(thread_metadata)
                    .await
                    .ok()
                    .unwrap();
            }

            let messages = manager.messages(thread, (Unbounded, Unbounded)).await;
            if messages.is_err() {
                log::error!("Failed to load last messages: {}", messages.err().unwrap());
                return Err(ApplicationError::InvalidRequest);
            }
            let messages = messages.unwrap();

            let mut axolotl_messages: Vec<AxolotlMessage> = Vec::new();
            for message in messages {
                match message {
                    Ok(m) => axolotl_messages.push(AxolotlMessage::from_message(m)),
                    Err(e) => {
                        log::error!("Error getting message: {}", e);
                        log::debug!("ignoring error");
                        // Err(ApplicationError::InvalidRequest)?;
                    }
                }
            }

            let mut out = Vec::new();
            self.write_as_json(&mut out, axolotl_messages)?;
            let response = AxolotlResponse {
                response_type: "message_list".to_string(),
                data: String::from_utf8(out).unwrap(),
            };
            Ok(Some(response))
        } else {
            log::debug!("Invalid request: {}", data);
            Err(ApplicationError::InvalidRequest)
        }
    }

    fn handle_ping(&self) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Got ping");
        let response = AxolotlResponse {
            response_type: "pong".to_string(),
            data: "".to_string(),
        };
        Ok(Some(response))
    }
    async fn handle_get_profile(
        &self,
        manager: &ManagerThread,
        data: Option<String>,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Getting profile");
        match data.clone() {
            Some(u_data) => {
                let profile_request: ProfileRequest = match serde_json::from_str(&u_data) {
                    Ok(request) => request,
                    Err(_) => return Err(ApplicationError::InvalidRequest),
                };
                let uuid = Uuid::parse_str(&profile_request.id).unwrap();
                log::debug!("Getting profile for: {}", uuid.to_string());
                let profile = manager.get_contact_by_id(uuid).await.ok().unwrap();
                let mut profile = match profile {
                    Some(p) => p,
                    None => {
                        //request contact sync
                        manager.request_contacts_sync().await.ok().unwrap();
                        return Err(ApplicationError::InvalidRequest);
                    }
                };
                if profile.name == "".to_string() {
                    //request contact sync
                    log::debug!("Updating contact from profile");
                    manager
                        .update_contact_from_profile(profile.uuid)
                        .await
                        .ok()
                        .unwrap();
                    profile = manager.get_contact_by_id(uuid).await.ok().unwrap().unwrap();
                    log::debug!("Updated contact {:?}", profile);
                }
                let response = AxolotlResponse {
                    response_type: "profile".to_string(),
                    data: serde_json::to_string(&profile).unwrap(),
                };
                Ok(Some(response))
            }
            None => {
                manager.update_contacts_from_profile().await.ok().unwrap();
                Ok(None)
            }
        }
    }
    async fn handle_send_message(
        &self,
        manager: &ManagerThread,
        data: Option<String>,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
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
                let thread = self.string_to_thread(&send_message_request.recipient)?;
                let result = match thread {
                    Thread::Contact(contact) => {
                        let message = ContentBody::DataMessage(data_message.clone());
                        manager
                            .send_message(contact, message.clone(), timestamp)
                            .await
                    }
                    Thread::Group(group) => {
                        let group_from_store = manager.get_group(group.clone()).await.ok().unwrap();
                        match group_from_store {
                            None => {
                                log::error!("Group not found");
                                return Err(ApplicationError::InvalidRequest);
                            }
                            Some(group_from_store) => {
                                let mut group_data_message = data_message.clone();
                                group_data_message.group_v2 = Some(GroupContextV2 {
                                    master_key: Some(group.to_vec()),
                                    group_change: None,
                                    revision: Some(group_from_store.revision),
                                });
                                manager
                                    .send_message_to_group(group, group_data_message, timestamp)
                                    .await
                            }
                        }
                    }
                };
                let is_failed = result.is_err();
                if is_failed {
                    log::error!("Error while sending the message. {:?}", result.err());
                }
                let mut message = AxolotlMessage::from_data_message(data_message);
                message.thread_id = Some(thread);
                // message.sender = Some(manager.uuid());
                let response_data = SendMessageResponse { message, is_failed };
                let response_data_json = serde_json::to_string(&response_data).unwrap();
                let response = AxolotlResponse {
                    response_type: "message_sent".to_string(),
                    data: response_data_json,
                };
                Ok(Some(response))
            }
            Err(e) => {
                log::error!("Error while parsing the request. {:?}", e);
                Err(ApplicationError::InvalidRequest)
            }
        }
    }
    async fn handle_open_chat(
        &self,
        manager: &ManagerThread,
        data: Option<String>,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        let data = data.ok_or(ApplicationError::InvalidRequest)?;
        let thread: Thread = match serde_json::from_str(data.as_str()) {
            Ok(thread) => thread,
            Err(_) => return Err(ApplicationError::InvalidRequest),
        };
        match thread {
            Thread::Contact(_contact) => {
                // manager.update_contact_from_profile(contact).await.ok().unwrap();
            }
            _ => {}
        }
        manager.open_chat(thread).await.ok().unwrap();
        let response = AxolotlResponse {
            response_type: "ping".to_string(),
            data: "".to_string(),
        };
        Ok(Some(response))
    }
    async fn handle_close_chat(
        &self,
        manager: &ManagerThread,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        manager.close_chat().await.ok().unwrap();
        Ok(None)
    }
    async fn handle_get_config(
        &self,
        _manager: &ManagerThread,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Getting config");
        // let my_uuid = manager.uuid();
        let mut platform = "linux".to_string();
        #[cfg(target_os = "windows")]
        {
            platform = "windows".to_string();
        }
        #[cfg(target_os = "macos")]
        {
            platform = "macos".to_string();
        }
        #[cfg(target_os = "android")]
        {
            platform = "android".to_string();
        }
        #[cfg(target_os = "ios")]
        {
            platform = "ios".to_string();
        }
        let mut feature = "".to_string();

        #[cfg(feature = "tauri")]
        {
            feature = "tauri".to_string();
        }
        #[cfg(feature = "ut")]
        {
            feature = "ut".to_string();
        }

        let config = AxolotlConfig {
            uuid: None,
            e164: None,
            platform: Some(platform),
            feature: Some(feature),
        };
        let response = AxolotlResponse {
            response_type: "config".to_string(),
            data: serde_json::to_string(&config).unwrap(),
        };
        Ok(Some(response))
    }
    async fn handle_unregister(
        &self,
        manager: &ManagerThread,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Unregistering");
        let mut store = Handler::get_config_store().await?;
        store.clear_registration()?;
        exit(0);
    }
    async fn handle_change_notifications_for_thread(
        &self,
        manager: &ManagerThread,
        data: Option<String>,
    ) -> Result<Option<AxolotlResponse>, ApplicationError> {
        log::info!("Changing notifications for thread");
        let data = data.ok_or(ApplicationError::InvalidRequest)?;
        if let Ok::<ChangeNotificationsForThreadRequest, SerdeError>(
            change_notifications_for_thread_request,
        ) = serde_json::from_str(&data)
        {
            let thread_metadata = manager
                .thread_metadata(&change_notifications_for_thread_request.thread)
                .await
                .ok()
                .unwrap();
            if thread_metadata.is_none() {
                self.create_thread_metadata(&change_notifications_for_thread_request.thread)
                    .await
                    .ok()
                    .unwrap();
            } else {
                let mut thread_metadata = thread_metadata.unwrap();
                thread_metadata.muted = change_notifications_for_thread_request.muted;
                manager
                    .save_thread_metadata(thread_metadata)
                    .await
                    .ok()
                    .unwrap();
            }
            let response = AxolotlResponse {
                response_type: "ping".to_string(),
                data: "".to_string(),
            };
            Ok(Some(response))
        } else {
            log::debug!("Invalid request: {}", data);
            Err(ApplicationError::InvalidRequest)
        }
    }
    /// Handles a websocket message
    async fn handle_websocket_message(
        &self,
        message: Message,
        mutex_sender: Arc<Mutex<SplitSink<WebSocket, warp::ws::Message>>>,
        manager: &ManagerThread,
    ) -> Result<(), ApplicationError> {
        // Skip any non-Text messages...
        let msg = if let Ok(s) = message.to_str() {
            s
        } else {
            "Invalid message"
        };
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
                "getContacts" => self.handle_get_contacts(manager).await,
                "getChatList" => self.handle_chat_list(manager).await,
                "getMessageList" => {
                    self.handle_get_message_list(manager, axolotl_request.data)
                        .await
                }
                "ping" => self.handle_ping(),
                "sendMessage" => {
                    self.handle_send_message(manager, axolotl_request.data)
                        .await
                }
                "uploadAttachment" => {
                    self.handle_upload_attachment(manager, axolotl_request.data)
                        .await
                }
                // #[cfg(feature = "ut")]
                "sendAttachment" => {
                    self.handle_upload_attachment_ut(manager, axolotl_request.data)
                        .await
                }
                "openChat" => self.handle_open_chat(manager, axolotl_request.data).await,
                "leaveChat" => self.handle_close_chat(manager).await,
                "getConfig" => self.handle_get_config(manager).await,
                "unregister" => self.handle_unregister(manager).await,
                "changeNotificationsForThread" => {
                    self.handle_change_notifications_for_thread(manager, axolotl_request.data)
                        .await
                }
                "getProfile" => self.handle_get_profile(manager, axolotl_request.data).await,
                _ => {
                    log::error!("Unhandled axolotl request {}", axolotl_request.request);
                    Err(ApplicationError::InvalidRequest)
                }
            };
            match response {
                Ok(Some(response)) => {
                    let mut unlocked_sender = mutex_sender.lock().await;
                    match unlocked_sender
                        .send(Message::text(serde_json::to_string(&response).unwrap()))
                        .await
                    {
                        Ok(_) => {}
                        Err(e) => {
                            log::error!("Error while sending response. {:?}", e);
                        }
                    };

                    std::mem::drop(unlocked_sender);
                }
                Ok(None) => {} //drop the message
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
}

/// Save a file on the disk
pub fn save_attachment(file_content: &[u8], file_name: &str) {
    // Create the attachments directory if needed
    let _ = fs::create_dir_all(&format!("{}/attachments/", get_app_dir()));

    let mut file = fs::OpenOptions::new()
        .create(true)
        .write(true)
        .open(&format!("{}/attachments/{}", get_app_dir(), file_name))
        .unwrap();

    let file_written = file.write_all(file_content);
    if let Err(e) = file_written {
        log::error!("Error while saving attachment. {:?}", e)
    }
}

/// Returns the path <configPath>/textsecure.nanuc
/// Example: ~/.config/textsecure.nanuc
pub fn get_app_dir() -> String {
    let config_path = dirs::config_dir()
        .unwrap()
        .into_os_string()
        .into_string()
        .unwrap();
    format!("{}/textsecure.nanuc", config_path)
}

fn read_a_file(file_path: String) -> std::io::Result<Vec<u8>> {
    let mut file = File::open(file_path)?;

    let mut data = Vec::new();
    file.read_to_end(&mut data)?;

    return Ok(data);
}
