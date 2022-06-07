package main

import (
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"sync"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/ui"
	"github.com/nanu-c/axolotl/app/webserver"
	"github.com/nanu-c/axolotl/app/worker"
	"github.com/signal-golang/textsecure/crayfish"
)

func setup() *app.App {
	a := &app.App{}

	// Flags parsed in app/config/config.go
	a.Config = config.SetupConfig()

	log.Infoln("[axolotl] Starting axolotl version", config.AppVersion)
	return a
}
func runBackend(a *app.App) {
	go worker.RunBackend()
	if a.Config.IsPushHelper {
		push.PushHelperProcess()
	}
}
func runUI(a *app.App) {
	defer wg.Done()
	if a.Config.Gui != "ut" && a.Config.Gui != "lorca" && a.Config.Gui != "qt" {
		ui.RunUi(a.Config)
		runElectron(a)
	} else {
		ui.RunUi(a.Config)
	}
	endAxolotlGracefully()
}
func endAxolotlGracefully() {
	log.Infoln("[axolotl] ending axolotl")
	err := crayfish.Instance.Stop()
	if err != nil {
		log.Errorf("[axolotl] error stopping crayfish: %s", err)
	}
	os.Exit(0)
}
func runElectron(a *app.App) {
	defer endAxolotlGracefully()
	l := log.New()
	electronPath := os.Getenv("XDG_DATA_HOME")
	if len(electronPath) == 0 {
		electronPath = a.Config.ConfigDir + "/electron"
	}

	electronSwitches := []string{"--disable-dev-shm-usage", "--no-sandbox"}
	if a.Config.ElectronFlag == "no-ozone" {
		electronSwitches = append(electronSwitches, "")
	} else {
		electronSwitches = append(electronSwitches, "--ozone-platform-hint=auto")
	}
	log.Infoln("[axolotl-electron] starting astilelectron with the following switches:", electronSwitches)

	var astilElectronOptions = astilectron.Options{
		AppName:            "axolotl",
		AppIconDefaultPath: "axolotl-web/public/axolotl.png", // If path is relative, it must be relative to the data directory
		AppIconDarwinPath:  "axolotl-web/public/axolotl.png", // Same here
		BaseDirectoryPath:  electronPath,
		VersionElectron:    "18.0.1",
		VersionAstilectron: "0.51.0",
		SingleInstance:     true,
		ElectronSwitches:   electronSwitches,
	}

	var ae *astilectron.Astilectron
	var err error

	if os.Getenv("AXOLOTL_ELECTRON_BUNDLED") == "true" {
		err = bootstrap.Run(bootstrap.Options{
			AstilectronOptions: astilElectronOptions,
			Logger:             l,
			OnWait: func(astielectron *astilectron.Astilectron, _ []*astilectron.Window, _ *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
				ae = astielectron
				return nil
			},
		})
	} else {
		ae, err = astilectron.New(l, astilElectronOptions)
	}

	if err != nil {
		log.Errorln(errors.Wrap(err, "[axolotl-electron]: creating astilectron failed"))
	}

	defer ae.Close()

	// Start astilectron
	ae.HandleSignals()

	if err = ae.Start(); err != nil {
		log.Errorln(errors.Wrap(err, "[axolotl-electron] main: starting astilectron failed"))
	}

	ae.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
		log.Errorln("[axolotl-electron] Electron App has crashed", e)
		return
	})
	ae.HandleSignals()
	// New window
	var w *astilectron.Window
	title := "Axolotl"
	center := true
	height := 800
	width := 600
	if w, err = ae.NewWindow("http://"+a.Config.ServerHost+":"+a.Config.ServerPort, &astilectron.WindowOptions{
		Title:  &title,
		Center: &center,
		Height: &height,
		Width:  &width,
	}); err != nil {
		log.Debugln("[axolotl-electron]", errors.Wrap(err, "main: new window failed"))
	}
	w.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
		log.Errorln("[axolotl-electron] Electron App has crashed")
		return
	})
	w.On(astilectron.EventNameAppClose, func(e astilectron.Event) (deleteListener bool) {
		log.Errorln("[axolotl-electron] Electron App was closed")
		return
	})
	w.On(astilectron.EventNameWindowEventDidFinishLoad, func(e astilectron.Event) (deleteListener bool) {
		log.Infoln("[axolotl-electron] Page loaded")
		return
	})
	w.On(astilectron.EventNameWindowEventWillNavigate, func(e astilectron.Event) (deleteListener bool) {
		log.Infoln("[axolotl-electron] Electron navigation", e.URL)
		if strings.Contains(e.URL, "signalcaptchas.org") {
			log.Infoln("[axolotl-electron] overriding onload", e.URL)

			w.ExecuteJavaScript(
				`
				// override the default onload function

				window.onload=function() {
					var action = "registration";
					var isDone = false;
					var sitekey = "6LfBXs0bAAAAAAjkDyyI1Lk5gBAUWfhI_bIyox5W";

					var widgetId = grecaptcha.enterprise.render("container", {
					sitekey: sitekey,
					size: "checkbox",
					callback: function (token) {
						isDone = true;
						document.body.removeAttribute("class");
						window.location = ["http://` + a.Config.ServerHost + `:` + a.Config.ServerPort + `/?token=signal-recaptcha-v2", sitekey, action, token].join(".");
					},
					});
				}
				// cleanup
				var bodyTag = document.getElementsByTagName('body')[0];
				bodyTag.innerHTML ='<div id="container"></div>'
				grecaptcha  = undefined

				// reload recaptcha
				var script = document.createElement('script');
				script.type = 'text/javascript';
				script.src = "https://www.google.com/recaptcha/enterprise.js?onload=onload&render=explicit";
				bodyTag.appendChild(script);
				`)
		}
		return
	})
	w.On(astilectron.EventNameWindowEventDidGetRedirectRequest, func(e astilectron.Event) (deleteListener bool) {
		log.Infoln("[axolotl-electron] Electron redirect request ", e.URLNew)
		return
	})
	w.On(astilectron.EventNameWindowEventWebContentsExecutedJavaScript, func(e astilectron.Event) (deleteListener bool) {
		log.Infoln("[axolotl-electron] Electron navigation js")
		return
	})
	// Create windows
	if err = w.Create(); err != nil {
		log.Errorln("[axolotl-electron]", errors.Wrap(err, "main: creating window failed"))
	}
	log.Debugln("[axolotl-electron] open dev tools", a.Config.ElectronDebug)

	if a.Config.ElectronDebug {
		w.OpenDevTools()
	}
	w.Session.ClearCache()
	// Blocking pattern
	ae.Wait()
}
func runWebserver(a *app.App) {
	defer endAxolotlGracefully()
	log.Printf("[axolotl] Axolotl server started")
	// Fetch the URL.
	webserver.Run(a)
}

var wg sync.WaitGroup

func main() {
	app := setup()
	wg.Add(1)
	go runBackend(app)
	log.Println("[axolotl] Setup completed")
	wg.Add(1)
	go runWebserver(app)
	if app.Config.Gui != "server" {
		wg.Add(1)
		go runUI(app)
	} else {
		log.Printf("[axolotl] Axolotl frontend is at http://" + app.Config.ServerHost + ":" + app.Config.ServerPort + "/")
	}

	wg.Wait()
	log.Println("[axolotl] Axolotl stopped")
}
