# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

NPM_VERSION := $(shell npm --version 2>/dev/null)
NODE_VERSION := $(shell node --version 2>/dev/null)
GO_VERSION := $(shell go version 2>/dev/null)
AXOLOTL_GIT_VERSION := $(shell git tag | tail --lines=1)
AXOLOTL_VERSION := $(subst v,,$(AXOLOTL_GIT_VERSION))
CURRENT_DIR = $(shell pwd)

define APPDATA_TEXT=
\\t\t\t<release version="$(NEW_VERSION)" date="$(shell date --rfc-3339='date')">\n\
\t\t\t\t\t<url>https://github.com/nanu-c/axolotl/releases/tag/v$(NEW_VERSION)</url>\n\
\t\t\t</release>
endef
export APPDATA_TEXT

NPM=$(shell which npm)
GO=$(shell which go)
FLATPAK=$(shell which flatpak)
FLATPAK_BUILDER=$(shell which flatpak-builder)
SNAPCRAFT=$(shell which snapcraft)
SNAP=$(shell which snap)

all: clean build run

build: build-axolotl-web build-axolotl

install: install-axolotl-web install-axolotl
	@sudo install -D -m 644 $(CURRENT_DIR)/scripts/axolotl.desktop /usr/share/applications/axolotl.desktop
	@sudo install -D -m 644 $(CURRENT_DIR)/snap/gui/axolotl.png /usr/share/icons/hicolor/128x128/apps/axolotl.png

uninstall:
	@sudo rm -rf /usr/bin/axolotl

build-axolotl-web:
	$(NPM) run build --prefix axolotl-web

build-axolotl:
	$(GO) build -v .

build-translation:
	$(NPM) run translate --prefix axolotl-web

check: check-axolotl check-axolotl-web

check-axolotl-web:
	$(NPM) run test --prefix axolotl-web

check-axolotl:
	$(GO) test -race ./...

run: build
	$(GO) run .

build-dependencies: build-dependencies-axolotl-web build-dependencies-axolotl

build-dependencies-axolotl-web:
	$(NPM) install --prefix axolotl-web

build-dependencies-axolotl:
	$(GO) mod download

install-axolotl-web: build-axolotl-web
	@sudo cp -r $(CURRENT_DIR)/axolotl-web/dist /usr/bin/axolotl/axolotl-web/dist

install-axolotl: build-axolotl
	@sudo install -D -m 755 $(CURRENT_DIR)/axolotl /usr/bin/axolotl

clean:
	rm -f axolotl
	rm -rf axolotl-web/dist

update-version:
ifeq ($(NEW_VERSION),)
	@echo Please specify the new version to use! Example: "make update-version NEW_VERSION=0.9.10"
else
	@echo Replacing current version $(AXOLOTL_VERSION) with new version $(NEW_VERSION)
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' manifest.json
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' app/config/config.go
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' snap/snapcraft.yaml
	@sed -i "32i $$APPDATA_TEXT" appimage/AppDir/axolotl.appdata.xml
	@sed -i "32i $$APPDATA_TEXT" flatpak/org.nanuc.Axolotl.appdata.xml
	@echo Update complete
endif

## Flatpak
build-dependencies-flatpak:
	$(FLATPAK) install org.freedesktop.Sdk.Extension.golang//20.08
	$(FLATPAK) install org.freedesktop.Sdk.Extension.node14//20.08

build-dependencies-flatpak-web: build-dependencies-flatpak
	$(FLATPAK) install org.freedesktop.Platform//20.08
	$(FLATPAK) install org.freedesktop.Sdk//20.08
	$(FLATPAK) install io.atom.electron.BaseApp//20.08

build-dependencies-flatpak-qt: build-dependencies-flatpak
	$(FLATPAK) install org.kde.Platform//5.15
	$(FLATPAK) install org.kde.Sdk//5.15
	$(FLATPAK) install io.qt.qtwebengine.BaseApp//5.15

install-flatpak-web:
	$(FLATPAK_BUILDER) --user --install build flatpak/web/org.nanuc.Axolotl.yml

install-flatpak-qt:
	$(FLATPAK_BUILDER) --user --install build flatpak/qt/org.nanuc.Axolotl.yml

## Snap
build-snap:
	@sudo $(SNAPCRAFT)

install-snap:
	@sudo $(SNAP) install axolotl_$(AXOLOTL_VERSION)_amd64.snap --dangerous
