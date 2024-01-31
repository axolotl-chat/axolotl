//! The messages module

use crate::error::ApplicationError;
use crate::manager_thread::ManagerThread;
use crate::requests::{AxolotlMessage, AxolotlResponse, SendMessageResponse};
use presage::prelude::*;
use presage::proto::{DataMessage, GroupContextV2};
use presage::{Manager, Thread};
use presage_store_sled::SledStore;
use std::time::UNIX_EPOCH;

/**
 * Send a message to one people or a group.
 *
 * - recipient is a String containing the UUID of the recipient. A contact or a
 *   group, both are supported.
 * - msg is an optional String containing the text body of the message. Most messages
 *   would have it.
 * - attachments is an optional Vec of AttachmentPointer. The attachments must be
 *   already uploaded, here they are only sent.
 * - manager is the instance of ManagerThread.
 * - response_type is a string slice containing the Axolotl response type. This
 *   parameter is mandatory because the method is used to send message but also to
 *   send attachments. Could be removed in the future if both handlers are merged.
 */
pub async fn send_message(
    recipient: Thread,
    msg: Option<String>,
    attachments: Option<Vec<AttachmentPointer>>,
    manager: &ManagerThread,
    response_type: &str,
) -> Result<AxolotlResponse, ApplicationError> {
    log::info!("Sending a message.");
    let timestamp = std::time::SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards")
        .as_millis() as u64;

    // Add the attachments to the message, if any
    let mut attachments_vec = Vec::new();
    if let Some(a) = attachments {
        attachments_vec = a;
    }

    log::debug!("Sending {} attachments", attachments_vec.len());
    let data_message = DataMessage {
        body: msg,
        timestamp: Some(timestamp),
        attachments: attachments_vec,
        ..Default::default()
    };

    // Search the recipient's UUID. A contact or a group
    let result = match recipient {
        Thread::Contact(uuid) => {
            log::debug!("Sending a message to a contact.");
            manager
                .send_message(uuid, data_message.clone(), timestamp)
                .await
        }
        Thread::Group(uuid) => {
            log::debug!("Sending a message to a group.");
            manager
                .send_message_to_group(uuid, data_message.clone(), timestamp)
                .await
        }
    };
    let is_failed = result.is_err();
    if is_failed {
        log::error!(
            "send_message: Error while sending the message. {:?}",
            result.err()
        );
    }
    let mut message = AxolotlMessage::from_data_message(data_message);
    message.thread_id = Some(recipient);
    // message.sender = Some(manager.uuid());
    let response_data = SendMessageResponse { message, is_failed };
    let response_data_json = serde_json::to_string(&response_data)?;
    let response = AxolotlResponse {
        response_type: response_type.to_string(),
        data: response_data_json,
    };
    Ok(response)
}

pub async fn send_message_to_group(
    msg: &str,
    master_key_str: &str,
    attachments: Option<Vec<AttachmentPointer>>,
    config_store: SledStore,
) {
    let mut manager = match Manager::load_registered(config_store).await {
        Ok(m) => m,
        Err(e) => {
            println!("Error while loading the manager: {:?}", e);
            return;
        }
    };
    // Send message
    let timestamp = std::time::SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards")
        .as_millis() as u64;

    // Add the attachments to the message, if any
    let mut attachments_vec = Vec::new();
    if let Some(a) = attachments {
        attachments_vec = a;
    }

    let master_key: [u8; 32] = hex::decode(master_key_str).unwrap().try_into().unwrap();

    let message = DataMessage {
        body: Some(msg.to_string()),
        timestamp: Some(timestamp),
        attachments: attachments_vec,
        group_v2: Some(GroupContextV2 {
            master_key: Some(master_key.to_vec()),
            revision: Some(0),
            ..Default::default()
        }),
        ..Default::default()
    };

    match manager.group(&master_key) {
        Ok(group) => match group {
            Some(_) => {
                manager
                    .send_message_to_group(&master_key, message, timestamp)
                    .await
                    .unwrap();
            }
            None => {
                println!("Group not found");
            }
        },
        Err(e) => {
            println!("Group not found: {:?}", e);
        }
    }
}
