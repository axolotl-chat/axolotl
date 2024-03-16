# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

.PHONY: build build-axolotl-web build-axolotl install install-axolotl uninstall uninstall-axolotl clean clean-axolotl-web clean-axolotl check check-axolotl-web check-axolotl build-translation download-dependencies download-dependencies-axolotl-web download-dependencies-axolotl update-version download-dependencies-flatpak install-flatpak build-snap install-snap

AXOLOTL_GIT_VERSION := $(shell git tag | tail --lines=1)
AXOLOTL_VERSION := $(subst v,,$(AXOLOTL_GIT_VERSION))
ARCH := $(shell uname --machine)
CURRENT_DIR := $(shell pwd)

APP_ID := axolotl
define APPDATA_TEXT=
\\t\t<release version="$(NEW_VERSION)" date="$(shell date --rfc-3339='date')">\n\
\t\t\t\t<url>https://github.com/nanu-c/axolotl/releases/tag/v$(NEW_VERSION)</url>\n\
\t\t</release>
endef
export APPDATA_TEXT

YARN := $(shell which yarn 2>/dev/null)
CARGO := $(shell which cargo 2>/dev/null)
FLATPAK := $(shell which flatpak 2>/dev/null)
FLATPAK_BUILDER := $(shell which flatpak-builder 2>/dev/null)
SNAPCRAFT := $(shell which snapcraft 2>/dev/null)
SNAP := $(shell which snap 2>/dev/null)

# See common variable names: https://cmake.org/cmake/help/latest/module/GNUInstallDirs.html
PREFIX := "/usr"
BINDIR := $(PREFIX)/bin
LIBDIR := $(PREFIX)/lib
DATADIR := $(PREFIX)/share
CARGO_PREFIX := ${HOME}/.cargo/bin

all: build

build: build-axolotl-web build-axolotl

install: install-axolotl install-metadata

uninstall: uninstall-axolotl

clean: clean-axolotl-web clean-axolotl

check: check-axolotl-web check-axolotl

download-dependencies: download-dependencies-axolotl-web download-dependencies-axolotl

# axolotl
download-dependencies-axolotl: Cargo.toml Cargo.lock
	$(CARGO) fetch --verbose

build-axolotl: download-dependencies-axolotl
	@echo "Building axolotl..."
	$(CARGO) build --features tauri --release --verbose

install-axolotl: build-axolotl
	@echo "Installing axolotl..."
	@install -D -m 755 $(CURRENT_DIR)/target/release/axolotl $(DESTDIR)$(BINDIR)/axolotl/axolotl

uninstall-axolotl:
	@echo "Uninstalling axolotl..."
	@rm -rf $(DESTDIR)$(BINDIR)/axolotl

clean-axolotl:
	rm -f $(CURRENT_DIR)/target

# axolotl-web
download-dependencies-axolotl-web: axolotl-web/package.json axolotl-web/yarn.lock
	$(YARN) --cwd axolotl-web install --frozen-lockfile

build-axolotl-web: download-dependencies-axolotl-web
	@echo "Building axolotl-web..."
	$(YARN) --cwd axolotl-web run build

check-axolotl-web:
	$(YARN) --cwd axolotl-web run depcheck
	$(YARN) --cwd axolotl-web run lint
	$(YARN) --cwd axolotl-web run test

clean-axolotl-web:
	rm -rf $(CURRENT_DIR)/axolotl-web/dist

# utilities
build-translation:
	$(YARN) --cwd axolotl-web run translate

update-version:
ifeq ($(NEW_VERSION),)
	@echo 'Please specify the new version to use! Example: "make update-version NEW_VERSION=0.9.10"'
else
	@echo "Replacing current version $(AXOLOTL_VERSION) with new version $(NEW_VERSION)"
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' manifest.json
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' snap/snapcraft.yaml
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' docs/INSTALL.md
	@sed -i "32i $$APPDATA_TEXT" data/axolotl.metainfo.xml
	@echo "Update complete"
endif

install-metadata:
	@install -D -m 644 data/icons/icon.png $(DESTDIR)$(DATADIR)/icons/hicolor/128x128/apps/$(APP_ID).png
	@install -D -m 644 data/axolotl.metainfo.xml $(DESTDIR)$(DATADIR)/metainfo/$(APP_ID).metainfo.xml
	@sed -i 's/@app-id@/$(APP_ID)/' $(DESTDIR)$(DATADIR)/metainfo/$(APP_ID).metainfo.xml
	@install -D -m 644 data/axolotl.desktop $(DESTDIR)$(DATADIR)/applications/$(APP_ID).desktop
	@sed -i 's/@icon@/$(APP_ID).png/' $(DESTDIR)$(DATADIR)/applications/$(APP_ID).desktop

# Flatpak
download-dependencies-flatpak:
	$(FLATPAK) install org.gnome.Platform//45
	$(FLATPAK) install org.gnome.Sdk//45
	$(FLATPAK) install org.freedesktop.Sdk.Extension.node18//22.08
	$(FLATPAK) install org.freedesktop.Sdk.Extension.rust-stable//22.08

build-flatpak:
	$(FLATPAK_BUILDER) flatpak/build --force-clean --ccache flatpak/org.nanuc.Axolotl.yml

install-flatpak:
	$(FLATPAK_BUILDER) flatpak/build --force-clean --user --install flatpak/org.nanuc.Axolotl.yml

debug-flatpak:
	$(FLATPAK_BUILDER) flatpak/build --run --verbose flatpak/org.nanuc.Axolotl.yml sh

uninstall-flatpak:
	$(FLATPAK) uninstall org.nanuc.Axolotl

clean-flatpak:
	rm -rf flatpak/.flatpak-builder flatpak/build

# Snap
build-snap:
	@sudo $(SNAPCRAFT)

install-snap:
	@sudo $(SNAP) install axolotl_$(AXOLOTL_VERSION)_amd64.snap --dangerous
