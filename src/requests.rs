//! This module lists the request structures.

use libsignal_service::prelude::Uuid;
use serde::{Deserialize, Serialize};
use presage::{
    prelude::{
        content::{Content, ContentBody},
    }
};

#[derive(Deserialize, Debug)]
pub struct SendMessageRequest {
    // The text content
    pub content: String,
    // The uuid
    pub recipient: String,
    // TODO: manage quote and attachment
}

#[derive(Deserialize, Debug)]
pub struct LinkDeviceRequest {
    pub device_name: String,
}

#[derive(Deserialize, Debug)]
pub struct AxolotlRequest {
    pub request: String,
    pub data: Option<String>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct AxolotlResponse {
    pub response_type: String,
    pub data: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct GetMessagesRequest {
    pub id: String,
    pub last_id: Option<u64>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct AxolotlMessage {
    pub message_type: String,
    sender:Option<Uuid>,
    message:Option<String>,
    timestamp:Option<u64>,
    is_outgoing:bool

}
impl AxolotlMessage {
    pub fn from_message(message: Content) -> AxolotlMessage {

        // log::info!( "{:?}", message);
        let body = &message.body;
        let message_type = match body{
            ContentBody::DataMessage(_) => "DataMessage",
            ContentBody::SynchronizeMessage(_) => "SyncMessage",
            ContentBody::CallMessage(_) => "CallMessage",
            ContentBody::ReceiptMessage(_) => "ReceiptMessage",
            ContentBody::TypingMessage(_) => "TypingMessage",
            _ => "Unknown",
        }.to_string();
        let mut is_outgoing = match body{
            ContentBody::DataMessage(_) => false,
            ContentBody::SynchronizeMessage(_) => true,
            _ => false,
        };
        let data_message = match body{
            ContentBody::DataMessage(data) =>{ 
                data.body.clone()
            },
            ContentBody::SynchronizeMessage(data) => {
                is_outgoing = true;
                if data.sent.is_some() && data.sent.clone().unwrap().message.is_some(){
                    let m = data.sent.clone().unwrap().message.clone().unwrap();
                    m.body.clone()
                } else {
                    log::info!("{:?}", data);
                    Some("SyncMessage".to_string())
                }
            },
            _ => None
        };
        let sender = match message.metadata.sender.uuid{
            Some(uuid) => uuid,
            None => Uuid::nil()
        };
        let timestamp:u64 = message.metadata.timestamp;
        AxolotlMessage {
            sender:Some(sender),
            message_type,
            message:data_message,
            timestamp:Some(timestamp),
            is_outgoing
        }
    }
}