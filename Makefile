# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

.PHONY: build clean build-axolotl-web build-axolotl install install-axolotl install-axolotl-web uninstall build-translation run check check-axolotl check-axolotl-web build-dependencies build-dependencies-axolotl-web build-dependencies-axolotl update-version build-zkgroup copy-zkgroup install-zkgroup uninstall-zkgroup install-clickable-zkgroup uninstall-clickable-zkgroup build-dependencies-flatpak build-dependencies-flatpak-web build-dependencies-flatpak-qt install-flatpak-web install-flatpak-qt build-snap install-snap check-platform-deb-arm64 dependencies-deb-arm64 build-deb-arm64 prebuild-package-deb-arm64 build-package-deb-arm64 install-deb-arm64 uninstall-deb-arm64 clean-deb-arm64 package-clean-deb-arm64

NPM_VERSION := $(shell npm --version 2>/dev/null)
NODE_VERSION := $(shell node --version 2>/dev/null)
GO_VERSION := $(shell go version 2>/dev/null)
CARGO_VERSION := $(shell cargo --version 2>/dev/null)
GIT_VERSION := $(shell git --version 2>/dev/null)
AXOLOTL_GIT_VERSION := $(shell git tag | tail --lines=1)
AXOLOTL_VERSION := $(subst v,,$(AXOLOTL_GIT_VERSION))
UNAME_S := $(shell uname -s)
HARDWARE_PLATFORM := $(shell uname --machine)
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
	&& mv -f $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/zkgroup/target/release/libzkgroup.so $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so
else
	@echo architecture not (yet) supported $(HARDWARE_PLATFORM)
	exit 1
endif

copy-zkgroup:
	$(GO) get -d github.com/nanu-c/zkgroup
	cp $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so $(CURRENT_DIR)/

install-zkgroup:
	sudo cp $(CURRENT_DIR)/libzkgroup_linux_$(HARDWARE_PLATFORM).so /usr/lib/

uninstall-zkgroup:
	sudo rm -f /usr/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so

install-clickable-zkgroup:
	$(GO) get -d github.com/nanu-c/zkgroup
	cp $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so $(CURRENT_DIR)/lib/

uninstall-clickable-zkgroup:
	rm -f $(CURRENT_DIR)/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so

## Flatpak
build-dependencies-flatpak:
	$(FLATPAK) install org.freedesktop.Sdk.Extension.golang//20.08
	$(FLATPAK) install org.freedesktop.Sdk.Extension.node14//20.08

build-dependencies-flatpak-web: build-dependencies-flatpak
	$(FLATPAK) install org.freedesktop.Platform//20.08
	$(FLATPAK) install org.freedesktop.Sdk//20.08
	$(FLATPAK) install org.electronjs.Electron2.BaseApp//20.08

build-dependencies-flatpak-qt: build-dependencies-flatpak
	$(FLATPAK) install org.kde.Platform//5.15
	$(FLATPAK) install org.kde.Sdk//5.15
	$(FLATPAK) install io.qt.qtwebengine.BaseApp//5.15

install-flatpak-web:
	$(FLATPAK_BUILDER) --user --install --force-clean build flatpak/web/org.nanuc.Axolotl.yml

install-flatpak-qt:
	$(FLATPAK_BUILDER) --user --install --force-clean build flatpak/qt/org.nanuc.Axolotl.yml

## Snap
build-snap:
	@sudo $(SNAPCRAFT)

install-snap:
	@sudo $(SNAP) install axolotl_$(AXOLOTL_VERSION)_amd64.snap --dangerous

## Debian arm64 building and packaging
## Please get the source via
##  go get -d -u github.com/nanu-c/axolotl/
check-platform-deb-arm64:
	@echo "Building Axolotl for Debian arm64/aarch64"
  ifneq ($(UNAME_S),Linux)
	@echo "Platform unsupported - only available for Linux" && exit 1
  endif
  ifneq ($(shell uname -m),aarch64)
	@echo "Machine unsupported - only available for arm64/aarch64" && exit 1
  endif
  ifneq ($(shell which apt),/usr/bin/apt)
	@echo "OS unsupported - apt not found" && exit 1
  endif

dependencies-deb-arm64: check-platform-deb-arm64
	@echo "Installing dependencies for building Axolotl..."
	@sudo apt update
	@sudo apt install nano git golang nodejs npm python

build-deb-arm64: check-platform-deb-arm64 dependencies-deb-arm64
	@echo "Downloading (go)..."
	@cd $(CURRENT_DIR) && go mod download
	@echo "Installing (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm ci
	@echo "Building (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm run build
	@mkdir -p $(CURRENT_DIR)/build/linux-arm64/axolotl-web
	@echo "Building (go)..."
	@cd $(CURRENT_DIR) && env GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/axolotl .
	@cp -r axolotl-web/dist build/linux-arm64/axolotl-web
	@cp -r guis build/linux-arm64
	@echo "Building complete."

prebuild-package-deb-arm64: package-clean-deb-arm64
	@echo "Prebuilding Debian package..."
# Get the source tarball
	@wget https://github.com/nanu-c/axolotl/archive/v$(AXOLOTL_VERSION).tar.gz -O $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION).tar.gz
# Prepare packaging folder
	@mkdir -p $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl
	@cp -r $(CURRENT_DIR)/build/linux-arm64/* $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl
	@cp $(CURRENT_DIR)/deb/LICENSE $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/LICENSE
# Run debmake
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debmake -e arno_nuehm@riseup.net -f "Arno Nuehm" -m
# Create target folders and copy additional files into package folder
	@mkdir -p $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps
	@mkdir -p $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications
	@mkdir -p $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin
	@mkdir -p $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/etc/profile.d
	@cp $(CURRENT_DIR)/README.md $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/README.Debian
	@cp $(CURRENT_DIR)/deb/axolotl.png $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps/axolotl.png
	@cp $(CURRENT_DIR)/deb/axolotl.desktop $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications
	@cp $(CURRENT_DIR)/deb/axolotl.sh $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/etc/profile.d
	@cp $(CURRENT_DIR)/deb/axolotl.install $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian
	@cp $(CURRENT_DIR)/deb/postinst $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian
	@cp $(CURRENT_DIR)/deb/control $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/control
	@wget https://github.com/nanu-c/zkgroup/raw/main/lib/libzkgroup_linux_aarch64.so -P  $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/lib
	@mv $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl/axolotl $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin
	@echo "Prebuilding Debian package complete"

build-package-deb-arm64:
	@echo "Building Debian package..."
# Prompt to edit changelog file
	@nano $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
# Prompt to edit copyright file
	@nano $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
# Build Debian package
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debuild -i -us -uc -b

install-deb-arm64: uninstall-deb-arm64
# Use for testing purposes only after prebuild-package-arm64
	@echo "Installing Axolotl"
# Copy libzkgroup
	@sudo wget https://github.com/nanu-c/zkgroup/raw/main/lib/libzkgroup_linux_aarch64.so -P /usr/lib
	@sudo mkdir -p /usr/share/axolotl
	@sudo cp -r $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl/* /usr/share/axolotl
	@sudo mv /usr/share/axolotl/axolotl /usr/bin/
	@sudo cp $(CURRENT_DIR)/deb/axolotl.desktop /usr/share/applications
	@sudo cp $(CURRENT_DIR)/deb/axolotl.png /usr/share/icons/hicolor/128x128/apps
	@sudo cp $(CURRENT_DIR)/deb/axolotl.sh /etc/profile.d
	@bash -c "source /etc/profile.d/axolotl.sh"
	@echo "Installation complete"

uninstall-deb-arm64:
	@sudo rm -rf /usr/share/axolotl
	@sudo rm -f /usr/bin/axolotl
	@sudo rm -f /usr/share/applications/axolotl.desktop
	@sudo rm -f /usr/share/icons/hicolor/128x128/apps/axolotl.png
	@sudo rm -f /etc/profile.d/axolotl.sh
	@sudo rm -f /usr/lib/libzkgroup_linux_aarch64.so
	@echo "Removing complete"

clean-deb-arm64:
	@rm -rf $(CURRENT_DIR)/build

package-clean-deb-arm64:
	@rm -rf $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)
