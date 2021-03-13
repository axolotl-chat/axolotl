# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

EXECUTABLES = go npm node
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),,$(error "No $(exec) in PATH")))

NPM=$(shell which npm)
GO=$(shell which go)

.PHONY: all
all: clean build run

.PHONY: build
build:
	$(GO) build -v .
	$(NPM) run build --prefix axolotl-web

.PHONY: run
run: build
	$(GO) run .

.PHONY: build-deps
build-deps:
	$(GO) mod download
	$(NPM) install --prefix axolotl-web

.PHONY: clean
clean:
	rm -f axolotl
	rm -rf axolotl-web/dist
