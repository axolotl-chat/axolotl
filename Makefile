# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

NPM_VERSION := $(shell npm --version 2>/dev/null)
NODE_VERSION := $(shell node --version 2>/dev/null)
GO_VERSION := $(shell go version 2>/dev/null)
HARDWARE_PLATFORM := $(shell uname --machine)

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
	$(GO) run .

build-dependencies: build-dependencies-axolotl-web build-dependencies-axolotl

build-dependencies-axolotl-web:
	$(NPM) install --prefix axolotl-web

build-dependencies-axolotl:
	$(GO) mod download
ifeq ($(HARDWARE_PLATFORM), x86_64)
	wget https://raw.githubusercontent.com/nanu-c/zkgroup/main/lib/libzkgroup_linux_amd64.so
else ifeq ($(HARDWARE_PLATFORM), aarch64)
	wget https://raw.githubusercontent.com/nanu-c/zkgroup/main/lib/libzkgroup_linux_arm64.so
else ifeq ($(HARDWARE_PLATFORM), armhf)
	wget https://raw.githubusercontent.com/nanu-c/zkgroup/main/lib/libzkgroup_linux_armhf.so
else
	@echo architecture not supported
	exit 1
endif

clean:
	rm -f axolotl
	rm -rf axolotl-web/dist
