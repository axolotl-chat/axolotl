use super::types::*;
use crate::error::{Error, Result};
use crate::requests::{registration, Request};

use libsignal_service::provisioning::VerificationCodeResponse;
use tokio::sync::{mpsc, oneshot};
use warp::ws::Message;

use super::types::ErrorMessage;

pub async fn handle_message(msg: Message, tx: &mpsc::Sender<Request>) -> Result<Message> {
    let content: &str = msg
        .to_str()
        .map_err(|_| Error::new("Message is not text. Failed to parse."))?;

    let type_check: RequestMessage<TypeCheck> = serde_json::from_str(content)?;

    if type_check.direction != CRAYFISH_WEBSOCKET_TYPE_REQUEST {
        Error::is("Received Message is no valid request")?
    }

    let response = match type_check.request.message_type {
        CRAYFISH_WEBSOCKET_MESSAGE_REGISTRATION => handle_registration(content, tx).await,
        CRAYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION => {
            handle_registration_confirm(content, tx).await
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

async fn handle_registration(content: &str, tx: &mpsc::Sender<Request>) -> Result<Message> {
    println!("Handle registration message");
    let req: NestedRequest<registration::Register> = serde_json::from_str(content)?;

    let (cb_tx, cb_rx) = oneshot::channel();
    tx.send(Request::Register(req.request.message, cb_tx))
        .await?;

    match cb_rx.await?? {
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

async fn handle_registration_confirm(content: &str, tx: &mpsc::Sender<Request>) -> Result<Message> {
    println!("Handle registration confirmation message");
    let req: NestedRequest<registration::ConfirmRegistration> = serde_json::from_str(content)?;

    let (cb_tx, cb_rx) = oneshot::channel();
    tx.send(Request::ConfirmRegistration(req.request.message, cb_tx))
        .await?;

    let response_data = cb_rx.await??;

    Ok(NestedResponse::new_msg(
        RegistrationData {
            uuid: *response_data.uuid.as_bytes(),
            storage_capable: response_data.storage_capable,
        },
        CRAYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION,
    ))
}
