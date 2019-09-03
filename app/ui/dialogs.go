package ui

import (
	"github.com/nanu-c/textsecure-qml/app/contact"
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

func GetStoragePassword() string {
	return GetTextFromDialog("getStoragePassword", "passwordPage", "passwordEntered")
}

func GetPhoneNumber() string {
	n := GetTextFromDialog("getPhoneNumber", "signinPage", "numberEntered")
	num, _ := libphonenumber.Parse(n, "")
	c := libphonenumber.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
	s := libphonenumber.GetNationalSignificantNumber(num)
	f := contact.FormatE164(s, c)
	return f
}

func GetVerificationCode() string {
	return GetTextFromDialog("getVerificationCode", "codeVerificationPage", "codeEntered")
}
func ShowError(err error) {
	// Win.Root().Call("error", err.Error())
	log.Errorf(err.Error())
}
