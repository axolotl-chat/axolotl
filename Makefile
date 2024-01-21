# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

.PHONY: build clean build-axolotl-web build-axolotl install install-axolotl install-axolotl-web uninstall build-translation run check check-axolotl check-axolotl-web build-dependencies build-dependencies-axolotl-web build-dependencies-axolotl update-version build-dependencies-flatpak build-dependencies-flatpak-web build-dependencies-flatpak-qt install-flatpak-web install-flatpak-qt build-snap install-snap check-platform-deb-arm64 dependencies-deb-arm64 build-deb-arm64 prebuild-package-deb-arm64 build-package-deb-arm64 install-deb-arm64 uninstall-deb-arm64 check-platform-deb-arm64-cc dependencies-deb-arm64-cc build-deb-arm64-cc prebuild-package-deb-arm64-cc build-package-deb-arm64-cc clean-deb-arm64 package-clean-deb-arm64 uninstall-deb-dependencies-cc

NPM_VERSION := $(shell npm --version 2>/dev/null)
NODE_VERSION := $(shell node --version 2>/dev/null)
CARGO_VERSION := $(shell cargo --version 2>/dev/null)
GIT_VERSION := $(shell git --version 2>/dev/null)
AXOLOTL_GIT_VERSION := $(shell git tag | tail --lines=1)
AXOLOTL_VERSION := $(subst v,,$(AXOLOTL_GIT_VERSION))
UNAME_S := $(shell uname -s)
HARDWARE_PLATFORM := $(shell uname --machine)
CURRENT_DIR := $(shell pwd)
DEBIAN_VERSION := $(shell lsb_release -cs)

define APPDATA_TEXT=
\\t\t<release version="$(NEW_VERSION)" date="$(shell date --rfc-3339='date')">\n\
\t\t\t\t<url>https://github.com/nanu-c/axolotl/releases/tag/v$(NEW_VERSION)</url>\n\
\t\t</release>
endef
export APPDATA_TEXT

NPM := $(shell which npm 2>/dev/null)
GIT := $(shell which git 2>/dev/null)
CARGO := $(shell which cargo 2>/dev/null)
FLATPAK := $(shell which flatpak 2>/dev/null)
FLATPAK_BUILDER := $(shell which flatpak-builder 2>/dev/null)
SNAPCRAFT := $(shell which snapcraft 2>/dev/null)
SNAP := $(shell which snap 2>/dev/null)
APT := $(shell which apt 2>/dev/null)
WGET := $(shell which wget 2>/dev/null)
RUST := $(shell which rustup 2>/dev/null)
CROSS := $(shell which cross 2>/dev/null)
DOCKER := $(shell which docker 2>/dev/null)
ASTILECTRON_BUILDER := $(shell which astilectron-bundler 2>/dev/null)

DESTDIR := /
INSTALL_PREFIX := usr/bin
LIBRARY_PREFIX := usr/lib
SHARE_PREFIX := usr/share
CARGO_PREFIX := ${HOME}/.cargo/bin

all: clean build

build: build-axolotl-web build-axolotl

install: install-axolotl install-axolotl-web
	@sudo install -D -m 644 $(CURRENT_DIR)/scripts/axolotl.desktop $(DESTDIR)$(SHARE_PREFIX)/applications/axolotl.desktop
	@sudo install -D -m 644 $(CURRENT_DIR)/data/icons/axolotl.png $(DESTDIR)$(SHARE_PREFIX)/icons/hicolor/128x128/apps/axolotl.png

uninstall: uninstall-axolotl uninstall-axolotl-web 

check: check-axolotl check-axolotl-web

build-dependencies: build-dependencies-axolotl-web

# axolotl
build-axolotl:
	@echo "Building axolotl..."
	$(CARGO) build --features tauri

check-axolotl:
	$(GO) test -race ./...

install-axolotl:
	@echo "Installing axolotl..."
	@install -D -m 755 $(CURRENT_DIR)/axolotl $(DESTDIR)$(INSTALL_PREFIX)/axolotl/axolotl

uninstall-axolotl:
	@echo "Uninstalling axolotl..."
	@rm -rf $(DESTDIR)$(INSTALL_PREFIX)/axolotl

# axolotl-web
build-dependencies-axolotl-web:
	$(NPM) install --prefix axolotl-web

build-axolotl-web:
	@echo "Building axolotl-web..."
	$(NPM) run build --prefix axolotl-web

check-axolotl-web:
	$(NPM) run test --prefix axolotl-web

install-axolotl-web:
	@echo "Installing axolotl-web..."
	@mkdir -p $(DESTDIR)$(INSTALL_PREFIX)/axolotl/axolotl-web/
	@cp -r $(CURRENT_DIR)/axolotl-web/dist $(DESTDIR)$(INSTALL_PREFIX)/axolotl/axolotl-web/dist

uninstall-axolotl-web:
	@echo "Uninstalling axolotl-web..."
	@rm -rf $(DESTDIR)$(INSTALL_PREFIX)/axolotl/axolotl-web

## utilities
build-translation:
	$(NPM) run translate --prefix axolotl-web

run:
	@echo "Found go with version $(GO_VERSION)"
	LD_LIBRARY_PATH=$(PWD) $(GO) run .

clean:
	rm -f $(CURRENT_DIR)/axolotl
	rm -rf $(CURRENT_DIR)/axolotl-web/dist

update-version:
ifeq ($(NEW_VERSION),)
	@echo 'Please specify the new version to use! Example: "make update-version NEW_VERSION=0.9.10"'
else
	@echo "Replacing current version $(AXOLOTL_VERSION) with new version $(NEW_VERSION)"
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' manifest.json
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' snap/snapcraft.yaml
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' docs/INSTALL.md
	@sed -i "32i $$APPDATA_TEXT" appimage/AppDir/axolotl.appdata.xml
	@sed -i "32i $$APPDATA_TEXT" flatpak/org.nanuc.Axolotl.appdata.xml
	@echo "Update complete"
endif

## Electron bundler
build-dependencies-axolotl-electron-bundle:
	$(GO) install github.com/asticode/go-astilectron-bundler/astilectron-bundler@latest

build-axolotl-electron-bundle:
	@echo "Building axolotl electron bundle..."
	$(ASTILECTRON_BUILDER)

install-axolotl-electron-bundle:
	@echo "Installing axolotl electron bundle..."
	@install -D -m 755 $(CURRENT_DIR)/output/$(GOOS)-$(GOARCH)/axolotl-electron-bundle $(DESTDIR)$(INSTALL_PREFIX)/

## Flatpak
build-dependencies-flatpak:
	$(FLATPAK) install org.freedesktop.Sdk.Extension.node18//22.08
	$(FLATPAK) install org.freedesktop.Sdk.Extension.rust-stable//22.08

build-dependencies-flatpak-web: build-dependencies-flatpak
	$(FLATPAK) install org.freedesktop.Platform//22.08
	$(FLATPAK) install org.freedesktop.Sdk//22.08
	$(FLATPAK) install org.electronjs.Electron2.BaseApp//22.08

build-dependencies-flatpak-qt: build-dependencies-flatpak
	$(FLATPAK) install org.kde.Platform//5.15-22.08
	$(FLATPAK) install org.kde.Sdk//5.15-22.08
	$(FLATPAK) install io.qt.qtwebengine.BaseApp//5.15-22.08

build-flatpak-web:
	$(FLATPAK_BUILDER) flatpak/build --verbose --force-clean --ccache flatpak/web/org.nanuc.Axolotl.yml

build-flatpak-qt:
	$(FLATPAK_BUILDER) flatpak/build --verbose --force-clean --ccache flatpak/qt/org.nanuc.Axolotl.yml

install-flatpak-web:
	$(FLATPAK_BUILDER) --user --install --force-clean flatpak/build flatpak/web/org.nanuc.Axolotl.yml

install-flatpak-qt:
	$(FLATPAK_BUILDER) --user --install --force-clean flatpak/build flatpak/qt/org.nanuc.Axolotl.yml

debug-flatpak-web:
	$(FLATPAK_BUILDER) --run --verbose flatpak/build flatpak/web/org.nanuc.Axolotl.yml sh

debug-flatpak-qt:
	$(FLATPAK_BUILDER) --run --verbose flatpak/build flatpak/qt/org.nanuc.Axolotl.yml sh

uninstall-flatpak:
	$(FLATPAK) uninstall org.nanuc.Axolotl

## Snap
build-snap:
	@sudo $(SNAPCRAFT)

install-snap:
	@sudo $(SNAP) install axolotl_$(AXOLOTL_VERSION)_amd64.snap --dangerous

## Debian arm64 building/cross-compiling and packaging on Debian 'testing'
## Please install the packages git and build-essential before getting the source via
## 'git clone --depth=1 https://github.com/nanu-c/axolotl/'
## and run 'make dependencies-deb-arm64(-cc)' once.

check-platform-deb-arm64:
ifneq ($(UNAME_S),Linux)
	@echo "Platform unsupported - only available for Linux" && exit 1
endif
ifneq ($(HARDWARE_PLATFORM),aarch64)
	@echo "Machine unsupported - only available for arm64/aarch64" && exit 1
endif
ifneq ($(APT),/usr/bin/apt)
	@echo "OS unsupported - apt not found" && exit 1
endif
ifneq ($(DEBIAN_VERSION),bookworm)
	@echo "Debian version not support - 'testing' is needed" && exit 1
endif

dependencies-deb-arm64: check-platform-deb-arm64
	@echo "Installing dependencies for building Axolotl on Debian 'testing' (bookworm)..."
	@sudo $(APT) update
	@sudo $(APT) install --assume-yes curl wget nodejs npm debmake
	@sudo $(APT) install --assume-yes --no-install-recommends libgtk-3-dev libjavascriptcoregtk-4.1-dev libsoup-3.0-dev libwebkit2gtk-4.1-dev protobuf-compiler
ifneq ($(RUST),${HOME}/.cargo/bin/rustup)
	@curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
endif
	@$(CARGO_PREFIX)/rustup update
	@echo "Dependencies installed."

build-deb-arm64: clean-deb-arm64
	@echo "Building Axolotl for arm64/aarch64 on Debian - Please use 'testing' release!)."
	@echo "Installing dependencies (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm ci
	@echo "Building (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm run build
	@echo "Building (rust)..."
	$(CARGO_PREFIX)/cargo build --features tauri --release
	@echo "Building complete."

prebuild-package-deb-arm64: package-clean-deb-arm64
	@echo "Prebuilding Debian package..."
# Get the source tarball
	@$(WGET) https://github.com/nanu-c/axolotl/archive/v$(AXOLOTL_VERSION).tar.gz --output-document=$(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION).tar.gz
# Prepare packaging folder
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl
	@cp $(CURRENT_DIR)/deb/LICENSE $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/LICENSE
# Run debmake
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debmake --yes --email arno_nuehm@riseup.net --fullname "Arno Nuehm" --monoarch
# Create target folders and copy additional files into package folder
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin
	@cp $(CURRENT_DIR)/README.md $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/README.Debian
	@cp $(CURRENT_DIR)/data/icons/axolotl.png $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps/axolotl.png
	@cp $(CURRENT_DIR)/deb/axolotl.desktop $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications/
	@cp $(CURRENT_DIR)/deb/axolotl.install $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/
	@cp $(CURRENT_DIR)/deb/control $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/control
	@cp $(CURRENT_DIR)/target/release/axolotl $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin/
	@echo "Prebuilding Debian package complete."

build-package-deb-arm64:
	@echo "Building Debian package..."
# Edit changelog file
	@sed -i '3d;4d' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
  @awk -i inplace 'NR == 3 {print "  * See upstream changelog below."} {print}' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
  @echo >> $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
  @cat $(CURRENT_DIR)/docs/CHANGELOG.md >> $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
# Edit copyright file
	@sed -i 's/<preferred name and address to reach the upstream project>/Aaron <aaron@nanu-c.org>/' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
  @sed -i 's/<url:\/\/example.com>/https:\/\/github.com\/nanu-c\/axolotl/' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
# Build Debian package
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debuild -i -us -uc -b

install-deb-arm64: uninstall-deb-arm64
# Use for testing purposes only
	@echo "Installing Axolotl..."
# Copy binary and helpers
	@sudo cp $(CURRENT_DIR)/target/release/build/axolotl $(DESTDIR)$(INSTALL_PREFIX)/
	@sudo cp $(CURRENT_DIR)/deb/axolotl.desktop $(DESTDIR)$(SHARE_PREFIX)/applications/
	@sudo cp $(CURRENT_DIR)/data/icons/axolotl.png $(DESTDIR)$(SHARE_PREFIX)/icons/hicolor/128x128/apps/
	@sudo update-icon-caches $(DESTDIR)$(SHARE_PREFIX)/icons/
	@echo "Installation complete."

uninstall-deb-arm64:
	@sudo rm --force $(DESTDIR)$(INSTALL_PREFIX)/axolotl
	@sudo rm --force $(DESTDIR)$(SHARE_PREFIX)/applications/axolotl.desktop
	@sudo rm --force $(DESTDIR)$(SHARE_PREFIX)/icons/hicolor/128x128/apps/axolotl.png
	@sudo update-icon-caches $(DESTDIR)$(SHARE_PREFIX)/icons/
	@echo "Removing complete."

## Cross-compiling via Makefile is not working properly at the moment!
check-platform-deb-arm64-cc:
ifneq ($(UNAME_S),Linux)
	@echo "Platform unsupported - only available for Linux" && exit 1
endif
ifneq ($(HARDWARE_PLATFORM),x86_64)
	@echo "Machine unsupported - x86_64 should be used" && exit 1
endif
ifneq ($(APT),/usr/bin/apt)
	@echo "OS unsupported - apt not found" && exit 1
endif
ifneq ($(DEBIAN_VERSION),bookworm)
	@echo "Debian version not support - 'testing' is needed" && exit 1
endif

dependencies-deb-arm64-cc: check-platform-deb-arm64-cc
	@echo "Installing dependencies for cross-compiling Axolotl... Be aware: This means Debian 'testing' (bookworm)!"
ifneq ($(DEBIAN_VERSION),bookworm)
	@echo "deb http://deb.debian.org/debian testing main contrib non-free" | sudo tee -a /etc/apt/sources.list
	@echo "deb-src http://deb.debian.org/debian testing main contrib non-free" | sudo tee -a /etc/apt/sources.list
	@sudo $(APT) update
	@sudo $(APT) --assume-yes upgrade
	@sudo $(APT) --assume-yes full-upgrade
endif
	@sudo $(APT) update
	@sudo dpkg --add-architecture arm64
	@sudo $(APT) install --assume-yes curl wget nodejs npm gcc-aarch64-linux-gnu linux-libc-dev-arm64-cross debmake
	@sudo $(APT) install --assume-yes --no-install-recommends libglib2.0-dev:arm64 libgtk-3-dev:arm64 libjavascriptcoregtk-4.1-dev:arm64 protobuf-compiler:arm64 libwebkit2gtk-4.1-dev:arm64 librsvg2-dev:arm64 libayatana-appindicator3-dev:arm64 libssl-dev:arm64 libjavascriptcoregtk-4.1-dev:arm64 g++ g++-aarch64-linux-gnu
ifneq ($(RUST),${HOME}/.cargo/bin/rustup)
	@curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
endif
	@$(CARGO_PREFIX)/rustup update
ifneq ($(CROSS),${HOME}/.cargo/bin/cross)
	@$(CARGO_PREFIX)/cargo install cross
endif
ifneq ($(DOCKER),/usr/bin/docker)
	@curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
	@echo "deb [arch=$(shell dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(DEBIAN_VERSION) stable" | sudo tee -a /etc/apt/sources.list.d/docker.list > /dev/null
	@sudo $(APT) update
	@sudo $(APT) install --assume-yes docker-ce docker-ce-cli containerd.io
	@sudo usermod -aG docker ${USER}
	@echo "Dependencies installed."
	@newgrp docker # This ends the current bash an starts a new one with docker added to groups.
endif
	@echo "Dependencies installed."

build-deb-arm64-cc: clean-deb-arm64
	@echo "Cross-compiling Axolotl for arm64/aarch64 on Debian 'testing'."
	@echo "Installing dependencies (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm --target_arch=arm64 ci
	@echo "Building (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm --target_arch=arm64 run build
	@echo "Building (rust)..."
	@sudo systemctl start docker
	@HOST_CC=gcc
	@CC_aarch64_unknown_linux_gnu=aarch64-linux-gnu-gcc
	@CXX_aarch64_unknown_linux_gnu=aarch64-linux-gnu-g++
	@CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
	@PKG_CONFIG_ALLOW_CROSS=1
	@PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig
	@PKG_CONFIG_SYSROOT_DIR=/
	@$(CARGO_PREFIX)/cross build --features tauri --release --target aarch64-unknown-linux-gnu
	@echo "Cross-compiling complete."

prebuild-package-deb-arm64-cc: package-clean-deb-arm64
	@echo "Prebuilding cross-compiled Debian package..."
# Get the source tarball
	@$(WGET) https://github.com/nanu-c/axolotl/archive/v$(AXOLOTL_VERSION).tar.gz --output-document=$(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION).tar.gz
# Prepare packaging folder
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/
	@cp $(CURRENT_DIR)/deb/LICENSE $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/LICENSE
# Run debmake
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debmake --yes --email arno_nuehm@riseup.net --fullname "Arno Nuehm" --monoarch
# Create target folders and copy additional files into package folder
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin
	@cp $(CURRENT_DIR)/README.md $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/README.Debian
	@cp $(CURRENT_DIR)/data/icons/axolotl.png $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps/axolotl.png
	@cp $(CURRENT_DIR)/deb/axolotl.desktop $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications/
	@cp $(CURRENT_DIR)/deb/axolotl.install $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/
	@cp $(CURRENT_DIR)/deb/control $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/control
	@cp $(CURRENT_DIR)/target/aarch64-unknown-linux-gnu/release/axolotl $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin/
	@echo "Prebuilding cross-compiled Debian package complete."

build-package-deb-arm64-cc:
	@echo "Building cross-compiled Debian package..."
# Edit changelog file
	@sed -i '3d;4d' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
  @awk -i inplace 'NR == 3 {print "  * See upstream changelog below."} {print}' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
  @echo >> $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
  @cat $(CURRENT_DIR)/docs/CHANGELOG.md >> $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
# Edit copyright file
	@sed -i 's/<preferred name and address to reach the upstream project>/Aaron <aaron@nanu-c.org>/' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
  @sed -i 's/<url:\/\/example.com>/https:\/\/github.com\/nanu-c\/axolotl/' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
# Build Debian package
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debuild -i -us -uc -b -aarm64

clean-deb-arm64:
	@rm --recursive --force $(CURRENT_DIR)/build/

package-clean-deb-arm64:
	@rm --recursive --force $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/

uninstall-deb-dependencies:
	@sudo apt purge curl wget git golang nodejs npm debmake
	@sudo apt autoremove && sudo apt autoclean
	@rustup self uninstall

uninstall-deb-dependencies-cc:
	@sudo apt purge curl wget git golang nodejs npm gcc-aarch64-linux-gnu debmake linux-libc-dev-arm64-cross docker-ce docker-ce-cli containerd.io libglib2.0-dev:arm64 libgtk-3-dev:arm64 libjavascriptcoregtk-4.1-dev:arm64 protobuf-compiler:arm64 libwebkit2gtk-4.1-dev:arm64 librsvg2-dev:arm64 libayatana-appindicator3-dev:arm64 libssl-dev:arm64 libjavascriptcoregtk-4.1-dev:arm64 g++ g++-aarch64-linux-gnu
	@sudo apt autoremove && sudo apt autoclean
	@rustup self uninstall
	@sudo rm /etc/apt/sources.list.d/docker.list
