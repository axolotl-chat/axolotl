package worker

import (
	log "github.com/Sirupsen/logrus"

	"github.com/godbus/dbus"
)

func notification() {
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	obj := conn.Object("com.ubuntu.Postal", "/com/ubuntu/Postal/textsecure_2Enanuc")
	call := obj.Call("com.ubuntu.Postal.Post", 0, "textsecure.nanuc_textsecure", uint32(0),
		"", "Test", "This is a test of the DBus bindings for go.", []string{},
		map[string]dbus.Variant{}, int32(5000))
	if call.Err != nil {
		log.Printf(call.Err.Error())
	}
}
