package main

import (
	"bufio"
	"flag"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"sync"

	astilectron "github.com/asticode/go-astilectron"
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
	var a, _ = astilectron.New(astilectron.Options{
		AppName:            "axolotl",
		AppIconDefaultPath: "axolotl-web/public/axolotl.png", // If path is relative, it must be relative to the data directory
		AppIconDarwinPath:  "axolotl-web/public/axolotl.png", // Same here
		BaseDirectoryPath:  "dist",
	})
	defer a.Close()

	// Start astilectron
	a.Start()
	var w, _ = a.NewWindow("http://localhost:9080", &astilectron.WindowOptions{
		Center: astilectron.PtrBool(true),
		Height: astilectron.PtrInt(600),
		Width:  astilectron.PtrInt(600),
	})
	w.Create()
	if config.ElectronDebug {
		w.OpenDevTools()
	}
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
