# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

.PHONY: build build-axolotl-web build-axolotl build-translation run build-dependencies build-dependencies-axolotl-web build-dependencies-axolotl clean check_platform_arm64 dependencies_arm64 build_arm64 prebuild_package_arm64 build_package_arm64 install_arm64 uninstall_arm64 clean_arm64 package_clean_arm64

NPM_VERSION := $(shell npm --version 2>/dev/null)
NODE_VERSION := $(shell node --version 2>/dev/null)
GO_VERSION := $(shell go version 2>/dev/null)

NPM=$(shell which npm)
GO=$(shell which go)

GOPATH=$(shell go env GOPATH)
WORKDIR=$(GOPATH)/src/github.com/nanu-c/axolotl
VERSION=$(shell head -c 5 $(WORKDIR)/docs/CHANGELOG.md)

all: clean build run

build: build-axolotl-web build-axolotl

build-axolotl-web:
	$(NPM) run build --prefix axolotl-web

build-axolotl:
	$(GO) build -v .

build-translation:
	$(NPM) run translate --prefix axolotl-web

run: build
	$(GO) run .

build-dependencies: build-dependencies-axolotl-web build-dependencies-axolotl

build-dependencies-axolotl-web:
	$(NPM) install --prefix axolotl-web

build-dependencies-axolotl:
	$(GO) mod download

clean:
	rm -f axolotl
	rm -rf axolotl-web/dist

check_platform_arm64:
	@echo "Building Axolotl for Mobian (aarch64/amd64)"
  ifneq ($(shell uname),Linux)
	@echo "Platform unsupported - only available for Linux" && exit 1
  endif
  ifneq ($(shell uname -m),aarch64)
	@echo "Machine unsupported - only available for aarch64/arm64" && exit 1
  endif
  ifneq ($(shell which apt),/usr/bin/apt)
	@echo "OS unsupported - apt not found" && exit 1
  endif

dependencies_arm64:
	@echo "Installing dependencies for building Axolotl..."
	@sudo apt update
	@sudo apt install golang nodejs npm python

build_arm64:
	@echo "Downloading (go)..."
	@cd $(WORKDIR) && go mod download
	@echo "Installing (npm)..."
	@cd $(WORKDIR)/axolotl-web && npm install
	@echo "node-sass does not support aarch64/amd64 so it has to be rebuilt"
	@echo "Rebuilding of npm-sass..."
	@cd $(WORKDIR)/axolotl-web && npm rebuild node-sass
	@echo "Building (npm)..."
	@cd $(WORKDIR)/axolotl-web && npm run build
	@mkdir -p $(WORKDIR)/build/linux-arm64/axolotl-web
	@echo "Building (go)..."
	@cd $(WORKDIR) && env GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/axolotl .
	@cp -r axolotl-web/dist build/linux-arm64/axolotl-web
	@cp -r guis build/linux-arm64
	@echo "Building complete."

prebuild_package_arm64: package_clean_arm64
	@echo "Prebuilding Debian package..."
# Get the source tarball
	@cd $(WORKDIR) && wget https://github.com/nanu-c/axolotl/archive/main.tar.gz
# Rename source tarball
	@mv $(WORKDIR)/main.tar.gz $(WORKDIR)/axolotl-$(VERSION).tar.gz
# Prepare packaging folder
	@mkdir -p $(WORKDIR)/axolotl-$(VERSION)/axolotl
	@cp -r $(WORKDIR)/build/linux-arm64/* $(WORKDIR)/axolotl-$(VERSION)/axolotl
	@cp $(WORKDIR)/LICENSE $(WORKDIR)/axolotl-$(VERSION)/LICENSE
# Run debmake
	@cd $(WORKDIR)/axolotl-$(VERSION) && debmake -e arno_nuehm@riseup.net -f "Arno Nuehm" -m
# Copy additional files in package folder
	@cp $(WORKDIR)/README.md $(WORKDIR)/axolotl-$(VERSION)/debian/README.Debian
	@mkdir -p $(WORKDIR)/axolotl-$(VERSION)/usr/share/icons/hicolor/128x128/apps
	@cp $(WORKDIR)/axolotl-$(VERSION)/axolotl/axolotl-web/dist/axolotl.png $(WORKDIR)/axolotl-$(VERSION)/usr/share/icons/hicolor/128x128/apps/axolotl.png
	@mkdir -p $(WORKDIR)/axolotl-$(VERSION)/usr/share/applications
	@cp $(WORKDIR)/deb/axolotl.desktop $(WORKDIR)/axolotl-$(VERSION)/usr/share/applications
	@cp $(WORKDIR)/deb/axolotl.install $(WORKDIR)/axolotl-$(VERSION)/debian
	@cp $(WORKDIR)/deb/postinst $(WORKDIR)/axolotl-$(VERSION)/debian
	@cp $(WORKDIR)/deb/postrm $(WORKDIR)/axolotl-$(VERSION)/debian
	@cp $(WORKDIR)/deb/control $(WORKDIR)/axolotl-$(VERSION)/debian/control

build_package_arm64:
	@echo "Building Debian package..."
# Prompt to edit changelog file
	@nano $(WORKDIR)/axolotl-$(VERSION)/debian/changelog
# Prompt to edit copyright file
	@nano $(WORKDIR)/axolotl-$(VERSION)/debian/copyright
# Build Debian package
	@cd $(WORKDIR)/axolotl-$(VERSION) && debuild -i -us -uc -b

install_arm64: uninstall_arm64
# Use for testing purposes only
	@sudo mkdir -p /usr/share/axolotl
	@sudo cp -r $(WORKDIR)/axolotl-$(VERSION)/axolotl/* /usr/share/axolotl
	@sudo ln -sf /usr/share/axolotl/axolotl /usr/bin/axolotl
	@sudo cp $(WORKDIR)/deb/axolotl.desktop /usr/share/applications
	@sudo cp $(WORKDIR)/axolotl-$(VERSION)/axolotl/axolotl-web/dist/axolotl.png /usr/share/icons/hicolor/128x128/apps

uninstall_arm64:
	@sudo rm -rf /usr/share/axolotl
	@sudo rm -f /usr/bin/axolotl
	@sudo rm -f /usr/share/applications/axolotl.desktop
	@sudo rm -f /usr/share/icons/hicolor/128x128/apps/axolotl.png

clean_arm64:
	@rm -rf $(WORKDIR)/build

package_clean_arm64:
	@rm -rf $(WORKDIR)/axolotl-$(VERSION)
