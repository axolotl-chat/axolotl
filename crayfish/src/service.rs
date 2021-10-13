mod handlers;
mod types;

use futures::{SinkExt, StreamExt};
use warp::ws::WebSocket;
use warp::Reply;
use warp::{Filter, Rejection};

use types::ErrorMessage;

pub async fn start_websocket() {
    let port = 9081;
    let path = "libsignal";
    let ip = [127, 0, 0, 1];

    println!(
        "Starting libsignal-service web socket at ws://{}.{}.{}.{}:{}/{}",
        ip[0], ip[1], ip[2], ip[3], port, path
    );

    let ws_route = warp::path(path).and(warp::ws()).and_then(ws_handler);
    warp::serve(ws_route).run((ip, port)).await;
}

async fn ws_handler(ws: warp::ws::Ws) -> Result<impl Reply, Rejection> {
    Ok(ws.on_upgrade(client_connection))
}

pub async fn client_connection(mut ws: WebSocket) {
    println!("establishing client connection... {:?}", ws);

    while let Some(result) = ws.next().await {
        let msg = match result {
            Ok(msg) => msg,
            Err(e) => {
                println!("error receiving message: {}", e);
                continue;
            }
        };

        if msg.is_close() {
            break;
        }

        if msg.is_ping() {
            continue;
        }

        let response = match handlers::handle_message(msg).await {
            Ok(response) => response,
            Err(e) => ErrorMessage::new_msg(e),
        };

        let result = ws.send(response).await;
        if let Err(e) = result {
            println!("error sending message back: {}", e);
        }
    }

    println!("client disconnected");
}
