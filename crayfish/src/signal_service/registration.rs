use super::utils::service_with_number;
use crate::error::Result;

use libsignal_service::provisioning::{
    generate_registration_id, ConfirmCodeMessage, ConfirmCodeResponse, ProvisioningManager,
    VerificationCodeResponse,
};
use phonenumber::PhoneNumber;
use serde::Deserialize;
use serde_big_array::BigArray;

#[derive(Deserialize, Debug)]
pub struct Register {
    #[serde(with = "serde_str")]
    pub number: PhoneNumber,
    pub password: String,
    pub captcha: String,
    pub use_voice: bool,
}

#[derive(Deserialize, Debug)]
pub struct ConfirmRegistration {
    #[serde(with = "serde_str")]
    pub number: PhoneNumber,
    pub password: String,
    pub confirm_code: u32,
    #[serde(with = "BigArray")]
    pub signaling_key: [u8; 52],
}

pub async fn register_user(data: Register) -> Result<VerificationCodeResponse> {
    let mut push_service = service_with_number(data.number.clone(), data.password.clone());
    let mut provisioning_manager =
        ProvisioningManager::new(&mut push_service, data.number, data.password);

    Ok(if data.use_voice {
        provisioning_manager
            .request_voice_verification_code(Some(&data.captcha), None)
            .await?
    } else {
        provisioning_manager
            .request_sms_verification_code(Some(&data.captcha), None)
            .await?
    })
}

pub async fn verify_user(data: ConfirmRegistration) -> Result<ConfirmCodeResponse> {
    let registration_id = generate_registration_id(&mut rand::thread_rng());
    let mut push_service = service_with_number(data.number.clone(), data.password.clone());
    let mut provisioning_manager =
        ProvisioningManager::new(&mut push_service, data.number, data.password);

    Ok(provisioning_manager
        .confirm_verification_code(
            data.confirm_code,
            ConfirmCodeMessage::new_without_unidentified_access(
                data.signaling_key.to_vec(),
                registration_id,
            ),
        )
        .await?)
}
