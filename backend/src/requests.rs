pub mod registration;

use libsignal_service::configuration::{ServiceCredentials, SignalServers};
use libsignal_service::provisioning::{ConfirmCodeResponse, VerificationCodeResponse};
use libsignal_service::USER_AGENT;
use libsignal_service_actix::prelude::AwcPushService;
use tokio::sync::mpsc;
use tokio::sync::oneshot;

use self::registration::{register_user, verify_user, ConfirmRegistration, Register};

pub type Result<T> = std::result::Result<T, String>;
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

pub fn authenticated_service_with_credentials(credentials: ServiceCredentials) -> AwcPushService {
    // TODO switch to production when working
    AwcPushService::new(SignalServers::Staging, Some(credentials), USER_AGENT.into())
}
