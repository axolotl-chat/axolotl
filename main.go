package main

import (
	"bufio"
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os/exec"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/push"
	"github.com/nanu-c/textsecure-qml/app/ui"
	"github.com/nanu-c/textsecure-qml/app/webserver"
	"github.com/nanu-c/textsecure-qml/app/worker"
)

var sys string

func init() {
	flag.StringVar(&config.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
	flag.StringVar(&sys, "sys", "", "Usage")
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
	log.Infof("Starting Signal for Ubuntu version %s", config.AppVersion)
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
	// cmd := exec.Command("webapp-container", "http://[::1]:8080/")

	var cmd *exec.Cmd
	if sys == "ut" {
		cmd = exec.Command("qmlscene", "--scaling", "qml/MainUt.qml")
	} else if sys == "me" {
		cmd = exec.Command("/home/nanu/Qt/5.13.0/gcc_64/bin/qmlscene", "--scaling", "qml/Main.qml")

	} else {
		cmd = exec.Command("qmlscene", "--scaling", "qml/Main.qml")

	}
	// cmd := exec.Command("webapp-container", "--app-id='textsecure.nanuc'", "$@", "axolotl-web/index.html")
	log.Printf("Starting Axolotl-gui")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err := cmd.Run()
	go print(stdout)
	go print(stderr)
	log.Printf("Axolotl-gui finished with error: %v", err)
	return nil
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
