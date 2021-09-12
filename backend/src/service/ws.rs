use crate::requests::Result;
use futures::SinkExt;
use futures::StreamExt;
use serde::Deserialize;
use serde::Serialize;
use warp::ws::Message;
use warp::ws::WebSocket;

pub async fn client_connection(mut ws: WebSocket) {
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

        let response = match handle_message(msg).await {
            Ok(response) => response,
            Err(e) => {
                println!("error sending message back: {}", e);
                Message::text("Invalid message")
            }
        };

        let result = ws.send(response).await;
        if let Err(e) = result {
            println!("error sending message back: {}", e);
        }
    }

    println!("client disconnected");
}

async fn handle_message(msg: Message) -> Result<Message> {
    let req: Request = serde_json::from_str(
        msg.to_str()
            .map_err(|_| "Message is not test. Failed to parse.".to_string())?,
    )
    .map_err(|e| e.to_string())?;

    println!("Received message: {:#?}", req);

    Ok(Message::text(
        serde_json::to_string(&Response {
            id: req.id,
            msg: "I received your Message :)".to_string(),
        })
        .map_err(|_| "Response could not be parsed")?,
    ))
}

#[derive(Deserialize, Debug)]
struct Request {
    id: u32,
    msg: String,
}

#[derive(Serialize)]
struct Response {
    id: u32,
    msg: String,
}
