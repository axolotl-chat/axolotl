use warp::Reply;
use warp::{Filter, Rejection};

mod ws;

type Result<T> = std::result::Result<T, Rejection>;

pub async fn start_websocket() {
    let port = 9080;
    let path = "libsignal";
    let ip = [127, 0, 0, 1];

    println!(
        "Starting libsignal-client-wrapper server at ws://{}.{}.{}.{}:{}/{}",
        ip[0], ip[1], ip[2], ip[3], port, path
    );

    let ws_route = warp::path(path).and(warp::ws()).and_then(ws_handler);
    warp::serve(ws_route).run((ip, port)).await;
}

pub async fn ws_handler(ws: warp::ws::Ws) -> Result<impl Reply> {
    println!("ws_handler");

    Ok(ws.on_upgrade(move |socket| ws::client_connection(socket)))
}
