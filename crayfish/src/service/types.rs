use crate::error::Error;
use serde::{Deserialize, Serialize};
use warp::ws::Message;

const _CRAYFISH_WEBSOCKET_TYPE_UNKNOWN: u32 = 0;
pub const CRAYFISH_WEBSOCKET_TYPE_REQUEST: u32 = 1;
pub const CRAYFISH_WEBSOCKET_TYPE_RESPONSE: u32 = 2;

pub const CRAYFISH_WEBSOCKET_MESSAGE_UNKNOWN: u32 = 0;
pub const CRAYFISH_WEBSOCKET_MESSAGE_REGISTRATION: u32 = 1;
pub const CRAYFISH_WEBSOCKET_MESSAGE_CONFIRM_REGISTRAION: u32 = 2;
pub const CRAYFISH_WEBSOCKET_MESSAGE_SEALED_SENDER_DECRYPT: u32 = 3;

pub type NestedRequest<T> = RequestMessage<RequestBody<T>>;
pub type NestedResponse<T> = ResponseMessage<ResponseBody<T>>;

#[derive(Deserialize)]
pub struct RequestMessage<T> {
    #[serde(rename = "type")]
    pub direction: u32,
    pub request: T,
}

#[derive(Deserialize)]
pub struct TypeCheck {
    #[serde(rename = "type")]
    pub message_type: u32,
}

#[derive(Deserialize)]
pub struct RequestBody<T> {
    pub message: T,
}

#[derive(Serialize)]
pub struct ResponseMessage<T> {
    #[serde(rename = "type")]
    direction: u32,
    response: T,
}

#[derive(Serialize)]
pub struct ResponseBody<T> {
    #[serde(rename = "type")]
    message_type: u32,
    message: T,
}

#[derive(Serialize)]
pub struct ErrorMessage {
    error: String,
}

#[derive(Serialize)]
pub struct Success {
    pub success: bool,
}

#[derive(Serialize)]
pub struct RegistrationData {
    pub uuid: [u8; 16],
    pub storage_capable: bool,
}
#[derive(Serialize)]
pub struct SealedSenderDecryptResponse {
    pub message: [u8; 32],
}

impl<T: Serialize> NestedResponse<T> {
    pub fn new_msg(message: T, message_type: u32) -> Message {
        Self::new(message, message_type).to_msg()
    }

    fn new(message: T, message_type: u32) -> Self {
        NestedResponse {
            direction: CRAYFISH_WEBSOCKET_TYPE_RESPONSE,
            response: ResponseBody {
                message_type,
                message,
            },
        }
    }

    fn to_msg(self) -> Message {
        Message::text(
            serde_json::to_string(&self)
                .unwrap_or(ErrorMessage::failure_response(self.response.message_type)),
        )
    }
}

impl ErrorMessage {
    pub fn from(error: Error) -> Self {
        Self { error: error.msg }
    }

    pub fn new_msg(error: Error) -> Message {
        Message::text(
            serde_json::to_string(&ErrorMessage::from(error))
                .unwrap_or(Self::failure_response(CRAYFISH_WEBSOCKET_MESSAGE_UNKNOWN)),
        )
    }

    fn failure_response(message_type: u32) -> String {
        let message = "{ \"error\": \"Failed to generate proper error response\" }";
        let response = format!("{{ \"type\": {}, \"message\": {} }}", message_type, message);
        let direction = CRAYFISH_WEBSOCKET_TYPE_RESPONSE;
        format!("{{ \"type\": {}, \"response\": {} }} ", direction, response)
    }
}
