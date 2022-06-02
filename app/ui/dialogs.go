package ui

import (
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/webserver"
	"github.com/signal-golang/libphonenumber"
	log "github.com/sirupsen/logrus"
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
func GetTextFromWs(fun string, wsApp *webserver.WsApp) string {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("[axolotl] Error: GetTextFromDialog: ", r)
		}
	}()
	log.Debugf("[axolotl] Opening Dialog: " + fun)
	text := wsApp.RequestInput(fun)
	log.Debugln("[axolotl] Dialog closed", fun)
	return text
}

func GetStoragePassword(wsApp *webserver.WsApp) string {
	return GetTextFromWs("getStoragePassword", wsApp)
}

func GetPhoneNumber(wsApp *webserver.WsApp) string {

	// time.Sleep(2 * time.Second)
	// n := GetTextFromDialog("getPhoneNumber", "signinPage", "numberEntered")
	n := GetTextFromWs("getPhoneNumber", wsApp)
	num, _ := libphonenumber.Parse(n, "")
	c := libphonenumber.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
	s := libphonenumber.GetNationalSignificantNumber(num)
	f := contact.FormatE164(s, c)
	return f
}

func GetVerificationCode(wsApp *webserver.WsApp) string {
	return GetTextFromWs("getVerificationCode", wsApp)
}
func GetPin(wsApp *webserver.WsApp) string {
	return GetTextFromWs("getPin", wsApp)
}
func GetCaptchaToken(wsApp *webserver.WsApp) string {
	return GetTextFromWs("getCaptchaToken", wsApp)
}
func GetEncryptionPw(wsApp *webserver.WsApp) string {
	return GetTextFromWs("getEncryptionPw", wsApp)
}
func GetUsername(wsApp *webserver.WsApp) string {
	return GetTextFromWs("getUsername", wsApp)
}
func ShowError(err error, wsApp *webserver.WsApp) {
	wsApp.ShowError(err.Error())
	log.Errorln("[axolotl] error: ", err.Error())
}
func ClearError(wsApp *webserver.WsApp) {
	wsApp.ClearError()
}
