# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

.PHONY: build clean build-axolotl-web build-axolotl install install-axolotl install-axolotl-web uninstall build-translation run check check-axolotl check-axolotl-web build-dependencies build-dependencies-axolotl-web build-dependencies-axolotl update-version build-zkgroup copy-zkgroup install-zkgroup uninstall-zkgroup install-clickable-zkgroup uninstall-clickable-zkgroup build-dependencies-flatpak build-dependencies-flatpak-web build-dependencies-flatpak-qt install-flatpak-web install-flatpak-qt build-snap install-snap check-platform-deb-arm64 dependencies-deb-arm64 build-deb-arm64 prebuild-package-deb-arm64 build-package-deb-arm64 install-deb-arm64 uninstall-deb-arm64 check-platform-deb-arm64-cc dependencies-deb-arm64-cc build-deb-arm64-cc prebuild-package-deb-arm64-cc build-package-deb-arm64-cc clean-deb-arm64 package-clean-deb-arm64 uninstall-deb-dependencies-cc

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
APT=$(shell which apt)
WGET=$(shell which wget)
RUST=$(shell which rustup)
CROSS=$(shell which cross)
DOCKER=$(shell which docker)

DESTDIR = /
INSTALL_PREFIX = usr/bin
LIBRARY_PREFIX = usr/lib
SHARE_PREFIX = usr/share
CARGO_PREFIX = ${HOME}/.cargo/bin

all: clean build

build: build-axolotl-web build-axolotl build-crayfish build-zkgroup

install: install-axolotl install-axolotl-web install-crayfish install-zkgroup
	@sudo install -D -m 644 $(CURRENT_DIR)/scripts/axolotl.desktop $(DESTDIR)$(SHARE_PREFIX)/applications/axolotl.desktop
	@sudo install -D -m 644 $(CURRENT_DIR)/snap/gui/axolotl.png $(DESTDIR)$(SHARE_PREFIX)/icons/hicolor/128x128/apps/axolotl.png

uninstall: uninstall-axolotl uninstall-axolotl-web uninstall-crayfish uninstall-zkgroup

check: check-axolotl check-axolotl-web

build-dependencies: build-dependencies-axolotl-web build-dependencies-axolotl

# axolotl
build-dependencies-axolotl:
	$(GO) mod download

build-axolotl:
	@echo "Building axolotl..."
	$(GO) build -v .

check-axolotl:
	$(GO) test -race ./...

install-axolotl:
	@echo "Installing axolotl..."
	@install -D -m 755 $(CURRENT_DIR)/axolotl $(DESTDIR)$(INSTALL_PREFIX)/axolotl

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
	@mkdir -p $(DESTDIR)$(INSTALL_PREFIX)/axolotl-web/
	@cp -r $(CURRENT_DIR)/axolotl-web/dist $(DESTDIR)$(INSTALL_PREFIX)/axolotl-web/dist

uninstall-axolotl-web:
	@echo "Uninstalling axolotl-web..."
	@rm -rf $(DESTDIR)$(INSTALL_PREFIX)/axolotl-web

## utilities
build-translation:
	$(NPM) run translate --prefix axolotl-web

run: build
	@echo "Found go with version $(GO_VERSION)"
	LD_LIBRARY_PATH=$(PWD) $(GO)  run .

clean:
	rm -f $(CURRENT_DIR)/axolotl
	rm -rf $(CURRENT_DIR)/axolotl-web/dist
	rm -rf $(CURRENT_DIR)/crayfish/target

update-version:
ifeq ($(NEW_VERSION),)
	@echo 'Please specify the new version to use! Example: "make update-version NEW_VERSION=0.9.10"'
else
	@echo "Replacing current version $(AXOLOTL_VERSION) with new version $(NEW_VERSION)"
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' manifest.json
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' app/config/config.go
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' snap/snapcraft.yaml
	@sed -i "32i $$APPDATA_TEXT" appimage/AppDir/axolotl.appdata.xml
	@sed -i "32i $$APPDATA_TEXT" flatpak/org.nanuc.Axolotl.appdata.xml
	@echo "Update complete"
endif

## crayfish
build-crayfish:
	@echo "Building crayfish..."
	@cd $(CURRENT_DIR)/crayfish && cargo build --release

install-crayfish:
	@echo "Installing crayfish..."
	@install -D -m 755 $(CURRENT_DIR)/crayfish/target/release/crayfish $(DESTDIR)$(INSTALL_PREFIX)/

uninstall-crayfish:
	@echo "Uninstalling crayfish..."
	@rm -f $(DESTDIR)$(LIBRARY_PREFIX)/crayfish

## zkgroup
build-zkgroup:
	@echo "Building zkgroup..."
ifeq ($(UNAME_S), Linux)
	@echo "get zkgroup $(PLATFORM)"
	$(GO) get -d github.com/nanu-c/zkgroup \
	&& cd $(GOPATH)/src/github.com/nanu-c/zkgroup \
	&& $(GIT) submodule update \
	&& cd $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/zkgroup \
	&& $(CARGO) build --release --verbose \
	&& mv -f $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/zkgroup/target/release/libzkgroup.so $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so
else
	@echo "Architecture not (yet) supported $(HARDWARE_PLATFORM)"
	exit 1
endif

copy-zkgroup:
	$(GO) get -d github.com/nanu-c/zkgroup
	cp $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so $(CURRENT_DIR)/

install-zkgroup:
	@cp $(CURRENT_DIR)/libzkgroup_linux_$(HARDWARE_PLATFORM).so $(DESTDIR)$(LIBRARY_PREFIX)/

uninstall-zkgroup:
	@rm -f $(DESTDIR)$(LIBRARY_PREFIX)/libzkgroup_linux_$(HARDWARE_PLATFORM).so

install-clickable-zkgroup:
	$(GO) get -d github.com/nanu-c/zkgroup
	cp $(GOPATH)/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so $(CURRENT_DIR)/lib/

uninstall-clickable-zkgroup:
	rm -f $(CURRENT_DIR)/lib/libzkgroup_linux_$(HARDWARE_PLATFORM).so

## Flatpak
build-dependencies-flatpak:
	$(FLATPAK) install org.freedesktop.Sdk.Extension.golang//21.08
	$(FLATPAK) install org.freedesktop.Sdk.Extension.node16//21.08
	$(FLATPAK) install org.freedesktop.Sdk.Extension.rust-stable//21.08

build-dependencies-flatpak-web: build-dependencies-flatpak
	$(FLATPAK) install org.freedesktop.Platform//21.08
	$(FLATPAK) install org.freedesktop.Sdk//21.08
	$(FLATPAK) install org.electronjs.Electron2.BaseApp//21.08

build-dependencies-flatpak-qt: build-dependencies-flatpak
	$(FLATPAK) install org.kde.Platform//5.15
	$(FLATPAK) install org.kde.Sdk//5.15
	$(FLATPAK) install io.qt.qtwebengine.BaseApp//5.15

build-flatpak-web:
	$(FLATPAK_BUILDER) flatpak/build --verbose --force-clean flatpak/web/org.nanuc.Axolotl.yml

build-flatpak-qt:
	$(FLATPAK_BUILDER) flatpak/build --verbose --force-clean flatpak/qt/org.nanuc.Axolotl.yml

install-flatpak-web:
	$(FLATPAK_BUILDER) --user --install --force-clean flatpak/build flatpak/web/org.nanuc.Axolotl.yml

install-flatpak-qt:
	$(FLATPAK_BUILDER) --user --install --force-clean flatpak/build flatpak/qt/org.nanuc.Axolotl.yml

debug-flatpak-web:
	$(FLATPAK_BUILDER) --run --verbose flatpak/build flatpak/web/org.nanuc.Axolotl.yml sh

debug-flatpak-qt:
	$(FLATPAK_BUILDER) --run --verbose flatpak/build flatpak/qt/org.nanuc.Axolotl.yml sh

## Snap
build-snap:
	@sudo $(SNAPCRAFT)

install-snap:
	@sudo $(SNAP) install axolotl_$(AXOLOTL_VERSION)_amd64.snap --dangerous

## Debian arm64 building and packaging
## Please get the source via
## env GO111MODULE=off go get -d -u github.com/nanu-c/axolotl/
check-platform-deb-arm64:
	@echo "Building Axolotl for Debian arm64/aarch64."
  ifneq ($(UNAME_S),Linux)
	@echo "Platform unsupported - only available for Linux" && exit 1
  endif
  ifneq ($(HARDWARE_PLATFORM),aarch64)
	@echo "Machine unsupported - only available for arm64/aarch64" && exit 1
  endif
  ifneq ($(APT),/usr/bin/apt)
	@echo "OS unsupported - apt not found" && exit 1
  endif

dependencies-deb-arm64: check-platform-deb-arm64
	@echo "Installing dependencies for building Axolotl..."
	@sudo $(APT) update
	@sudo $(APT) install nano curl wget git golang nodejs npm debmake
  ifneq ($(RUST),${HOME}/.cargo/bin/rustup)
	@curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
  endif
	@$(CARGO_PREFIX)/rustup update
	@echo "Dependencies installed."

build-deb-arm64: dependencies-deb-arm64
	@echo "Downloading (go)..."
	@cd $(CURRENT_DIR) && go mod download
	@echo "Installing (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm ci
	@echo "Building (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm run build
	@mkdir -p $(CURRENT_DIR)/build/linux-arm64/axolotl-web
	@echo "Building (go)..."
	@cd $(CURRENT_DIR) && go build -o build/linux-arm64/axolotl .
	@cp --recursive $(CURRENT_DIR)/axolotl-web/dist $(CURRENT_DIR)/build/linux-arm64/axolotl-web/
	@cp --recursive $(CURRENT_DIR)/guis $(CURRENT_DIR)/build/linux-arm64/
	@echo "Building (rust)..."
	@cd $(CURRENT_DIR) && git submodule init && git submodule update
	@cd $(CURRENT_DIR)/crayfish && $(CARGO_PREFIX)/cargo build --release
	@echo "Building complete."

prebuild-package-deb-arm64: package-clean-deb-arm64
	@echo "Prebuilding Debian package..."
# Get the source tarball
	@$(WGET) https://github.com/nanu-c/axolotl/archive/v$(AXOLOTL_VERSION).tar.gz --output-document=$(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION).tar.gz
# Prepare packaging folder
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl
	@cp --recursive $(CURRENT_DIR)/build/linux-arm64/* $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl/
	@cp $(CURRENT_DIR)/deb/LICENSE $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/LICENSE
# Run debmake
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debmake --yes --email arno_nuehm@riseup.net --fullname "Arno Nuehm" --monoarch
# Create target folders and copy additional files into package folder
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/etc/profile.d
	@cp $(CURRENT_DIR)/README.md $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/README.Debian
	@cp $(CURRENT_DIR)/deb/axolotl.png $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps/axolotl.png
	@cp $(CURRENT_DIR)/deb/axolotl.desktop $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications/
	@cp $(CURRENT_DIR)/deb/axolotl.sh $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/etc/profile.d/
	@cp $(CURRENT_DIR)/deb/axolotl.install $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/
	@cp $(CURRENT_DIR)/deb/postinst $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/
	@cp $(CURRENT_DIR)/deb/control $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/control
	@$(WGET) https://github.com/nanu-c/zkgroup/raw/main/lib/libzkgroup_linux_aarch64.so --directory-prefix=$(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/lib/
	@mv $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl/axolotl $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin/
	@cp $(CURRENT_DIR)/crayfish/target/release/crayfish $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin/
	@echo "Prebuilding Debian package complete."

build-package-deb-arm64:
	@echo "Building Debian package..."
# Edit changelog file
	@sed -i '4d' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
	@sed -e '/Initial/ {' -e 'r $(CURRENT_DIR)/docs/CHANGELOG.md' -e 'd' -e '}' -i $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
	@sed -i '3,4d' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
	@sed -i 's/*/  */g' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
# Edit copyright file
	@sed -i 's/<preferred name and address to reach the upstream project>/aaron@nanu-c.org/' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
	@sed -i 's/<url:\/\/example.com>/https:\/\/github.com\/nanu-c\/axolotl/' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
# Build Debian package
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debuild -i -us -uc -b

install-deb-arm64: uninstall-deb-arm64
# Use for testing purposes only
	@echo "Installing Axolotl..."
# Copy libzkgroup
	@sudo $(WGET) https://github.com/nanu-c/zkgroup/raw/main/lib/libzkgroup_linux_aarch64.so --directory-prefix=$(DESTDIR)$(LIBRARY_PREFIX)/
#	Copy binary and helpers
	@sudo mkdir --parents $(DESTDIR)$(SHARE_PREFIX)/axolotl
	@sudo cp --recursive $(CURRENT_DIR)/build/linux-arm64/* $(DESTDIR)$(SHARE_PREFIX)/axolotl/
	@sudo mv $(DESTDIR)$(SHARE_PREFIX)/axolotl/axolotl $(DESTDIR)$(INSTALL_PREFIX)/
	@sudo cp $(CURRENT_DIR)/deb/axolotl.desktop $(DESTDIR)$(SHARE_PREFIX)/applications/
	@sudo cp $(CURRENT_DIR)/deb/axolotl.png $(DESTDIR)$(SHARE_PREFIX)/icons/hicolor/128x128/apps/
	@sudo xdg-icon-resource forceupdate
	@sudo cp $(CURRENT_DIR)/deb/axolotl.sh /etc/profile.d
	@bash -c "source /etc/profile.d/axolotl.sh"
	@sudo cp $(CURRENT_DIR)/crayfish/target/release/crayfish $(DESTDIR)$(INSTALL_PREFIX)/
	@echo "Installation complete."

uninstall-deb-arm64:
	@sudo rm --recursive --force $(DESTDIR)$(SHARE_PREFIX)/axolotl/
	@sudo rm --force $(DESTDIR)$(INSTALL_PREFIX)/axolotl
	@sudo rm --force $(DESTDIR)$(SHARE_PREFIX)/applications/axolotl.desktop
	@sudo rm --force $(DESTDIR)$(SHARE_PREFIX)/icons/hicolor/128x128/apps/axolotl.png
	@sudo xdg-icon-resource forceupdate
	@sudo rm --force /etc/profile.d/axolotl.sh
	@sudo rm --force $(DESTDIR)$(LIBRARY_PREFIX)/libzkgroup_linux_aarch64.so
	@sudo rm --force $(DESTDIR)$(INSTALL_PREFIX)/crayfish
	@echo "Removing complete."

check-platform-deb-arm64-cc:
	@echo "Cross-compiling Axolotl for Debian arm64/aarch64."
  ifneq ($(UNAME_S),Linux)
	@echo "Platform unsupported - only available for Linux" && exit 1
  endif
  ifneq ($(HARDWARE_PLATFORM),x86_64)
	@echo "Machine unsupported - x86_64 should be used" && exit 1
  endif
  ifneq ($(APT),/usr/bin/apt)
	@echo "OS unsupported - apt not found" && exit 1
  endif

dependencies-deb-arm64-cc: check-platform-deb-arm64-cc
	@echo "Installing dependencies for cross-compiling Axolotl..."
	@sudo dpkg --add-architecture arm64
	@sudo $(APT) update
	@sudo $(APT) install nano curl wget git golang nodejs npm gcc-aarch64-linux-gnu debmake libc-dev:arm64
  ifneq ($(RUST),${HOME}/.cargo/bin/rustup)
	@curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
  endif
	@$(CARGO_PREFIX)/rustup update
  ifneq ($(CROSS),${HOME}/.cargo/bin/cross)
	@$(CARGO_PREFIX)/cargo install cross
  endif
  ifneq ($(DOCKER),/usr/bin/docker)
	@curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
	@echo "deb [arch=$(shell dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(shell lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
	@sudo apt update
	@sudo apt install docker-ce docker-ce-cli containerd.io
	@sudo usermod -aG docker ${USER}
	@echo "Dependencies installed."
	@newgrp docker # This ends the current bash an starts a new one with docker added to groups.
  endif
	@echo "Dependencies installed."

build-deb-arm64-cc:
	@echo "Downloading (go)..."
	@cd $(CURRENT_DIR) && go mod download
	@echo "Installing (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm --target_arch=arm64 ci
	@echo "Building (npm)..."
	@cd $(CURRENT_DIR)/axolotl-web && npm --target_arch=arm64 run build
	@mkdir -p $(CURRENT_DIR)/build/linux-arm64/axolotl-web
	@echo "Building (go)..."
	@cd $(CURRENT_DIR) && env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig go build -o build/linux-arm64/axolotl .
	@cp --recursive $(CURRENT_DIR)/axolotl-web/dist $(CURRENT_DIR)/build/linux-arm64/axolotl-web/
	@cp --recursive $(CURRENT_DIR)/guis $(CURRENT_DIR)/build/linux-arm64/
	@echo "Building (rust)..."
	@sudo systemctl start docker
	@cd $(CURRENT_DIR) && git submodule init && git submodule update
	@cd $(CURRENT_DIR)/crayfish && $(CARGO_PREFIX)/cross build --release --target aarch64-unknown-linux-gnu
	@echo "Cross-compiling complete."

prebuild-package-deb-arm64-cc: package-clean-deb-arm64
	@echo "Prebuilding cross-compiled Debian package..."
# Get the source tarball
	@$(WGET) https://github.com/nanu-c/axolotl/archive/v$(AXOLOTL_VERSION).tar.gz --output-document=$(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION).tar.gz
# Prepare packaging folder
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl
	@cp --recursive $(CURRENT_DIR)/build/linux-arm64/* $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl/
	@cp $(CURRENT_DIR)/deb/LICENSE $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/LICENSE
# Run debmake
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debmake --yes --email arno_nuehm@riseup.net --fullname "Arno Nuehm" --monoarch
# Create target folders and copy additional files into package folder
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin
	@mkdir --parents $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/etc/profile.d
	@cp $(CURRENT_DIR)/README.md $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/README.Debian
	@cp $(CURRENT_DIR)/deb/axolotl.png $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/icons/hicolor/128x128/apps/axolotl.png
	@cp $(CURRENT_DIR)/deb/axolotl.desktop $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/share/applications/
	@cp $(CURRENT_DIR)/deb/axolotl.sh $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/etc/profile.d/
	@cp $(CURRENT_DIR)/deb/axolotl.install $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/
	@cp $(CURRENT_DIR)/deb/postinst $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/
	@cp $(CURRENT_DIR)/deb/control $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/control
	@$(WGET) https://github.com/nanu-c/zkgroup/raw/main/lib/libzkgroup_linux_aarch64.so --directory-prefix=$(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/lib/
	@mv $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/axolotl/axolotl $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin/
	@cp $(CURRENT_DIR)/crayfish/target/aarch64-unknown-linux-gnu/release/crayfish $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/usr/bin/
	@echo "Prebuilding cross-compiled Debian package complete."

build-package-deb-arm64-cc:
	@echo "Building cross-compiled Debian package..."
# Edit changelog file
	@sed -i '4d' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
	@sed -e '/Initial/ {' -e 'r $(CURRENT_DIR)/docs/CHANGELOG.md' -e 'd' -e '}' -i $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
	@sed -i '3,4d' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
	@sed -i 's/*/  */g' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/changelog
# Edit copyright file
	@sed -i 's/<preferred name and address to reach the upstream project>/aaron@nanu-c.org/' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
	@sed -i 's/<url:\/\/example.com>/https:\/\/github.com\/nanu-c\/axolotl/' $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/debian/copyright
# Build Debian package
	@cd $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION) && debuild -i -us -uc -b -aarm64

clean-deb-arm64:
	@rm --recursive --force $(CURRENT_DIR)/build/

package-clean-deb-arm64:
	@rm --recursive --force $(CURRENT_DIR)/axolotl-$(AXOLOTL_VERSION)/

uninstall-deb-dependencies:
	@sudo apt purge wget git golang nodejs npm debmake
	@sudo apt autoremove && sudo apt autoclean
	@rustup self uninstall

uninstall-deb-dependencies-cc:
	@sudo dpkg --remove-architecture arm64
	@sudo apt purge wget git golang nodejs npm gcc-aarch64-linux-gnu debmake libc-dev:arm64 docker-ce docker-ce-cli containerd.io
	@sudo apt autoremove && sudo apt autoclean
	@rustup self uninstall
	@sudo rm /etc/apt/sources.list.d/docker.list
