use crate::requests::Request;
use crate::requests::Result;
use crate::service::RequestSender;
use futures::SinkExt;
use futures::StreamExt;
use libsignal_service::provisioning::VerificationCodeResponse;
use serde::Serialize;
use tokio::sync::mpsc;
use tokio::sync::oneshot;
use warp::ws::Message;
use warp::ws::WebSocket;

pub async fn client_connection(mut ws: WebSocket, tx: RequestSender) {
    println!("establishing client connection... {:?}", ws);

    while let Some(result) = ws.next().await {
        let msg = match result {
            Ok(msg) => msg,
            Err(e) => {
                println!("error receiving message: {}", e);
                break;
            }
        };

        if msg.is_close() {
            break;
        }

        let response = match handle_message(msg, &tx).await {
            Ok(response) => response,
            Err(e) => {
                println!("error sending message back: {}", e);
                Message::text(
                    serde_json::to_string(&Error { error: e })
                        .unwrap_or("{\"error\": \"Failed to generate response\"}".to_string()),
                )
            }
        };

        let result = ws.send(response).await;
        if let Err(e) = result {
            println!("error sending message back: {}", e);
        }
    }

    println!("client disconnected");
}

async fn handle_message(msg: Message, tx: &mpsc::Sender<Request>) -> Result<Message> {
    let content: &str = msg
        .to_str()
        .map_err(|_| "Message is not text. Failed to parse.".to_string())?;

    let register = serde_json::from_str(content);
    let confirm = serde_json::from_str(content);

    if let Ok(data) = register {
        let (cb_tx, cb_rx) = oneshot::channel();
        tx.send(Request::Register(data, cb_tx))
            .await
            .map_err(|e| e.to_string())?;

        match cb_rx
            .await
            .map_err(|e| e.to_string())?
            .map_err(|e| e.to_string())?
        {
            VerificationCodeResponse::Issued => {
                println!("Registration request was sent");
                return Ok(Message::text(
                    serde_json::to_string(&Success { success: true }).map_err(|e| e.to_string())?,
                ));
            }
            VerificationCodeResponse::CaptchaRequired => {
                println!("Server requires a Captcha. Please generate one and try again!");
                return Ok(Message::text(
                    serde_json::to_string(&Error {
                        error: "Captcha required".to_string(),
                    })
                    .map_err(|e| e.to_string())?,
                ));
            }
        }
    };

    if let Ok(data) = confirm {
        let (cb_tx, cb_rx) = oneshot::channel();
        tx.send(Request::ConfirmRegistration(data, cb_tx))
            .await
            .map_err(|e| e.to_string())?;

        let response_data = cb_rx
            .await
            .map_err(|e| e.to_string())?
            .map_err(|e| e.to_string())?;

        return Ok(Message::text(
            serde_json::to_string(&RegistrationData {
                uuid: response_data.uuid.as_u128(),
                storage_capable: response_data.storage_capable,
            })
            .map_err(|e| e.to_string())?,
        ));
    };

    Ok(Message::text(
        serde_json::to_string(&Error {
            error: format!(
                "The message could not be parsed as any known type: {:?}, {:?}",
                register, confirm
            ),
        })
        .map_err(|_| "Response could not be parsed")?,
    ))
}

#[derive(Serialize)]
struct Error {
    error: String,
}

#[derive(Serialize)]
struct Success {
    success: bool,
}

#[derive(Serialize)]
struct RegistrationData {
    uuid: u128,
    storage_capable: bool,
}
