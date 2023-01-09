use crate::messages::send_message;
use crate::requests::{LinkDeviceRequest, SendMessageRequest};
use futures::StreamExt;
use futures::stream::SplitSink;
use presage::SledStore;
use presage::prelude::Uuid;
use serde_json::error::Error as SerdeError;
use std::str::FromStr;
use warp::ws::WebSocket;
use warp::filters::ws::Message;


/// Handles a client connection
pub async fn handle_ws_client(websocket: warp::ws::WebSocket) {
    let (mut sender, mut receiver) = websocket.split();
    
    // Read config
    let config_store = SledStore::open_with_passphrase("/tmp/presage-test/db", None::<&str>, presage::MigrationConflictStrategy::BackupAndDrop)
        .expect("Unable to open the database");
    
    // While messages come, handle them
    while let Some(body) = receiver.next().await {
        let message = match body {
            Ok(msg) => msg,
            Err(e) => {
                continue;
            }
        };

        handle_websocket_message(message, &mut sender, config_store.clone()).await;
    }
}

/// Handles a websocket message
async fn handle_websocket_message(message: Message, sender: &mut SplitSink<WebSocket, Message>, config_store: SledStore) {
    // Skip any non-Text messages...
    let msg = if let Ok(s) = message.to_str() {
        s
    } else {
        "Invalid message"
    };

    // Check the type of request
    if let Ok::<SendMessageRequest, SerdeError>(send_message_request) = serde_json::from_str(msg) {
        // Send a message
        let uuid = Uuid::from_str(&send_message_request.recipient).unwrap();
        send_message(&send_message_request.content, uuid, config_store).await;
    } else if let Ok::<LinkDeviceRequest, SerdeError>(link_device_request) = serde_json::from_str(msg) {
        // Link a device
    } else {
        // Error or unhandled request
    }


    //sender.send(Message::text("working")).await.unwrap();
}