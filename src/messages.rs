//! The messages module

use crate::error::ApplicationError;
use crate::manager_thread::ManagerThread;
use crate::requests::{AxolotlMessage, AxolotlResponse, SendMessageResponse};
use presage::prelude::proto::AttachmentPointer;
use presage::prelude::*;
use presage::{Manager, SledStore, Thread};
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
    recipient: String,
    msg: Option<String>,
    attachments: Option<Vec<AttachmentPointer>>,
    manager: &ManagerThread,
    response_type: &str
) -> Result<AxolotlResponse, ApplicationError>
{
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
    let thread = match Thread::try_from(&recipient) {
        Ok(t) => t,
        Err(e) => {
            log::error!("Error while parsing the request. {:?}", e);
            return Err(ApplicationError::InvalidRequest);
        }
    };

    let result = match thread {
        Thread::Contact(contact) => {
            log::debug!("Sending a message to a contact.");
            let message = ContentBody::DataMessage(data_message.clone());
            manager
                .send_message(contact, message.clone(), timestamp)
                .await
        }
        Thread::Group(group) => {
            log::debug!("Sending a message to a group.");
            let group_master_key = GroupMasterKey::new(group.clone());
            let group_from_store =
                manager.get_group_v2(group_master_key).await.ok().unwrap();
            let group_members = group_from_store.members.iter();
            let mut group_members_service_addresses: Vec<ServiceAddress> = Vec::new();

            for member in group_members {
                group_members_service_addresses.push(ServiceAddress {
                    uuid: Some(member.uuid.clone()),
                    phonenumber: None,
                    relay: None,
                });
            }
            let mut group_data_message = data_message.clone();
            group_data_message.group_v2 = Some(GroupContextV2 {
                master_key: Some(group.to_vec()),
                group_change: None,
                revision: Some(group_from_store.version),
            });
            manager
                .send_message_to_group(
                    group_members_service_addresses,
                    group_data_message,
                    timestamp,
                )
                .await
        }
    };

    let is_failed = result.is_err();
    if is_failed {
        log::error!("Error while sending the message. {:?}", result.err());
    }
    let mut message = AxolotlMessage::from_data_message(data_message);
    message.thread_id = Some(thread.to_string());
    message.sender = Some(manager.uuid());
    let response_data = SendMessageResponse { message, is_failed };
    let response_data_json = serde_json::to_string(&response_data).unwrap();
    let response = AxolotlResponse {
        response_type: response_type.to_string(),
        data: response_data_json,
    };
    Ok(response)
}

pub async fn send_message_to_group(msg: &str, master_key_str: &str, config_store: SledStore)
{
    let mut manager = Manager::load_registered(config_store).unwrap();
    // Send message
    let timestamp = std::time::SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards")
        .as_millis() as u64;
    
    // 158
    let master_key: [u8; 32] = hex::decode(master_key_str)
        .unwrap()
        .try_into()
        .unwrap();

    let message = DataMessage {
        body: Some(msg.to_string()),
        timestamp: Some(timestamp),
        group_v2: Some(GroupContextV2 {
            master_key: Some(master_key.to_vec()),
            revision: Some(0),
            ..Default::default()
        }),
        ..Default::default()
    };

    let group = manager
        .get_group_v2(GroupMasterKey::new(master_key))
        .await
        .unwrap();

    let me = manager.whoami().await.unwrap().uuid;

    manager
        .send_message_to_group(
            group.members.into_iter()
            .filter(|m| m.uuid != me)
            .map(|m| m.uuid).map(Into::into),
            message,
            timestamp
        ).await.unwrap();     
}
