package main

import "github.com/gosexy/gettext"

var (
	sessionReset string
	youLeftGroup string
)

func setupTranslations() {
	gettext.Textdomain(appName)
	gettext.BindTextdomain(appName, "./share/locale")
	gettext.SetLocale(gettext.LC_ALL, "")

	sessionReset = gettext.Gettext("Secure session reset.")
	youLeftGroup = gettext.Gettext("You have left the group.")
}
