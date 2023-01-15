pub mod messages;
pub mod handlers;
pub mod requests;
pub mod error;

mod manager_thread;
use crate::handlers::handle_ws_client;
use warp::Filter;

#[tokio::main]
async fn main() {
    env_logger::Builder::from_default_env().parse_env("info").init();
    println!("Starting the server");

   let axolotl_route = warp::path("ws")
        .and(warp::ws())
        .map(|ws: warp::ws::Ws| {
            ws.on_upgrade(|socket| handle_ws_client(socket))
        });

    warp::serve(axolotl_route)
        .run(([127, 0, 0, 1], 9080)).await;
}

