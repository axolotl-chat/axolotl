package main

import "github.com/ttacon/libphonenumber"

func getTextFromDialog(fun, obj, signal string) string {
	win.Root().Call(fun)
	p := win.Root().ObjectByName(obj)
	ch := make(chan string)
	p.On(signal, func(text string) {
		ch <- text
	})
	text := <-ch
	return text
}

func getStoragePassword() string {
	return getTextFromDialog("getStoragePassword", "passwordPage", "passwordEntered")
}

func getPhoneNumber() string {
	n := getTextFromDialog("getPhoneNumber", "signInPage", "numberEntered")

	num, _ := libphonenumber.Parse(n, "")
	c := libphonenumber.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
	s := libphonenumber.GetNationalSignificantNumber(num)
	f := formatE164(s, c)
	return f
}

func getVerificationCode() string {
	return getTextFromDialog("getVerificationCode", "codeVerificationPage", "codeEntered")
}
