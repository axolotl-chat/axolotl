use futures::SinkExt;
use futures::StreamExt;
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

        let result = ws.send(Message::text("Hello, dear client!")).await;
        if let Err(e) = result {
            println!("error sending message back: {}", e);
        }
    }

    println!("client disconnected");
}
