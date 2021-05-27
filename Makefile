# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

NPM_VERSION := $(shell npm --version 2>/dev/null)
NODE_VERSION := $(shell node --version 2>/dev/null)
GO_VERSION := $(shell go version 2>/dev/null)
CARGO_VERSION := $(shell cargo --version 2>/dev/null)
GIT_VERSION := $(shell git --version 2>/dev/null)
AXOLOTL_GIT_VERSION := $(shell git tag | tail --lines=1)
AXOLOTL_VERSION := $(subst v,,$(AXOLOTL_GIT_VERSION))
UNAME_S := $(shell uname -s)
UNAME_HARDWARE_PLATTFORM := $(shell uname --hardware-platform)
CURRENT_DIR = $(shell pwd)


define APPDATA_TEXT=
\\t\t\t<release version="$(NEW_VERSION)" date="$(shell date --rfc-3339='date')">\n\
\t\t\t\t\t<url>https://github.com/nanu-c/axolotl/releases/tag/v$(NEW_VERSION)</url>\n\
\t\t\t</release>
endef
export APPDATA_TEXT

NPM=$(shell which npm)
GO=$(shell which go)
GIT=$(shell which git)
CARGO=$(shell which cargo)
FLATPAK=$(shell which flatpak)
FLATPAK_BUILDER=$(shell which flatpak-builder)
SNAPCRAFT=$(shell which snapcraft)
SNAP=$(shell which snap)

all: clean build

build: build-axolotl-web build-axolotl

install: install-axolotl install-axolotl-web
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
		@echo "Found go with version $(GO_VERSION)"
		LD_LIBRARY_PATH=$(PWD) $(GO)  run .

build-dependencies: build-dependencies-axolotl-web build-dependencies-axolotl

build-dependencies-axolotl-web:
	$(NPM) install --prefix axolotl-web

build-dependencies-axolotl:
	$(GO) mod download

install-axolotl-web: build-axolotl-web
	@sudo cp -r $(CURRENT_DIR)/axolotl-web/dist /usr/bin/axolotl/axolotl-web/dist

install-axolotl: build-axolotl
	@sudo install -D -m 755 $(CURRENT_DIR)/axolotl /usr/bin/axolotl/axolotl

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

## zkgroup
build-zkgroup:
	@echo "Found cargo with version $(CARGO_VERSION)"
	@echo "Found go with version $(GO_VERSION)"
	@echo "Found git with version $(GIT_VERSION)"
ifeq ($(UNAME_S), Linux)
	@echo "get zkgroup $(PLATFORM)"
	$(GO) get -d github.com/nanu-c/zkgroup \
	&& cd $(GOPATH)/src/github.com/nanu-c/zkgroup \
	&& $(GIT) submodule update \
	&& cd $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/zkgroup \
	&& $(CARGO) build --release --verbose \
	&& mv -f $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/zkgroup/target/release/libzkgroup.so $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(UNAME_HARDWARE_PLATTFORM).so
else
	@echo architecture not (yet) supported $(UNAME_HARDWARE_PLATTFORM)
	exit 1
endif

copy-zkgroup:
ifdef GO_VERSION
	@echo "Found go with version $(GO_VERSION)"
else
	@echo go not found, please install go
	exit 1
endif
ifeq ($(UNAME_S), Linux)
	$(GO) get -d github.com/nanu-c/zkgroup
	cp $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(UNAME_HARDWARE_PLATTFORM).so $(CURRENT_DIR)/
else
	@echo "platform not supported $(UNAME_S)"
	exit 1
endif

install-zkgroup:
ifeq ($(UNAME_S), Linux)
ifeq ($(UNAME_HARDWARE_PLATTFORM), x86_64)
	@echo "install libzkgroup to /usr/lib"
	sudo cp $(CURRENT_DIR)/libzkgroup_linux_amd64.so /usr/lib/
else ifeq ($(UNAME_HARDWARE_PLATTFORM), aarch64)
	@echo "install libzkgroup to /usr/lib"
	sudo cp $(CURRENT_DIR)/libzkgroup_linux_arm64.so  /usr/lib/
else
	@echo architecture not  supported $(UNAME_HARDWARE_PLATTFORM)
	exit 1
endif
else
	@echo "platform not supported $(UNAME_S)"
	exit 1
endif

install-clickable-zkgroup:
	rm $(CURRENT_DIR)/lib/*.so |true
	$(GO) get -d github.com/nanu-c/zkgroup
	cp $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(UNAME_HARDWARE_PLATTFORM).so $(CURRENT_DIR)/lib/

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
