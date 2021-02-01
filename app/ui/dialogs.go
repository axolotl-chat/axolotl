package ui

import (
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/webserver"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
)

func GetTextFromDialog(fun, obj, signal string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("[axolotl] Error: GetTextFromDialog: ", r)
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
func GetTextFromWs(fun string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("[axoltol] Error: GetTextFromDialog: ", r)
		}
	}()
	log.Debugf("[axoltol] Opening Dialog: " + fun)
	text := webserver.RequestInput(fun)
	return text
}

func GetStoragePassword() string {
	return GetTextFromWs("getStoragePassword")
}

func GetPhoneNumber() string {

	// time.Sleep(2 * time.Second)
	// n := GetTextFromDialog("getPhoneNumber", "signinPage", "numberEntered")
	n := GetTextFromWs("getPhoneNumber")
	num, _ := libphonenumber.Parse(n, "")
	c := libphonenumber.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
	s := libphonenumber.GetNationalSignificantNumber(num)
	f := contact.FormatE164(s, c)
	return f
}

func GetVerificationCode() string {
	return GetTextFromWs("getVerificationCode")
}
func GetPin() string {
	return GetTextFromWs("getPin")
}
func GetEncryptionPw() string {
	return GetTextFromWs("getEncryptionPw")
}
func ShowError(err error) {
	webserver.ShowError(err.Error())
	log.Errorln("[axolotl] error: ", err.Error())
}
func ClearError() {
	webserver.ClearError()
}
