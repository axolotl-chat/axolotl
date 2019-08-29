package ui

import (
	log "github.com/sirupsen/logrus"
	qml "github.com/nanu-c/qml-go"
	"github.com/nanu-c/textsecure"
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
	settings.SettingsModel.Registered = true
}
func SetComponent() error {
	component, err := Engine.LoadFile(config.MainQml)
	if err != nil {
		log.Println(err)
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
	} else {
		if settings.SettingsModel.Registered {
			log.Debugf("Already registered")

		}
	}
	Engine.Context().SetVar("contactsModel", store.ContactsModel)
	Engine.Context().SetVar("settingsModel", settings.SettingsModel)
	Engine.Context().SetVar("sessionsModel", store.SessionsModel)
	// textsecure.LinkedDevices()
	Engine.Context().SetVar("linkedDevicesModel", store.LinkedDevicesModel)
	Engine.Context().SetVar("storeModel", store.DS)

	go store.UpdateTimestamps()
}
