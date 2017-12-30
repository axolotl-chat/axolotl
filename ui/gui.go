package ui

import (
	log "github.com/Sirupsen/logrus"
	"github.com/janimo/textsecure"
	"github.com/nanu-c/textsecure-qml/store"
	qml "gopkg.in/qml.v1"
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
func ShowError(err error) {
	Win.Root().Call("error", err.Error())
	log.Errorf(err.Error())
}
func RegistrationDone() {
	log.Println("Registered")
	Win.Root().Call("registered")
	textsecure.WriteConfig(store.ConfigFile, store.Config)
}
func SetComponent() error {
	component, err := Engine.LoadFile(store.MainQml)
	if err != nil {
		return err
	}
	Win = component.CreateWindow(nil)
	return nil
}
func SetEngine() {
	Engine = qml.NewEngine()
}
