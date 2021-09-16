mod error;
mod requests;
mod service;

#[actix_rt::main]
async fn main() {
    service::start_websocket().await;
}
