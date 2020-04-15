package main

import (
	"bufio"
	"flag"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"sync"

	astikit "github.com/asticode/go-astikit"
	astilectron "github.com/asticode/go-astilectron"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/ui"
	"github.com/nanu-c/axolotl/app/webserver"
	"github.com/nanu-c/axolotl/app/worker"
)

var e string

func init() {
	flag.StringVar(&config.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
	flag.StringVar(&config.Gui, "e", "", "use either electron, ut, lorca or me")
	flag.BoolVar(&config.ElectronDebug, "eDebug", false, "use to show development console in electron")
}
func print(stdout io.ReadCloser) {
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		m := scanner.Text()
		log.Println("[axolotl] ", m)
	}
}
func setup() {
	config.SetupConfig()
	helpers.SetupLogging()
	log.SetLevel(log.DebugLevel)
	log.Infoln("[axolotl] Starting Signal for Ubuntu version", config.AppVersion)
}
func runBackend() {
	ui.SetEngine()
	//
	// ui.Engine.AddImageProvider("avatar", store.AvatarImageProvider)
	ui.SetComponent()
	go worker.RunBackend()
	if config.IsPushHelper {
		push.PushHelperProcess()
	}
}
func runUI() error {
	defer wg.Done()
	if config.Gui != "ut" && config.Gui != "me" && config.Gui != "lorca" {
		ui.RunUi(config.Gui)
		runElectron()
	} else {
		ui.RunUi(config.Gui)
	}
	os.Exit(0)
	return nil
}
func runElectron() {
	log.Infoln("[axolotl] Start electron")
	l := log.New()
	electronPath := os.Getenv("SNAP_USER_DATA")
	if len(electronPath) == 0 {
		electronPath = config.ConfigDir + "/electron"
	}
	var a, _ = astilectron.New(l, astilectron.Options{
		AppName:            "axolotl",
		AppIconDefaultPath: "axolotl-web/public/axolotl.png", // If path is relative, it must be relative to the data directory
		AppIconDarwinPath:  "axolotl-web/public/axolotl.png", // Same here
		BaseDirectoryPath:  electronPath,
		VersionElectron:    "7.1.7",
		SingleInstance:     true,
		ElectronSwitches:   []string{"--disable-dev-shm-usage", "--no-sandbox"}})
	//
	// --disable-dev-shm-usage

	defer a.Close()
	// Start astilectron

	if err := a.Start(); err != nil {
		log.Debugln(errors.Wrap(err, "main: starting astilectron failed"))
	}

	a.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
		log.Errorln("[axolotl-electron] Electron App has crashed", e)
		return
	})
	a.HandleSignals()
	// New window
	var w *astilectron.Window
	var err error
	if w, err = a.NewWindow("http://localhost:9080", &astilectron.WindowOptions{
		Center: astikit.BoolPtr(true),
		Height: astikit.IntPtr(600),
		Width:  astikit.IntPtr(600),
	}); err != nil {
		log.Debugln("[axolotl-electron]", errors.Wrap(err, "main: new window failed"))
	}

	// Create windows
	if err = w.Create(); err != nil {
		log.Debugln(errors.Wrap(err, "main: creating window failed"))
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
	defer wg.Done()
	log.Printf("[axolotl] Axolotl server started")
	// Fetch the URL.
	webserver.Run()
}

var wg sync.WaitGroup

func main() {
	setup()
	runBackend()
	log.Println("[axolotl] Setup completed")
	wg.Add(1)
	go runWebserver()
	wg.Add(1)
	go runUI()
	wg.Wait()
}
