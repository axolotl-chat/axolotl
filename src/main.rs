use std::{process::exit, thread};

use axolotl::handlers::handle_ws_client;
use presage::{MigrationConflictStrategy, SledStore};
use warp::Filter;

use clap::Parser;

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
    let axolotl_route = warp::path("ws")
        .and(warp::ws())
        .map(|ws: warp::ws::Ws| ws.on_upgrade(|socket| handle_ws_client(socket)));

    warp::serve(axolotl_route).run(([127, 0, 0, 1], 9080)).await;
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

#[cfg(feature = "tauri")]
async fn start_tauri() {
    let t = tauri::Builder::default().setup(|app| {
        tauri::WindowBuilder::new(app, "label", tauri::WindowUrl::App("index.html".into()))
            .initialization_script(INIT_SCRIPT)
            .title("Axolotl")
            .build();
        Ok(())
    });
    t.run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[cfg(feature = "ut")]
async fn start_ut() {
    use std::process::Command;
    use std::process::{exit, Stdio};

    log::info!("Starting the ubuntu touch client");
    thread::spawn(|| {
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
