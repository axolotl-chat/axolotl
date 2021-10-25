mod handlers;
mod types;

use futures::{SinkExt, StreamExt};
use warp::ws::WebSocket;
use warp::Reply;
use warp::{Filter, Rejection};

use types::ErrorMessage;

use crate::signal_service::Queue;

pub async fn start_websocket(queue: Queue) {
    let port = 9081;
    let path = "libsignal";
    let ip = [127, 0, 0, 1];

    println!(
        "Starting libsignal-service web socket at ws://{}.{}.{}.{}:{}/{}",
        ip[0], ip[1], ip[2], ip[3], port, path
    );

    let ws_route = warp::path(path)
        .and(warp::ws())
        .and(with_queue(queue))
        .and_then(ws_handler);
    warp::serve(ws_route).run((ip, port)).await;
}

async fn ws_handler(ws: warp::ws::Ws, queue: Queue) -> Result<impl Reply, Rejection> {
    Ok(ws.on_upgrade(move |socket| client_connection(socket, queue)))
}

pub fn with_queue(
    queue: Queue,
) -> impl Filter<Extract = (Queue,), Error = std::convert::Infallible> + Clone {
    warp::any().map(move || queue.clone())
}

pub async fn client_connection(mut ws: WebSocket, queue: Queue) {
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

        let response = match handlers::handle_message(msg, queue.clone()).await {
            Ok(response) => response,
            Err(e) => ErrorMessage::new_msg(e),
        };

        let result = ws.send(response).await;
        if let Err(e) = result {
            println!("error sending message back: {}", e);
        }

        if queue.is_closed() {
            println!("Backend has stopped. Shutting down web socket...");
            break;
        }
    }

    println!("client disconnected");
}
