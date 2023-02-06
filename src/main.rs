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

    let axolotl_route = warp::path("ws")
        .and(warp::ws())
        .and(warp::any().map(move || handler.clone()))
        .and_then(handle_ws_client);

    warp::serve(axolotl_route).run(([127, 0, 0, 1], 9080)).await;
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
        alert("init");
        window.renderCallback = function (scheme, sitekey, action, token) {
        
            var targetURL = "tauri://localhost/?token=" + [scheme, sitekey, action, token].join(".");
            var link = document.createElement("a");
            link.href = targetURL;
            link.innerText = "open axolotl";
        
            document.body.removeAttribute("class");
            alert(targetURL);
            setTimeout(function () {
            document.getElementById("container").appendChild(link);
            }, 2000);
        
            window.location.href = targetURL;
        };
        function onload() {
            alert("onload");
            var action = document.location.href.indexOf("challenge") !== -1 ?
              "challenge" : "registration";
            var isDone = false;
            var sitekey = "6LfBXs0bAAAAAAjkDyyI1Lk5gBAUWfhI_bIyox5W";
          
            var widgetId = grecaptcha.enterprise.render("captcha", {
              sitekey: sitekey,
              size: "checkbox",
              theme: getTheme(),
              callback: function (token) {
                isDone = true;
                renderCallback("signal-recaptcha-v2", sitekey, action, token);
              },
            });
          
            function execute() {
              if (isDone) {
                return;
              }
          
              grecaptcha.enterprise.execute(widgetId, { action: action });
          
              // Below, we immediately reopen if the user clicks outside the widget. If they
              //   close it some other way (e.g., by pressing Escape), we force-reopen it
              //   every second.
              setTimeout(execute, 1000);
            }
          
            // If the user clicks outside the widget, reCAPTCHA will open it, but we'll
            //   immediately reopen it. (We use onclick for maximum browser compatibility.)
            document.body.onclick = function () {
              if (!isDone) {
                grecaptcha.enterprise.execute(widgetId, { action: action });
              }
            };
          
            execute();
          }
        onload();
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
