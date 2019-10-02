package main

import (
	"bufio"
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"sync"

	astilectron "github.com/asticode/go-astilectron"
	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/push"
	"github.com/nanu-c/textsecure-qml/app/ui"
	"github.com/nanu-c/textsecure-qml/app/webserver"
	"github.com/nanu-c/textsecure-qml/app/worker"
)

var e string

func init() {
	flag.StringVar(&config.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
	flag.StringVar(&config.Gui, "e", "", "use either electron, ut, lorca or me")
}
func print(stdout io.ReadCloser) {
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
}

func setup() {
	config.SetupConfig()
	helpers.SetupLogging()
	log.SetLevel(log.DebugLevel)
	log.Infoln("Starting Signal for Ubuntu version", config.AppVersion)
}
func runBackend() {
	ui.SetEngine()
	//
	// ui.Engine.AddImageProvider("avatar", store.AvatarImageProvider)
	ui.InitModels()
	//
	// ui.Engine.Context().SetVar("textsecure", worker.Api)
	// ui.Engine.Context().SetVar("appVersion", config.AppVersion)
	// ui.Engine.Context().SetVar("cacheDir", config.CacheDir)
	ui.SetComponent()
	//
	// ui.Win.Show()
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
	log.Infoln("Start electron")
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
	w.OpenDevTools()
	// Blocking pattern
	a.Wait()
}
func runWebserver() {
	// Decrement the counter when the goroutine completes.
	defer wg.Done()
	log.Printf("Axolotl server started")

	// Fetch the URL.
	webserver.Run()
}

var wg sync.WaitGroup

func main() {

	setup()
	runBackend()
	log.Println("Setup completed")
	wg.Add(1)
	go runWebserver()
	wg.Add(1)
	go runUI()
	wg.Wait()
}
