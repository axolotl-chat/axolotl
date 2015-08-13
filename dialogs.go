package main

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
	return getTextFromDialog("getPhoneNumber", "signInPage", "numberEntered")
}

func getVerificationCode() string {
	return getTextFromDialog("getVerificationCode", "codeVerificationPage", "codeEntered")
}
