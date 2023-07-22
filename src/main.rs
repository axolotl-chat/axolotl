use axolotl::handlers::get_app_dir;
use std::{process::exit, sync::Arc};

use axolotl::handlers::Handler;
use tokio::sync::Mutex;
use warp::{Filter, Rejection, Reply};

use clap::Parser;

#[derive(clap::ValueEnum, Clone, PartialEq)]
enum Mode {
    Daemon,
    #[cfg(feature = "ut")]
    UbuntuTouch,
    #[cfg(feature = "tauri")]
    Tauri,
}

#[derive(Parser)]
#[clap(about = "a basic signal CLI to try things out")]
struct Args {
    #[clap(long = "mode", short = 'm')]
    mode: Mode,
}

#[tokio::main]
async fn main() {
    env_logger::Builder::from_default_env()
        .parse_env("debug")
        .init();
    let args = Args::parse();

    let server_task = tokio::spawn(async {
        let handler = Handler::new().await.unwrap_or_else(|e| {
            log::error!("Error while starting the server: {}", e);
            exit(1);
        });
        log::info!("Axolotl handler started");
        start_websocket(Arc::new(Mutex::new(handler))).await;
    });

    start_ui(args.mode).await;

    server_task.await.unwrap();
}

async fn start_websocket(handler: Arc<Mutex<Handler>>) {
    log::info!("Starting the websocket server");

    let axolotl_ws_route = warp::path("ws")
        .and(warp::ws())
        .and(warp::any().map(move || handler.clone()))
        .and_then(handle_ws_client);

    // Just serve the attachments/ directory
    let axolotl_http_attachments_route = warp::path("attachments").and(warp::fs::dir(format!(
        "{}/{}",
        get_app_dir(),
        "attachments"
    )));

    warp::serve(axolotl_ws_route.or(axolotl_http_attachments_route))
        .run(([127, 0, 0, 1], 9080))
        .await;
    log::info!("Server stopped");
}

pub async fn handle_ws_client(
    websocket: warp::ws::Ws,
    handler: Arc<Mutex<Handler>>,
) -> Result<impl Reply, Rejection> {
    Ok(websocket.on_upgrade(move |websocket| async move {
        log::debug!("New websocket connection");
        handler.lock().await.handle_ws_client(websocket).await;
        log::debug!("websocket connection was closed");
    }))
}

async fn start_ui(mode: Mode) {
    match mode {
        #[cfg(feature = "tauri")]
        Mode::Tauri => {
            log::info!("Starting the tauri client");
            tauri::start_tauri().await;
        }
        #[cfg(feature = "ut")]
        Mode::UbuntuTouch => {
            log::info!("Starting the Ubuntu Touch client");
            ut::start_ut().await;
        }
        Mode::Daemon => {
            log::info!("Running headless");
        }
    }
}

#[cfg(feature = "tauri")]
mod tauri {
    const INIT_SCRIPT: &str = r#"
    document.addEventListener('DOMContentLoaded', function () {
        console.log("DOMContentLoaded");
        window.renderCallback = function (scheme, sitekey, action, token) {
        
            var targetURL = "tauri://localhost/?token=" + [scheme, sitekey, action, token].join(".");
            var link = document.createElement("a");
            link.href = targetURL;
            link.innerText = "open axolotl";
        
            document.body.removeAttribute("class");
            setTimeout(function () {
            document.getElementById("container").appendChild(link);
            }, 2000);
        
            window.location.href = targetURL;
        };
        window.intercept = function() {
            console.log("intercept")
            console.log("resetting captcha")
            document.getElementById("captcha").innerHTML = "";
            if(useHcaptcha)onloadHcaptcha();
            else onload();
          }
        if (!window.location.href.includes("localhost")){
            intercept();
        } else {
            console.log("localhost detected, not intercepting");
        }
    });
"#;

    pub async fn start_tauri() {
        let t = tauri::Builder::default().setup(|app| {
            tauri::WindowBuilder::new(app, "label", tauri::WindowUrl::App("index.html".into()))
                .initialization_script(INIT_SCRIPT)
                .title("Axolotl")
                .build()
                .unwrap();
            Ok(())
        });
        t.run(tauri::generate_context!())
            .expect("error while running tauri application");
    }
}

#[cfg(feature = "ut")]
mod ut {
    use super::*;
    use std::process::Command;
    use std::process::Stdio;

    pub async fn start_ut() {
        log::info!("Starting the ubuntu touch client");
        let _warp = tokio::spawn(async {
            let route = warp::fs::dir("./axolotl-web/dist");
            warp::serve(route).run(([127, 0, 0, 1], 9081)).await;
        });
        Command::new("qmlscene")
            .arg("--scaling")
            .arg("--webEngineArgs ")
            .arg("--remote-debugging-port")
            .arg("ut/MainUt.qml")
            .stdout(Stdio::piped())
            .spawn()
            .expect("GUI failed to start")
            .wait()
            .unwrap();

        exit(0);
    }
}