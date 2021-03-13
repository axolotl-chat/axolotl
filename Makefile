# This is the Makefile for axolotl.
# For more info about the syntax, see https://makefiletutorial.com/

NPM=$(shell which npm)
GO=$(shell which go)

.PHONY: all
all: clean build run

.PHONY: build
build: build-axolotl-web build-axolotl

.PHONY: build-axolotl-web
build-axolotl-web:
	$(NPM) run build --prefix axolotl-web

.PHONY: build-axolotl
build-axolotl:
	$(GO) build -v .

.PHONY: run
run: build
	$(GO) run .

.PHONY: build-dependencies
build-dependencies: build-dependencies-axolotl-web build-dependencies-axolotl

.PHONY: build-dependencies-axolotl-web
build-dependencies-axolotl-web:
	$(NPM) install --prefix axolotl-web

.PHONY: build-dependencies-axolotl
build-dependencies-axolotl:
	$(GO) mod download

.PHONY: clean
clean:
	rm -f axolotl
	rm -rf axolotl-web/dist
