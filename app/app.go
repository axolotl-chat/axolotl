package app

import (
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/settings"
)

// TODO: WIP 831: Does App need to exist in addition to WsApp?
type App struct {
	Config   *config.Config
	Settings *settings.Settings
}
