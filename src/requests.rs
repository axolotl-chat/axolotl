//! This module lists the request structures.

use serde::Deserialize;

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