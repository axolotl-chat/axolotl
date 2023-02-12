//! The messages module

use std::time::UNIX_EPOCH;
use presage::prelude::*;
use presage::{Manager, SledStore};

use crate::manager_thread::ManagerThread;

/**
 * Send a message to one people.
 * 
 * Currently it only sends text message. TODO: make it more abstract to send pictures and so on. 
 */
pub async fn send_message(msg: String, uuid: Uuid, manager: &ManagerThread)
{
    // Send message
    let timestamp = std::time::SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards")
        .as_millis() as u64;
    let message = ContentBody::DataMessage(DataMessage {
        body: Some(msg.to_string()),
        timestamp: Some(timestamp),
        ..Default::default()
    });
    manager.send_message(uuid, message, timestamp).await.unwrap();     
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

    match manager
        .group(&master_key)
        {
            Ok(group) => {
                match group{
                    Some(_) => {
                        manager
                            .send_message_to_group(
                                &master_key,
                                message,
                                timestamp
                            ).await.unwrap(); 
                    },
                    None => {
                        println!("Group not found");
                    }
                }
 
            },
            Err(e) => {
                println!("Group not found: {:?}", e);
            }
        }

   
}
