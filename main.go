package main

import (
	"flag"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"sync"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/ui"
	"github.com/nanu-c/axolotl/app/webserver"
	"github.com/nanu-c/axolotl/app/worker"
	"github.com/signal-golang/textsecure/crayfish"
)

func init() {
	flag.StringVar(&config.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
	flag.StringVar(&config.Gui, "e", "", "Specify runtime environment. Use either electron, ut, lorca, qt or server")
	flag.StringVar(&config.AxolotlWebDir, "axolotlWebDir", "./axolotl-web/dist", "Specify the directory to use for axolotl-web")
	flag.BoolVar(&config.ElectronDebug, "eDebug", false, "Open electron development console")
	flag.BoolVar(&config.PrintVersion, "version", false, "Print version info")
	flag.StringVar(&config.ServerHost, "host", "127.0.0.1", "Host to serve UI from.")
	flag.StringVar(&config.ServerPort, "port", "9080", "Port to serve UI from.")
	flag.StringVar(&config.ElectronFlag, "electron-flag", "", "Specify electron flag. Use no-ozone to disable Ozone/Wayland platform")
}
func setup() {
	config.SetupConfig()
	log.Infoln("[axolotl] Starting axolotl version", config.AppVersion)
}
func runBackend() {
	errorChannel := make(chan error)
	go worker.RunBackend(errorChannel)
	if config.IsPushHelper {
		go push.PushHelperProcess()
	}
	err := <-errorChannel
	if err != nil {
		os.Exit(1)
	}
}
func runUI() error {
	defer wg.Done()
	if config.Gui != "ut" && config.Gui != "lorca" && config.Gui != "qt" {
		ui.RunUi(config.Gui)
		runElectron()
	} else {
		ui.RunUi(config.Gui)
	}
	endAxolotlGracefully()
	return nil
}
func endAxolotlGracefully() {
	log.Infoln("[axolotl] ending axolotl")
	err := crayfish.Instance.Stop()
	if err != nil {
		log.Errorf("[axolotl] error stopping crayfish: %s", err)
	}
	os.Exit(0)
}
func runElectron() {
	defer endAxolotlGracefully()
	l := log.New()
	electronPath := os.Getenv("XDG_DATA_HOME")
	if len(electronPath) == 0 {
		electronPath = config.ConfigDir + "/electron"
	}

	electronSwitches := []string{"--disable-dev-shm-usage", "--no-sandbox"}
	if config.ElectronFlag == "no-ozone" {
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
		VersionElectron:    "20.2.0",
		VersionAstilectron: "0.56.0",
		SingleInstance:     true,
		ElectronSwitches:   electronSwitches,
	}

	var a *astilectron.Astilectron
	var err error

	if os.Getenv("AXOLOTL_ELECTRON_BUNDLED") == "true" {
		err = bootstrap.Run(bootstrap.Options{
			AstilectronOptions: astilElectronOptions,
			Logger:             l,
			OnWait: func(astielectron *astilectron.Astilectron, _ []*astilectron.Window, _ *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
				a = astielectron
				return nil
			},
		})
	} else {
		a, err = astilectron.New(l, astilElectronOptions)
	}

	if err != nil {
		log.Errorln(errors.Wrap(err, "[axolotl-electron]: creating astilectron failed"))
	}

	defer a.Close()

	// Start astilectron
	a.HandleSignals()

	if err = a.Start(); err != nil {
		log.Errorln(errors.Wrap(err, "[axolotl-electron] main: starting astilectron failed"))
	}

	a.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
		log.Errorln("[axolotl-electron] Electron App has crashed", e)
		return
	})
	a.HandleSignals()
	// New window
	var w *astilectron.Window
	title := "Axolotl"
	center := true
	height := 800
	width := 600
	if w, err = a.NewWindow("http://"+config.ServerHost+":"+config.ServerPort, &astilectron.WindowOptions{
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
						window.location = ["http://` + config.ServerHost + `:` + config.ServerPort + `/?token=signal-recaptcha-v2", sitekey, action, token].join(".");
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
	log.Debugln("[axolotl-electron] open dev tools", config.ElectronDebug)

	if config.ElectronDebug {
		w.OpenDevTools()
	}
	w.Session.ClearCache()
	// Blocking pattern
	a.Wait()
}
func runWebserver() {
	defer endAxolotlGracefully()
	log.Printf("[axolotl] Axolotl server started")
	// Fetch the URL.
	webserver.Run()
}

var wg sync.WaitGroup

func main() {
	setup()
	wg.Add(1)
	go runBackend()
	log.Println("[axolotl] Setup completed")
	wg.Add(1)
	go runWebserver()
	if config.Gui != "server" {
		wg.Add(1)
		go runUI()
	} else {
		log.Printf("[axolotl] Axolotl frontend is at http://" + config.ServerHost + ":" + config.ServerPort + "/")
	}

	wg.Wait()
	log.Println("[axolotl] Axolotl stopped")
}
