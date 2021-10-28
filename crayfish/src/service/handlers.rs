use super::types::*;
use crate::error::{Error, Result};
use crate::signal_service::registration::{ConfirmRegistration, Register};
use crate::signal_service::requests::Request;
use crate::signal_service::{sealedsender, Queue};
// , decryptSealedMessage};

use libsignal_service::provisioning::VerificationCodeResponse;

use tokio::sync::oneshot;
use warp::ws::Message;

use super::types::ErrorMessage;

pub async fn handle_message(msg: Message, queue: Queue) -> Result<Message> {
    let content: &str = msg
        .to_str()
        .map_err(|_| Error::new("Message is not text. Failed to parse."))?;

    let type_check: RequestMessage<TypeCheck> = serde_json::from_str(content)?;

    if type_check.direction != CRAYFISH_WEBSOCKET_TYPE_REQUEST {
        Error::is("Received Message is no valid request")?
    }

    let response = match type_check.request.message_type {
        CRAYFISH_WEBSOCKET_MESSAGE_REGISTRATION => handle_registration(content, queue).await,
        CRAYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION => {
            handle_registration_confirm(content, queue).await
        }
        CRAYFISH_WEBSOCKET_MESSAGE_SEALED_SENDER_DECRYPT => {
            handle_sealed_sender_decrypt(content, queue).await
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

async fn handle_registration(content: &str, queue: Queue) -> Result<Message> {
    println!("Handle registration message");

    let (tx, rx) = oneshot::channel();
    let nested: NestedRequest<Register> = serde_json::from_str(content)?;
    let req = Request::Register(nested.request.message, tx);

    queue.send(req).await?;

    match rx.await?? {
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

async fn handle_registration_confirm(content: &str, queue: Queue) -> Result<Message> {
    println!("Handle registration confirmation message");

    let (tx, rx) = oneshot::channel();
    let nested: NestedRequest<ConfirmRegistration> = serde_json::from_str(content)?;
    let req = Request::ConfirmRegistration(nested.request.message, tx);

    queue.send(req).await?;
    let response_data = rx.await??;

    Ok(NestedResponse::new_msg(
        RegistrationData {
            uuid: *response_data.uuid.as_bytes(),
            storage_capable: response_data.storage_capable,
        },
        CRAYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION,
    ))
}

async fn handle_sealed_sender_decrypt(content: &str, queue: Queue) -> Result<Message> {
    println!("Handle sealed sender decrypt message");

    let (tx, rx) = oneshot::channel();
    let nested: NestedRequest<sealedsender::SealedSenderMessage> = serde_json::from_str(content)?;
    let req = Request::DecryptSealedSender(nested.request.message, tx);

    queue.send(req).await?;
    let response_data = rx.await??;

    Ok(NestedResponse::new_msg(
        SealedSenderDecryptResponse {
            message: response_data.message,
        },
        CRAYFISH_WEBSOCKET_MESSAGE_SEALED_SENDER_DECRYPT,
    ))
}