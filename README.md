# Axolotl

Axolotl is a complete cross-platform [Signal](https://www.signal.org) client, compatible with the Ubuntu Phone and more.
Unlike the desktop Signal client, **Axolotl is completely autonomous** and doesn't require you to have created an
account with the official Signal application.

It is built upon the go [textsecure package](https://github.com/nanu-c/textsecure/) and a Vue frontend that runs in an
electron/qml WebEngineView container.

<p align="center">
  <kbd>
    <img src="https://raw.githubusercontent.com/nanu-c/axolotl/main/screenshot.png" alt="Screenshot of axolotl" width="300px"/>
  </kbd>
</p>

## Features

 * Phone registration
 * Contact discovery
 * Direct messages
 * Group messages *mostly*
 * Photo, video, audio and contact attachments in both direct and group mode
 * Preview for photo and audio attachments
 * Storing conversations
 * Encrypted message store
 * Desktop client provisioning/syncing *partially*

### Planned

 * Push notifications
 * Most settings that are available in the Android app
 * Encrypted phone calls

There are still bugs and UI/UX quirks.

## Installation

Axolotl can be installed through different means.

| Package | Maintainer | Comment  |
| ------- | ---------- | -------- |
| <a href='https://open-store.io/app/textsecure.nanuc'><img width='130' alt="Get it from the OpenStore" src="https://open-store.io/badges/en_US.png"></a> | nanu-c | For Ubuntu Touch |
| <a href='https://snapcraft.io/axolotl'><img width='130' alt="Get it from the Snap Store" src="https://snapcraft.io/static/images/badges/en/snap-store-black.svg"></a> | nanu-c | For Ubuntu desktop |
| <a href='https://flathub.org/apps/details/org.nanuc.Axolotl'><img width='130' alt='Download on Flathub' src='https://flathub.org/assets/badges/flathub-badge-en.png'/></a> | olof-nord | https://github.com/flathub/org.nanuc.Axolotl |
| <a href='https://github.com/nuehm-arno/axolotl-mobian-package'><img alt="mobian version" src="https://img.shields.io/badge/axolotl-deb-%23A80030"></a> | nuehm-arno | https://github.com/nuehm-arno/axolotl-mobian-package |

## Building

To find out how to build from source and install yourself, please see below.

* with Clickable: see [here](docs/INSTALL.md#clickable).
* with Snap: see [here](docs/INSTALL.md#snap).
* with Flatpak: see [here](docs/INSTALL.md#flatpak).
* with AppImage: see [here](docs/INSTALL.md#appimage).
* for Mobian: see [here](docs/INSTALL.md#mobian).

### Manually

Requires Go, and node/npm.

If running Ubuntu, these steps should get you started.

First, install build dependencies.

```shell
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go nodejs npm build-essential
```

Then, install axolotl go and npm dependencies.

```shell
make build-dependencies
```

Now we are good to go. To start, simply use the following:

```shell
make run
```

When setting up for the first time and maybe occasionally later you need to update the browser list with your installed browsers.

- change into the `axolotl-web` subfolder
- run the following command: `npx browserslist@latest --update-db`

## Run flags

* `-axolotlWebDir` Specify the directory to use for axolotl-web. Defaults to "./axolotl-web/dist".
* `-e` for either
    `lorca`-> native chromium (has to be installed),
    `ut` -> runs in the ut enviroment,
    `me` -> qmlscene,
    `server` -> just run the webserver. Defaults to run with `electron`.
* `-eDebug` show developer console in electron mode
* `-version` Print version info
* `-host` Set the host to run the webserver from. Defaults to localhost.
* `-port` Set the port to run the webserver from. Defaults to 9080.

## Environment variables

* `AXOLOTL_WEB_DIR` Specify the directory to use for axolotl-web. This is used by `axolotl` during startup.
* `AXOLOTL_GUI_DIR` Specifies the directory used for GUI specifications. This is used by `axolotl` only when in `qt` mode.

## Contributing

* Please fill issues here on github https://github.com/nanu-c/axolotl/issues
* Help translate Axolotl to your language(s). For information how to translate, please see [TRANSLATE.md](docs/TRANSLATE.md).
* Contribute code by making PR's (pull requests)

If you contribute new strings, please:

- make them translatable using v-translate in the enclosing tag
- avoid linebreaks within one tag, that will break extracting the strings for translation
- try to reduce formatting tags within translatable strings

examples:

- `<p v-translate>Translate me!</p>` instead of `<p>Translate me!</p>`
- `<p><strong v-translate>Translate me!</strong></p>` instead of `<p v-translate><strong>Translate me!</strong></p>`
- `<p v-translate>Translate me!</p><br/><p v-translate> Please...</p>` instead of `<p v-translate>Translate me! <br/> Please...</p>`

## Migrating from `janimo/axolotl`

For information how to migrate from `janimo/axolotl`, please see [MIGRATE.md](docs/MIGRATE.md).
