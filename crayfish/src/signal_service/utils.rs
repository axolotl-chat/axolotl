use libsignal_service::configuration::{ServiceCredentials, SignalServers};
use libsignal_service::USER_AGENT;
use libsignal_service_hyper::prelude::HyperPushService;
use phonenumber::PhoneNumber;

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
