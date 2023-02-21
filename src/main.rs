use axolotl::handlers::{get_app_dir, handle_ws_client};
use clap::Parser;
use std::collections::HashMap;
use std::thread;
use warp::Filter;

#[derive(Parser)]
#[clap(about = "a basic signal CLI to try things out")]
struct Args {
    #[clap(long = "deamon", short = 'd')]
    deamon: bool,
    #[clap(long = "ubuntu-touch", short = 'u')]
    ubuntu_touch: bool,
}

#[tokio::main]
async fn main() {
    env_logger::Builder::from_default_env()
        .parse_env("debug")
        .init();
    let args = Args::parse();
    thread::spawn(|| {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
            start_websocket().await;
        });
    });
    if args.deamon {
        log::info!("Starting the deamon");
        loop {
            std::thread::sleep(std::time::Duration::from_secs(1));
        }
    } else {
        log::info!("Starting the client");
        start_ui().await;
    }
}

async fn start_websocket() {
    log::info!("Starting the server");
    let axolotl_ws_route = warp::path("ws")
        .and(warp::ws())
        .map(|ws: warp::ws::Ws| ws.on_upgrade(|socket| handle_ws_client(socket)));

    // Just serve the attachments/ directory
    let axolotl_http_attachments_route = warp::path("attachments")
    .and(warp::fs::dir(format!("{}/{}", get_app_dir(), "attachments")));

    warp::serve(axolotl_ws_route.or(axolotl_http_attachments_route)).run(([127, 0, 0, 1], 9080)).await;
    log::info!("Server stopped");
}
async fn start_ui() {
    #[cfg(feature = "tauri")]
    start_tauri().await;
    #[cfg(feature = "ut")]
    start_ut().await;
    log::error!("No client found. Either use the tauri, the ubuntu touch client or the deamon.");
}
#[cfg(feature = "tauri")]
async fn start_tauri() {
    tauri::Builder::default()
        // .invoke_handler(tauri::generate_handler![start_websocket])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[cfg(feature = "ut")]
async fn start_ut() {
    use std::process::{exit, Stdio};
    use std::process::Command;


    log::info!("Starting the ubuntu touch client");
    thread::spawn( || {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
            let route = warp::fs::dir("./axolotl-web/dist");
            warp::serve(route).run(([127, 0, 0, 1], 9081)).await;
        });

    });
    Command::new("qmlscene")
        .arg("--scaling")
        .arg("--webEngineArgs ")
        .arg("--remote-debugging-port")
        .arg("ut/MainUt.qml")
        .stdout(Stdio::piped())
        .output()
        .expect("ls command failed to start");

    exit(0);
}
