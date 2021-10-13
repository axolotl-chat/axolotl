mod error;
mod requests;
mod service;
mod store;
mod config;

#[tokio::main]
async fn main() {
    service::start_websocket().await;
}
