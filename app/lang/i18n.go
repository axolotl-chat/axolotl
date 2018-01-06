package lang

import (
	"github.com/gosexy/gettext"
)

var (
	SessionReset string
	YouLeftGroup string
)

func SetupTranslations(AppName string) {
	gettext.Textdomain(AppName)
	gettext.BindTextdomain(AppName, "./share/locale")
	gettext.SetLocale(gettext.LC_ALL, "")

	SessionReset = gettext.Gettext("Secure session reset.")
	YouLeftGroup = gettext.Gettext("You have left the group.")
}
