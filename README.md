# Axolotl is a crossplattform [Signal](https://www.signal.org) client

## For the Ubuntu Phone and more

Axolotl is a complete Signal client, it allows you to create a Signal account and have discussions with your contacts.
Unlike the desktop Signal client, **Axolotl is completely autonomous** and doesn't require you to have created an account with the official Signal application.

It is built upon the [Go textsecure package](https://github.com/nanu-c/textsecure/) and a vuejs app that runs in a electron/qml WebEngineView container.

To use it from your Ubuntu Touch device, simply install it from the open store:  
[![OpenStore](https://open-store.io/badges/en_US.png)](https://open-store.io/app/textsecure.nanuc)

Axolotl is also available as a snap package, to install it on Ubuntu desktop:  
[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/axolotl)

What works
-----------

 * Phone registration
 * Contact discovery
 * Direct messages
 * Group messages *mostly*
 * Photo, video, audio and contact attachments in both direct and group mode
 * Preview for photo and audio attachments
 * Storing conversations
 * Encrypted message store
 * Desktop client provisioning/syncing *partially*

What is missing
---------------

 * Push notifications
 * Most settings that are available in the Android app
 * Encrypted phone calls

There are still bugs and UI/UX quirks.

Installation of development environment
------------
* Install [Golang](https://golang.org/doc/install)
* Install node js (see the [.nvmrc](axolotl-web/.nvmrc)) file for the supported version
* Add gopath to ~/.bashrc https://github.com/golang/go/wiki/SettingGOPATH
* install dependencies `sudo apt install mercurial gettext`
* Check out this git `go get -d github.com/nanu-c/axolotl`
* `cd $(go env GOPATH)/src/github.com/nanu-c/axolotl`
* get go dependencies `go mod download`
* install axolotl-web dependencies: `cd axolotl-web&&npm install`

When setting up for the first time and maybe occasionally later you need to update the browser list with your installed browsers. Change into the axolotl-web subfolder and run the following command:

`npx browserslist@latest --update-db`

Run development
------------
* `cd $(go env GOPATH)/src/github.com/nanu-c/axolotl`
* `go run .`
* in a new terminal `cd axolotl-web&&npm run serve`
* point a browser to the link printed in the terminal  like `http://localhost:9080`

Run frontend and connect to phone ip
--------------
That way running the backend is avoided, instead your current registration on ubuntu touch is used
* `cd axolotl-web`
* `VUE_APP_WS_ADDRESS=10.0.0.2 npm run serve` replace 10.0.0.2 with the ip of your phone

Installation
------------
Axolotl can be built and installed in different ways.

To find out how to build and install, please see below:

* with Clickable: see [here](docs/INSTALL.md#clickable).
* with Snap: see [here](docs/INSTALL.md#snap).
* with Flatpak: see [here](docs/INSTALL.md#flatpak).
* with AppImage: see [here](docs/INSTALL.md#appimage).
* for Mobian: see [here](docs/INSTALL.md#mobian).


Run flags
-----------
* `-axolotlWebDir` Specify the directory to use for axolotl-web. Defaults to "./axolotl-web/dist".
* `-e` for either
    `lorca`-> native chromium (has to be installed),
    `ut` -> runs in the ut enviroment,
    `me` -> qmlscene,
    `server` -> just run the webserver. Defaults to run with `electron`.
* `-eDebug` show developer console in electron mode
* `-host` Set the host to run the webserver from. Defaults to localhost.
* `-port` Set the port to run the webserver from. Defaults to 9080.

Environment variables
-----------
* `AXOLOTL_WEB_DIR` Specify the directory to use for axolotl-web. This is used by `axolotl` during startup.
* `AXOLOTL_GUI_DIR` Specifies the directory used for GUI specifications. This is used by `axolotl` only when in `qt` mode.

Contributing
-----------
* Please fill issues here on github https://github.com/nanu-c/axolotl/issues
* Help translating Axolotl to your language(s). For information how to translate, please see [TRANSLATE.md](docs/TRANSLATE.md).
* Contribute code by making PR's (pull requests)

If you contribute new strings, please:

- make them translatable using v-translate in the enclosing tag
- avoid linebreaks within one tag, that will break extracting the strings for translation
- try to reduce formatting tags within translatable strings

examples:

- `<p v-translate>Translate me!</p>` instead of `<p>Translate me!</p>`
- `<p><strong v-translate>Translate me!</strong></p>` instead of `<p v-translate><strong>Translate me!</strong></p>`
- `<p v-translate>Translate me!</p><br/><p v-translate> Please...</p>` instead of `<p v-translate>Translate me! <br/> Please...</p>`

Migrating from `janimo/axolotl`
--------------------------------------

For information how to migrate from `janimo/axolotl`, please see [MIGRATE.md](docs/MIGRATE.md).
