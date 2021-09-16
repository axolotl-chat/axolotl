mod handlers;
mod types;

use futures::{SinkExt, StreamExt};
use std::convert::Infallible;
use tokio::sync::mpsc;
use warp::ws::WebSocket;
use warp::Reply;
use warp::{Filter, Rejection};

use crate::requests::handle_requests;
use crate::requests::RequestSender;
use types::ErrorMessage;

pub async fn start_websocket() {
    let port = 9081;
    let path = "libsignal";
    let ip = [127, 0, 0, 1];
    let request_queue_size = 10;

    let (tx, rx) = mpsc::channel(request_queue_size);
    println!(
        "Starting libsignal-service web socket at ws://{}.{}.{}.{}:{}/{}",
        ip[0], ip[1], ip[2], ip[3], port, path
    );

    let ws_route = warp::path(path)
        .and(warp::ws())
        .and(with_channel(tx))
        .and_then(ws_handler);
    let server_handle = tokio::task::spawn(warp::serve(ws_route).run((ip, port)));

    handle_requests(rx).await;
    server_handle.await.expect("Web Socket panicked.");
}

fn with_channel(
    tx: RequestSender,
) -> impl Filter<Extract = (RequestSender,), Error = Infallible> + Clone {
    warp::any().map(move || tx.clone())
}

async fn ws_handler(ws: warp::ws::Ws, tx: RequestSender) -> Result<impl Reply, Rejection> {
    Ok(ws.on_upgrade(move |socket| client_connection(socket, tx)))
}

pub async fn client_connection(mut ws: WebSocket, tx: RequestSender) {
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

        let response = match handlers::handle_message(msg, &tx).await {
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
