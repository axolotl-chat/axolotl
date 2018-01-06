package main

import (
	"flag"
	_ "image/jpeg"
	_ "image/png"

	log "github.com/Sirupsen/logrus"

	"github.com/amlwwalker/qml"
	"github.com/nanu-c/textsecure-qml/app/models"
	"github.com/nanu-c/textsecure-qml/app/push"
	"github.com/nanu-c/textsecure-qml/app/store"
	"github.com/nanu-c/textsecure-qml/app/ui"
	"github.com/nanu-c/textsecure-qml/app/worker"
)

func init() {
	flag.StringVar(&store.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
}

func setup() {
	models.SetupLogging()
	store.SetupConfig()
	if store.IsPushHelper {
		push.PushHelperProcess()
	}
}

func RunUI() error {
	ui.SetEngine()
	ui.Engine.AddImageProvider("avatar", store.AvatarImageProvider)
	ui.InitModels()
	ui.Engine.Context().SetVar("textsecure", worker.Api)
	ui.Engine.Context().SetVar("appVersion", store.AppVersion)

	ui.SetComponent()
	ui.Win.Show()

	go worker.RunBackend()
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
