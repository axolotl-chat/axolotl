package ui

import (
	"github.com/nanu-c/textsecure-qml/app/contact"
	"github.com/nanu-c/textsecure-qml/app/webserver"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
)

func GetTextFromDialog(fun, obj, signal string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("Error: GetTextFromDialog: ", r)
		}
	}()
	log.Debugf("Opening Dialog: " + fun)
	// Win.Root().Call(fun)
	// p := Win.Root().ObjectByName(obj)
	ch := make(chan string)
	// p.On(signal, func(text string) {
	// 	ch <- text
	// })
	text := <-ch
	return text
}
func GetTextFromWs(fun, obj, signal string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("Error: GetTextFromDialog: ", r)
		}
	}()
	log.Debugf("Opening Dialog: " + fun)
	// Win.Root().Call(fun)
	// p := Win.Root().ObjectByName(obj)
	// ch := make(chan string)

	// p.On(signal, func(text string) {
	// 	ch <- text
	// })

	text := webserver.RequestInput(fun)
	return text
}

func GetStoragePassword() string {
	return GetTextFromWs("getStoragePassword", "passwordPage", "passwordEntered")
}

func GetPhoneNumber() string {

	// time.Sleep(2 * time.Second)
	// n := GetTextFromDialog("getPhoneNumber", "signinPage", "numberEntered")
	n := GetTextFromWs("getPhoneNumber", "signinPage", "numberEntered")
	num, _ := libphonenumber.Parse(n, "")
	c := libphonenumber.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
	s := libphonenumber.GetNationalSignificantNumber(num)
	f := contact.FormatE164(s, c)
	return f
}

func GetVerificationCode() string {
	return GetTextFromWs("getVerificationCode", "codeVerificationPage", "codeEntered")
}
func ShowError(err error) {
	// Win.Root().Call("error", err.Error())
	log.Errorf(err.Error())
}
