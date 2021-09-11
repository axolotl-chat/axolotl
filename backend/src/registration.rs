use libsignal_service::configuration::ServiceCredentials;
use libsignal_service::provisioning::{
    generate_registration_id, ConfirmCodeMessage, ConfirmCodeResponse, ProvisioningManager,
    VerificationCodeResponse,
};
use libsignal_service_actix::prelude::AwcPushService;
use phonenumber::PhoneNumber;

use super::requests::authenticated_service_with_credentials;
use super::requests::Result;

pub async fn register_user(
    phonenumber: &PhoneNumber,
    password: &str,
    captcha: Option<String>,
    use_voice: bool,
) -> Result<VerificationCodeResponse> {
    let mut push_service = authenticated_service_with_credentials(ServiceCredentials {
        uuid: None,
        phonenumber: phonenumber.clone(),
        password: Some(password.to_string()),
        signaling_key: None,
        device_id: None, // !77
    });
    let mut provisioning_manager: ProvisioningManager<AwcPushService> =
        ProvisioningManager::new(&mut push_service, phonenumber.clone(), password.to_string());

    if use_voice {
        provisioning_manager
            .request_voice_verification_code(captcha.as_deref(), None)
            .await
            .map_err(|e| e.to_string())
    } else {
        provisioning_manager
            .request_sms_verification_code(captcha.as_deref(), None)
            .await
            .map_err(|e| e.to_string())
    }
}

pub async fn verify_user(
    phonenumber: &PhoneNumber,
    password: &str,
    confirm_code: u32,
    signaling_key: [u8; 52],
) -> Result<ConfirmCodeResponse> {
    let registration_id = generate_registration_id(&mut rand::thread_rng());

    let mut push_service = authenticated_service_with_credentials(ServiceCredentials {
        uuid: None,
        phonenumber: phonenumber.clone(),
        password: Some(password.to_string()),
        signaling_key: None,
        device_id: None, // !77
    });

    let mut provisioning_manager = ProvisioningManager::<AwcPushService>::new(
        &mut push_service,
        phonenumber.clone(),
        password.to_string(),
    );
    provisioning_manager
        .confirm_verification_code(
            confirm_code,
            ConfirmCodeMessage::new_without_unidentified_access(
                signaling_key.to_vec(),
                registration_id,
            ),
        )
        .await
        .map_err(|e| e.to_string())
}
