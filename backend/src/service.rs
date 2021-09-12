use std::convert::Infallible;
use tokio::sync::mpsc;
use warp::Reply;
use warp::{Filter, Rejection};

use crate::requests::handle_requests;
use crate::requests::RequestSender;

pub mod ws;

pub async fn start_websocket() {
    let port = 9080;
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
    Ok(ws.on_upgrade(move |socket| ws::client_connection(socket, tx)))
}
