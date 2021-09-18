use super::types::*;
use crate::error::{Error, Result};
use crate::requests::registration::{self, register_user, verify_user};

use libsignal_service::provisioning::VerificationCodeResponse;
use warp::ws::Message;

use super::types::ErrorMessage;

pub async fn handle_message(msg: Message) -> Result<Message> {
    let content: &str = msg
        .to_str()
        .map_err(|_| Error::new("Message is not text. Failed to parse."))?;

    let type_check: RequestMessage<TypeCheck> = serde_json::from_str(content)?;

    if type_check.direction != CRAYFISH_WEBSOCKET_TYPE_REQUEST {
        Error::is("Received Message is no valid request")?
    }

    let response = match type_check.request.message_type {
        CRAYFISH_WEBSOCKET_MESSAGE_REGISTRATION => handle_registration(content).await,
        CRAYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION => {
            handle_registration_confirm(content).await
        }
        _ => Error::is("Received request type is unknown"),
    };

    response.or_else(|e| {
        Ok(NestedResponse::new_msg(
            ErrorMessage::from(e),
            type_check.request.message_type,
        ))
    })
}

async fn handle_registration(content: &str) -> Result<Message> {
    println!("Handle registration message");
    let req: NestedRequest<registration::Register> = serde_json::from_str(content)?;

    match register_user(req.request.message).await? {
        VerificationCodeResponse::Issued => {
            println!("Registration request was sent");
            Ok(NestedResponse::new_msg(
                Success { success: true },
                CRAYFISH_WEBSOCKET_MESSAGE_REGISTRATION,
            ))
        }
        VerificationCodeResponse::CaptchaRequired => {
            println!("Server requires a Captcha. Please generate one and try again!");
            Error::is("Captcha required")
        }
    }
}

async fn handle_registration_confirm(content: &str) -> Result<Message> {
    println!("Handle registration confirmation message");
    let req: NestedRequest<registration::ConfirmRegistration> = serde_json::from_str(content)?;

    let response_data = verify_user(req.request.message).await?;

    Ok(NestedResponse::new_msg(
        RegistrationData {
            uuid: *response_data.uuid.as_bytes(),
            storage_capable: response_data.storage_capable,
        },
        CRAYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION,
    ))
}
