package app

import (
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/settings"
	"github.com/nanu-c/axolotl/app/store"
)

// TODO: WIP 831: Does App need to exist in addition to WsApp?
type App struct {
	Config   *config.Config
	Settings *settings.Settings
	Store    *store.Store
}
