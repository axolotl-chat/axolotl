//! This module lists the request structures.

use libsignal_service::prelude::Uuid;
use presage::libsignal_service::prelude::AttachmentIdentifier;
use presage::{
    prelude::{
        content::{Content, ContentBody},
        DataMessage,
    },
    Thread,
};
use serde::{Deserialize, Serialize};

#[derive(Deserialize, Debug)]
pub struct SendMessageRequest {
    // The text content
    pub text: String,
    // The uuid
    pub recipient: String,
    // TODO: manage quote, attachment and reaction
}

#[derive(Deserialize, Debug)]
pub struct UploadAttachmentRequest {
    // The data URL containing the base64-encoded file
    pub attachment: String,
    // The uuid
    pub recipient: String,
}
// #[cfg(feature = "ut")]
#[derive(Deserialize, Debug)]
pub struct UploadAttachmentUtRequest {
    // The path to the file
    pub path: String,
    pub recipient: String,
    pub mimetype: String,
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
}

#[derive(Serialize, Deserialize, Debug)]
pub struct ProfileRequest {
    pub id: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct ChangeNotificationsForThreadRequest {
    pub thread: Thread,
    pub muted: bool,
    pub archived: bool,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct SendMessageResponse {
    pub message: AxolotlMessage,
    pub is_failed: bool,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct AxolotlConfig {
    pub uuid: Option<String>,
    pub e164: Option<String>,
    pub platform: Option<String>,
    pub feature: Option<String>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct AttachmentMessage {
    // The content type
    pub ctype: String,
    // The filename
    pub filename: String,
}

impl AttachmentMessage {
    pub fn new(ctype: &str, filename: &str) -> Self {
        // Use the first part of the MIME type
        // image/png becomes image
        let content_type = if ctype.contains('/') {
            ctype.split('/').collect::<Vec<&str>>()[0].to_string()
        } else {
            ctype.to_string()
        };

        AttachmentMessage {
            ctype: content_type,
            filename: filename.to_string(),
        }
    }
}

#[derive(Serialize, Deserialize, Debug)]
pub struct AxolotlMessage {
    pub message_type: String,
    pub sender: Option<Uuid>,
    message: Option<String>,
    timestamp: Option<u64>,
    is_outgoing: bool,
    pub thread_id: Option<Thread>,
    attachments: Vec<AttachmentMessage>,
    is_sent: bool,
}
impl AxolotlMessage {
    pub fn from_message(message: Content) -> AxolotlMessage {
        //log::info!( "{:?}", message);
        let body = &message.body;
        let message_type = match body {
            ContentBody::DataMessage(_) => "DataMessage",
            ContentBody::SynchronizeMessage(_) => "SyncMessage",
            ContentBody::CallMessage(_) => "CallMessage",
            ContentBody::ReceiptMessage(_) => "ReceiptMessage",
            ContentBody::TypingMessage(_) => "TypingMessage",
            ContentBody::NullMessage(_) => "NullMessage",
            ContentBody::StoryMessage(_) => "StoryMessage",
            ContentBody::PniSignatureMessage(_) => "PniSignatureMessage",
            _ => "Unknown",
        }
        .to_string();
        let mut is_outgoing = match body {
            ContentBody::DataMessage(_) => false, // todo mark own messages as outgoing
            ContentBody::SynchronizeMessage(_) => true,
            _ => false,
        };

        let mut attachments: Vec<AttachmentMessage> = Vec::new();
        if let ContentBody::DataMessage(data) = body {
            if !data.attachments.is_empty() {
                for attachment in &data.attachments {
                    if let Some(AttachmentIdentifier::CdnId(id)) = attachment.attachment_identifier
                    {
                        let content_type = attachment.content_type();
                        attachments.push(AttachmentMessage::new(content_type, &id.to_string()));
                    }
                    if let Some(AttachmentIdentifier::CdnKey(id)) =
                        attachment.attachment_identifier.clone()
                    {
                        let content_type = attachment.content_type();
                        attachments.push(AttachmentMessage::new(content_type, &id));
                    }
                }
            }
        };
        if let ContentBody::SynchronizeMessage(data) = body {
            if data.sent.is_some() && data.sent.clone().unwrap().message.is_some() {
                let m = data.sent.clone().unwrap().message.unwrap();
                if !m.attachments.is_empty() {
                    for attachment in &m.attachments {
                        if let Some(AttachmentIdentifier::CdnId(id)) =
                            attachment.attachment_identifier
                        {
                            let content_type = attachment.content_type();
                            attachments.push(AttachmentMessage::new(content_type, &id.to_string()));
                        }
                        if let Some(AttachmentIdentifier::CdnKey(id)) =
                            attachment.attachment_identifier.clone()
                        {
                            let content_type = attachment.content_type();
                            attachments.push(AttachmentMessage::new(content_type, &id));
                        }
                    }
                }
            }
        };

        let data_message = match body {
            ContentBody::DataMessage(data) => {
                if data.reaction.is_some() {
                    data.reaction.clone().unwrap().emoji
                } else {
                    data.body.clone()
                }
            }
            ContentBody::SynchronizeMessage(data) => {
                is_outgoing = true;
                if data.sent.is_some() && data.sent.clone().unwrap().message.is_some() {
                    let m = data.sent.clone().unwrap().message.unwrap();
                    m.body
                } else {
                    Some("SyncMessage".to_string())
                }
            }
            _ => None,
        };
        let sender = message.metadata.sender.uuid;
        let timestamp: u64 = message.metadata.timestamp;
        AxolotlMessage {
            sender: Some(sender),
            message_type,
            message: data_message,
            timestamp: Some(timestamp),
            attachments,
            is_outgoing,
            thread_id: None,
            is_sent: true, // TODO
        }
    }
    pub fn from_content_body(body: ContentBody) -> AxolotlMessage {
        let message_type = match body {
            ContentBody::DataMessage(_) => "DataMessage",
            ContentBody::SynchronizeMessage(_) => "SyncMessage",
            ContentBody::CallMessage(_) => "CallMessage",
            ContentBody::ReceiptMessage(_) => "ReceiptMessage",
            ContentBody::TypingMessage(_) => "TypingMessage",
            ContentBody::NullMessage(_) => "NullMessage",
            ContentBody::StoryMessage(_) => "StoryMessage",
            ContentBody::PniSignatureMessage(_) => "PniSignatureMessage",
            _ => "Unknown",
        }
        .to_string();
        let mut is_outgoing = match body {
            ContentBody::DataMessage(_) => false, // todo mark own messages as outgoing
            ContentBody::SynchronizeMessage(_) => true,
            _ => false,
        };

        let mut attachments: Vec<AttachmentMessage> = Vec::new();
        if let ContentBody::DataMessage(ref data) = body {
            if !data.attachments.is_empty() {
                for attachment in &data.attachments {
                    if let Some(AttachmentIdentifier::CdnId(id)) =
                        attachment.attachment_identifier.clone()
                    {
                        let content_type = attachment.content_type();
                        attachments.push(AttachmentMessage::new(content_type, &id.to_string()));
                    }
                    if let Some(AttachmentIdentifier::CdnKey(id)) =
                        attachment.attachment_identifier.clone()
                    {
                        let content_type = attachment.content_type();
                        attachments.push(AttachmentMessage::new(content_type, &id));
                    }
                }
            }
        };

        let data_message = match &body {
            ContentBody::DataMessage(data) => {
                if data.reaction.is_some() {
                    data.reaction.clone().unwrap().emoji
                } else if !data.attachments.is_empty() {
                    if data.body.is_some() {
                        Some(format!(
                            "Unsuported attachment. {}",
                            data.body.clone().unwrap()
                        ))
                    } else {
                        Some("Unsuported attachment.".to_string())
                    }
                } else {
                    data.body.clone()
                }
            }
            ContentBody::SynchronizeMessage(data) => {
                is_outgoing = true;
                if data.sent.is_some() && data.sent.clone().unwrap().message.is_some() {
                    let m = data.sent.clone().unwrap().message.unwrap();
                    if m.reaction.is_some() {
                        m.reaction.unwrap().emoji
                    } else if !m.attachments.is_empty() {
                        for attachment in &m.attachments {
                            if let Some(AttachmentIdentifier::CdnId(id)) =
                                attachment.attachment_identifier
                            {
                                let content_type = attachment.content_type();
                                attachments
                                    .push(AttachmentMessage::new(content_type, &id.to_string()));
                            }
                            if let Some(AttachmentIdentifier::CdnKey(id)) =
                                attachment.attachment_identifier.clone()
                            {
                                let content_type = attachment.content_type();
                                attachments.push(AttachmentMessage::new(content_type, &id));
                            }
                        }
                        if m.body.is_some() {
                            Some(format!("Unsuported attachment. {}", m.body.unwrap()))
                        } else {
                            Some("Unsuported attachment.".to_string())
                        }
                    } else {
                        m.body
                    }
                } else {
                    Some("SyncMessage".to_string())
                }
            }
            _ => None,
        };
        let timestamp: Option<u64> = match body {
            ContentBody::DataMessage(m) => m.timestamp,
            ContentBody::SynchronizeMessage(m) => match m.sent {
                Some(s) => s.timestamp,
                None => None,
            },
            _ => None,
        };
        AxolotlMessage {
            sender: None,
            message_type,
            message: data_message,
            timestamp,
            is_outgoing,
            thread_id: None,
            attachments,
            is_sent: true, // TODO
        }
    }
    pub fn from_data_message(data: DataMessage) -> AxolotlMessage {
        let message_type = "DataMessage".to_string();
        let is_outgoing = false;

        let mut attachments: Vec<AttachmentMessage> = Vec::new();
        if !data.attachments.is_empty() {
            for attachment in &data.attachments {
                if let Some(AttachmentIdentifier::CdnId(id)) =
                    attachment.attachment_identifier.clone()
                {
                    let content_type = attachment.content_type();
                    attachments.push(AttachmentMessage::new(content_type, &id.to_string()));
                }
                if let Some(AttachmentIdentifier::CdnKey(id)) =
                    attachment.attachment_identifier.clone()
                {
                    let content_type = attachment.content_type();
                    attachments.push(AttachmentMessage::new(content_type, &id));
                }
            }
        };

        let data_message = if data.reaction.is_some() {
            data.reaction.clone().unwrap().emoji
        } else if !data.attachments.is_empty() {
            if data.body.is_some() {
                Some(format!(
                    "Unsuported attachment. {}",
                    data.body.clone().unwrap()
                ))
            } else {
                Some("Unsuported attachment.".to_string())
            }
        } else {
            data.body.clone()
        };
        let timestamp: u64 = data.timestamp.unwrap();
        AxolotlMessage {
            sender: None,
            message_type,
            message: data_message,
            timestamp: Some(timestamp),
            is_outgoing,
            thread_id: None,
            attachments,
            is_sent: true, // TODO
        }
    }
}
