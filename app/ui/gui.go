package ui

import (
	log "github.com/Sirupsen/logrus"
	"github.com/morph027/textsecure"
	qml "github.com/nanu-c/qml-go"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/settings"
	"github.com/nanu-c/textsecure-qml/app/store"
)

var Win *qml.Window
var Engine *qml.Engine

func GroupUpdateMsg(tels []string, title string) string {
	s := ""
	if len(tels) > 0 {
		for _, t := range tels {
			s += store.TelToName(t) + ", "
		}
		s = s[:len(s)-2] + " joined the group. "
	}

	return s + "Title is now '" + title + "'."
}
func RegistrationDone() {
	log.Println("Registered")
	Win.Root().Call("registered")
	textsecure.WriteConfig(config.ConfigFile, config.Config)
}
func SetComponent() error {
	component, err := Engine.LoadFile(config.MainQml)
	if err != nil {
		return err
	}
	Win = component.CreateWindow(nil)
	return nil
}
func SetEngine() {
	Engine = qml.NewEngine()
}
func InitModels() {
	var err error
	settings.SettingsModel, err = settings.LoadSettings()
	if err != nil {
		log.Println(err)
	}
	Engine.Context().SetVar("contactsModel", store.ContactsModel)
	Engine.Context().SetVar("settingsModel", settings.SettingsModel)
	Engine.Context().SetVar("sessionsModel", store.SessionsModel)
	// textsecure.LinkedDevices()
	Engine.Context().SetVar("linkedDevicesModel", store.LinkedDevicesModel)
	Engine.Context().SetVar("storeModel", store.DS)

	go store.UpdateTimestamps()
}
