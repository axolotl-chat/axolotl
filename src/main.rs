pub mod messages;
pub mod handlers;
pub mod requests;

use crate::handlers::handle_ws_client;
use warp::Filter;


#[tokio::main]
async fn main() {

    let axolotl_route = warp::path("axolotl")
        .and(warp::ws())
        .map(|ws: warp::ws::Ws| {
            ws.on_upgrade(handle_ws_client)
        });

    warp::serve(axolotl_route)
        .run(([127, 0, 0, 1], 9231)).await;
}

