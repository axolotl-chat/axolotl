use libsignal_service::configuration::ServiceCredentials;
use libsignal_service::provisioning::{
    generate_registration_id, ConfirmCodeMessage, ConfirmCodeResponse, ProvisioningManager,
    VerificationCodeResponse,
};
use libsignal_service_actix::prelude::AwcPushService;
use phonenumber::PhoneNumber;
use serde::Deserialize;
use serde_big_array::BigArray;
use std::str::FromStr;

use super::authenticated_service_with_credentials;
use super::Result;

#[derive(Deserialize, Debug)]
pub struct Register {
    pub number: String,
    pub password: String,
    pub captcha: String,
    pub use_voice: bool,
}

#[derive(Deserialize, Debug)]
pub struct ConfirmRegistration {
    pub number: String,
    pub password: String,
    pub confirm_code: u32,
    #[serde(with = "BigArray")]
    pub signaling_key: [u8; 52],
}

pub async fn register_user(data: Register) -> Result<VerificationCodeResponse> {
    // TODO move this parsing into the deserialization
    let phonenumber = PhoneNumber::from_str(&data.number).map_err(|e| e.to_string())?;

    let mut push_service = authenticated_service_with_credentials(ServiceCredentials {
        uuid: None,
        phonenumber: phonenumber.clone(),
        password: Some(data.password.to_string()),
        signaling_key: None,
        device_id: None, // !77
    });
    let mut provisioning_manager: ProvisioningManager<AwcPushService> = ProvisioningManager::new(
        &mut push_service,
        phonenumber.clone(),
        data.password.to_string(),
    );

    if data.use_voice {
        provisioning_manager
            .request_voice_verification_code(Some(&data.captcha), None)
            .await
            .map_err(|e| e.to_string())
    } else {
        provisioning_manager
            .request_sms_verification_code(Some(&data.captcha), None)
            .await
            .map_err(|e| e.to_string())
    }
}

pub async fn verify_user(data: ConfirmRegistration) -> Result<ConfirmCodeResponse> {
    let phonenumber = PhoneNumber::from_str(&data.number).map_err(|e| e.to_string())?;
    let registration_id = generate_registration_id(&mut rand::thread_rng());

    let mut push_service = authenticated_service_with_credentials(ServiceCredentials {
        uuid: None,
        phonenumber: phonenumber.clone(),
        password: Some(data.password.to_string()),
        signaling_key: None,
        device_id: None, // !77
    });

    let mut provisioning_manager = ProvisioningManager::<AwcPushService>::new(
        &mut push_service,
        phonenumber.clone(),
        data.password.to_string(),
    );
    provisioning_manager
        .confirm_verification_code(
            data.confirm_code,
            ConfirmCodeMessage::new_without_unidentified_access(
                data.signaling_key.to_vec(),
                registration_id,
            ),
        )
        .await
        .map_err(|e| e.to_string())
}
