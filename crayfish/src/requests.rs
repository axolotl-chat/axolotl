pub mod registration;

use crate::error::Result;

use libsignal_service::configuration::{ServiceCredentials, SignalServers};
use libsignal_service::provisioning::{ConfirmCodeResponse, VerificationCodeResponse};
use libsignal_service::USER_AGENT;
use libsignal_service_hyper::prelude::HyperPushService;
use phonenumber::PhoneNumber;
use tokio::sync::{mpsc, oneshot};

use self::registration::{register_user, verify_user, ConfirmRegistration, Register};

pub type RequestSender = mpsc::Sender<Request>;

pub type Callback<T> = oneshot::Sender<T>;

pub enum Request {
    Register(Register, Callback<Result<VerificationCodeResponse>>),
    ConfirmRegistration(ConfirmRegistration, Callback<Result<ConfirmCodeResponse>>),
}

pub async fn handle_requests(mut rx: mpsc::Receiver<Request>) {
    while let Some(req) = rx.recv().await {
        process_request(req).await;
    }
}

pub async fn process_request(req: Request) {
    match req {
        Request::Register(data, cb) => {
            let _ = cb.send(register_user(data).await);
        }
        Request::ConfirmRegistration(data, cb) => {
            let _ = cb.send(verify_user(data).await);
        }
    };
}

pub fn service_with_credentials(credentials: ServiceCredentials) -> HyperPushService {
    HyperPushService::new(
        SignalServers::Production,
        Some(credentials),
        USER_AGENT.into(),
    )
}

pub fn service_with_number(phonenumber: PhoneNumber, password: String) -> HyperPushService {
    service_with_credentials(ServiceCredentials {
        uuid: None,
        phonenumber,
        password: Some(password),
        signaling_key: None,
        device_id: None,
    })
}
