package main

import (
	"flag"
	_ "image/jpeg"
	_ "image/png"

	log "github.com/Sirupsen/logrus"

	"github.com/nanu-c/qml-go"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/push"
	"github.com/nanu-c/textsecure-qml/app/store"
	"github.com/nanu-c/textsecure-qml/app/ui"
	"github.com/nanu-c/textsecure-qml/app/worker"
)

func init() {
	flag.StringVar(&config.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
}

func setup() {
	helpers.SetupLogging()
	log.Infof("Starting Signal for Ubuntu version %s", config.AppVersion)
	config.SetupConfig()
	//encrypted?

}

func RunUI() error {
	ui.SetEngine()
	ui.Engine.AddImageProvider("avatar", store.AvatarImageProvider)
	ui.InitModels()
	ui.Engine.Context().SetVar("textsecure", worker.Api)
	ui.Engine.Context().SetVar("appVersion", config.AppVersion)
	ui.Engine.Context().SetVar("cacheDir", config.CacheDir)

	ui.SetComponent()
	ui.Win.Show()

	go worker.RunBackend()
	// store.SetupDB()
	// if store.DS == nil {
	// 	var err error
	// 	log.Printf("Open unencrypted DB")
	// 	store.DS, err = store.NewStorage("")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	// store.LoadMessagesFromDB()
	if config.IsPushHelper {
		push.PushHelperProcess()
	}
	ui.Win.Wait()
	return nil
}

func main() {
	setup()
	log.Println("Setup completed")

	err := qml.Run(RunUI)
	if err != nil {
		log.Fatal(err)
	}
}
