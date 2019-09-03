package main

import (
	"bufio"
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/push"
	"github.com/nanu-c/textsecure-qml/app/ui"
	"github.com/nanu-c/textsecure-qml/app/webserver"
	"github.com/nanu-c/textsecure-qml/app/worker"
)

func init() {
	flag.StringVar(&config.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
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
	go webserver.Run()

	config.SetupConfig()
	helpers.SetupLogging()
	log.SetLevel(log.DebugLevel)
	// log.SetLevel(log.DebugLevel)
	log.Infof("Starting Signal for Ubuntu version %s", config.AppVersion)
}

func RunUI() error {
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
	// cmd := exec.Command("webapp-container", "http://[::1]:8080/")
	cmd := exec.Command("qmlscene", "qml/Main.qml")
	// cmd := exec.Command("webapp-container", "--app-id='textsecure.nanuc'", "$@", "axolotl-web/index.html")
	log.Printf("Starting Axolotl-gui")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err := cmd.Run()
	go print(stdout)
	go print(stderr)
	log.Printf("Axolotl-gui finished with error: %v", err)

	// ui.Wmsgstr ""

	return nil
}

func main() {
	setup()
	log.Println("Setup completed")
	RunUI()
	// err := qml.Run(RunUI)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
