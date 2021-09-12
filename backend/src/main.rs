mod requests;
mod service;

use crate::service::start_websocket;

#[actix_rt::main]
async fn main() {
    start_websocket().await;
}
