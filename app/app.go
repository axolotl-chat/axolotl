package app

import (
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/settings"
)

type App struct {
	Config   *config.Config
	Settings *settings.Settings
}
