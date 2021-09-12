use crate::requests::registration;
use crate::requests::Request;
use crate::requests::Result;
use crate::service::RequestSender;
use futures::SinkExt;
use futures::StreamExt;
use libsignal_service::provisioning::VerificationCodeResponse;
use serde::Deserialize;
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

    let type_check: RequestMessage<TypeCheck> =
        serde_json::from_str(content).map_err(|e| e.to_string())?;

    if type_check.type_t != CRYFISH_WEBSOCKET_TYPE_REQUEST {
        return Err("Received Message is no valid request".to_string());
    }

    match type_check.request.type_t {
        CRYFISH_WEBSOCKET_MESSAGE_REGISTRATION => handle_register(content, tx).await,
        CRYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION => handle_register_confirm(content, tx).await,
        _ => Err("Received request type is invalid".to_string()),
    }
}

async fn handle_register(content: &str, tx: &mpsc::Sender<Request>) -> Result<Message> {
    println!("Handle registration message");
    let req: NestedRequest<registration::Register> =
        serde_json::from_str(content).map_err(|e| e.to_string())?;

    let (cb_tx, cb_rx) = oneshot::channel();
    tx.send(Request::Register(req.request.message, cb_tx))
        .await
        .map_err(|e| e.to_string())?;

    match cb_rx
        .await
        .map_err(|e| e.to_string())?
        .map_err(|e| e.to_string())?
    {
        VerificationCodeResponse::Issued => {
            println!("Registration request was sent");
            Ok(Message::text(
                serde_json::to_string(&NestedResponse::new(
                    Success { success: true },
                    CRYFISH_WEBSOCKET_MESSAGE_REGISTRATION,
                ))
                .map_err(|e| e.to_string())?,
            ))
        }
        VerificationCodeResponse::CaptchaRequired => {
            println!("Server requires a Captcha. Please generate one and try again!");
            Ok(Message::text(
                serde_json::to_string(&NestedResponse::new(
                    Error {
                        error: "Captcha required".to_string(),
                    },
                    CRYFISH_WEBSOCKET_MESSAGE_REGISTRATION,
                ))
                .map_err(|e| e.to_string())?,
            ))
        }
    }
}

async fn handle_register_confirm(content: &str, tx: &mpsc::Sender<Request>) -> Result<Message> {
    println!("Handle registration confirmation message");
    let req: NestedRequest<registration::ConfirmRegistration> =
        serde_json::from_str(content).map_err(|e| e.to_string())?;

    let (cb_tx, cb_rx) = oneshot::channel();
    tx.send(Request::ConfirmRegistration(req.request.message, cb_tx))
        .await
        .map_err(|e| e.to_string())?;

    let response_data = cb_rx
        .await
        .map_err(|e| e.to_string())?
        .map_err(|e| e.to_string())?;

    Ok(Message::text(
        serde_json::to_string(&NestedResponse::new(
            RegistrationData {
                uuid: response_data.uuid.as_u128(),
                storage_capable: response_data.storage_capable,
            },
            CRYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION,
        ))
        .map_err(|e| e.to_string())?,
    ))
}

const CRYFISH_WEBSOCKET_TYPE_UNKNOWN: u32 = 0;
const CRYFISH_WEBSOCKET_TYPE_REQUEST: u32 = 1;
const CRYFISH_WEBSOCKET_TYPE_RESPONSE: u32 = 2;

const CRYFISH_WEBSOCKET_MESSAGE_UNKNOWN: u32 = 0;
const CRYFISH_WEBSOCKET_MESSAGE_REGISTRATION: u32 = 1;
const CRYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION: u32 = 2;

impl<T> NestedResponse<T> {
    fn new(message: T, type_t: u32) -> Self {
        NestedResponse {
            type_t: CRYFISH_WEBSOCKET_TYPE_RESPONSE,
            response: ResponseBody { type_t, message },
        }
    }
}

type NestedRequest<T> = RequestMessage<RequestBody<T>>;
type NestedResponse<T> = ResponseMessage<ResponseBody<T>>;

#[derive(Deserialize)]
struct RequestMessage<T> {
    #[serde(rename = "type")]
    type_t: u32,
    request: T,
}

#[derive(Deserialize)]
struct TypeCheck {
    #[serde(rename = "type")]
    type_t: u32,
}

#[derive(Deserialize)]
struct RequestBody<T> {
    #[serde(rename = "type")]
    type_t: u32,
    message: T,
}

#[derive(Serialize)]
struct ResponseMessage<T> {
    #[serde(rename = "type")]
    type_t: u32,
    response: T,
}

#[derive(Serialize)]
struct ResponseBody<T> {
    #[serde(rename = "type")]
    type_t: u32,
    message: T,
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
