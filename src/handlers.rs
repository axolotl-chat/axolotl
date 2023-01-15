use crate::manager_thread::ManagerThread;
use crate::messages::send_message;
use crate::requests::{LinkDeviceRequest, SendMessageRequest};
use futures::stream::SplitSink;
use futures::StreamExt;
use presage::prelude::Uuid;
use serde_json::error::Error as SerdeError;
use std::str::FromStr;
use warp::filters::ws::Message;
use warp::ws::WebSocket;

const MESSAGE_BOUND: usize = 10;
use futures::channel::oneshot;
use std::process::exit;

use tokio::sync::mpsc;
use presage::SledStore;
use presage::MigrationConflictStrategy;

/// Handles a client connection
pub async fn handle_ws_client(websocket: warp::ws::WebSocket) {
    println!("Setting up the manager");
    let (provisioning_link_tx, provisioning_link_rx) = oneshot::channel();
    let (error_tx, error_rx) = oneshot::channel();

    let (send_content, mut receive_content) = mpsc::unbounded_channel();
    let (send_error, mut receive_error) = mpsc::channel(MESSAGE_BOUND);
    let db_path = "/home/nanu/.config/presage";
    println!("Opening the database at {}", db_path);
    let config_store = SledStore::open_with_passphrase(
        db_path,
        None::<&str>,
        MigrationConflictStrategy::BackupAndDrop,
    )
    .ok()
    .unwrap();

    let manager = ManagerThread::new(
        config_store,
        "presage".to_string(),
        provisioning_link_tx,
        error_tx,
        send_content,
        send_error,
    )
    .await;
    if manager.is_none() {
        if let Some(error_opt) = receive_error.recv().await {
            println!("Got error after linking device: {}", error_opt);
            exit(0);
        }
    }
    println!("Awaiting for error linking");
    match error_rx.await {
        Ok(err) => {
            println!("Got error linking device: {}", err);
            exit(0);
        }
        Err(_e) => println!("Manager setup successfull"),
    }
    let m = manager.unwrap();
    // if manager.is_none() {
    //     if let Some(error_opt) = receive_error.recv().await {
    //         println!("Got error after linking device: {}", error_opt);
    //         return Err(error_opt);
    //     }
    // }
    let (mut sender, mut receiver) = websocket.split();

    // While messages come, handle them
    while let Some(body) = receiver.next().await {
        let message = match body {
            Ok(msg) => msg,
            Err(e) => {
                continue;
            }
        };

        handle_websocket_message(message, &mut sender, &m).await;
    }
}

/// Handles a websocket message
async fn handle_websocket_message(
    message: Message,
    sender: &mut SplitSink<WebSocket, Message>,
    manager: &ManagerThread,
) {
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
        send_message(send_message_request.content, uuid, manager).await;
    } else if let Ok::<LinkDeviceRequest, SerdeError>(link_device_request) =
        serde_json::from_str(msg)
    {
        // Link a device
    } else {
        // Error or unhandled request
        println!("Unhandled request {}", msg);
    }

    //sender.send(Message::text("working")).await.unwrap();
}
