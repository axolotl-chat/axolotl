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


define APPDATA_TEXT=
\\t\t\t<release version="$(NEW_VERSION)" date="$(shell date --rfc-3339='date')">\n\
\t\t\t\t\t<url>https://github.com/nanu-c/axolotl/releases/tag/v$(NEW_VERSION)</url>\n\
\t\t\t</release>
endef
export APPDATA_TEXT

NPM=$(shell which npm)
GO=$(shell which go)

all: clean build run

build: build-axolotl-web build-axolotl

build-axolotl-web:
	$(NPM) run build --prefix axolotl-web

build-axolotl:
	$(GO) build -v .

build-translation:
	$(NPM) run translate --prefix axolotl-web

run: build
		@echo "Found go with version $(GO_VERSION)"
		LD_LIBRARY_PATH=$(PWD) $(GO)  run .

build-dependencies: build-dependencies-axolotl-web build-dependencies-axolotl

build-dependencies-axolotl-web:
	$(NPM) install --prefix axolotl-web

build-dependencies-axolotl:
	$(GO) mod download

clean:
	rm -f axolotl
	rm -rf axolotl-web/dist

update-version:
ifeq ($(NEW_VERSION),)
	@echo Please specify the new version to use! Example: "make prepare-release NEW_VERSION=0.9.10"
else
	@echo Replacing current version $(AXOLOTL_VERSION) with new version $(NEW_VERSION)
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' manifest.json
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' app/config/config.go
	@sed -i 's/$(AXOLOTL_VERSION)/$(NEW_VERSION)/' snap/snapcraft.yaml
	@sed -i "32i $$APPDATA_TEXT" appimage/AppDir/axolotl.appdata.xml
	@sed -i "32i $$APPDATA_TEXT" flatpak/org.nanuc.Axolotl.appdata.xml
	@echo Update complete
endif

build-zkgroup:
	@echo "Found cargo with version $(CARGO_VERSION)"
	@echo "Found go with version $(GO_VERSION)"
	@echo "Found go with version $(GIT_VERSION)"
ifeq ($(UNAME_S), Linux)
ifeq ($(UNAME_HARDWARE_PLATTFORM), x86_64)
	@echo "get zkgroup $(PLATFORM)"
	go get -d github.com/nanu-c/zkgroup
	&& git submodule update \
	&& cd lib/zkgroup \
	&& cargo build --release --verbose
	mv libzkgroup.so libzkgroup_linux_amd64.so
else ifeq ($(UNAME_HARDWARE_PLATTFORM), aarch64)
	@echo "get zkgroup $(PLATFORM)"
	go get -d github.com/nanu-c/zkgroup
	&& git submodule update \
	&& cd lib/zkgroup \
	&& cargo build --release --verbose
	mv libzkgroup.so libzkgroup_linux_arm64.so
else
	@echo architecture not supported
	exit 1
endif
else
	@echo "platform not supported $(UNAME_S)"
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
ifeq ($(UNAME_HARDWARE_PLATTFORM), x86_64)
	@echo "get zkgroup $(PLATFORM)"
	go get -d github.com/nanu-c/zkgroup
else ifeq ($(UNAME_HARDWARE_PLATTFORM), aarch64)
	@echo "get zkgroup $(PLATFORM)"
	go get -d github.com/nanu-c/zkgroup
else
	@echo architecture not supported
	exit 1
endif
else
	@echo "platform not supported $(UNAME_S)"
	exit 1
endif

install-zkgroup:
ifeq ($(UNAME_S), Linux)
ifeq ($(UNAME_HARDWARE_PLATTFORM), x86_64)
	@echo "install libzkgroup to /usr/lib"
	cp ./libzkgroup_linux_amd64.so /usr/lib/
else ifeq ($(UNAME_HARDWARE_PLATTFORM), aarch64)
	@echo "install libzkgroup to /usr/lib"
	cp ./libzkgroup_linux_arm64.so  /usr/lib/

else
	@echo architecture not supported
	exit 1
endif
else
	@echo "platform not supported $(UNAME_S)"
	exit 1
endif
