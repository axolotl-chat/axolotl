use libsignal_service::configuration::{ServiceCredentials, SignalServers};
use libsignal_service::USER_AGENT;
use libsignal_service_actix::prelude::AwcPushService;

pub type Result<T> = std::result::Result<T, String>;

pub fn authenticated_service_with_credentials(credentials: ServiceCredentials) -> AwcPushService {
    // TODO switch to production when working
    AwcPushService::new(SignalServers::Staging, Some(credentials), USER_AGENT.into())
}
