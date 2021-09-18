mod error;
mod requests;
mod service;

#[tokio::main]
async fn main() {
    service::start_websocket().await;
}
