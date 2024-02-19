# Flathub publishing

Flathub is the largest, de facto standard location for Flatpak software.

To publish an application there, a list of [App Requirements](https://github.com/flathub/flathub/wiki/App-Requirements)
do all need to be fulfilled.

One of these requirements is that of "Stable releases, reproducible builds".

## Dependencies

To be published, all dependencies of the application needs to be listed in the Flatpak manifest.

There is a set of [flatpak builder tools](https://github.com/flatpak/flatpak-builder-tools) provided as to assist with
this dependency listing.

### axolotl-web

Generate npm/yarn dependencies via [flatpak-node-generator](https://github.com/flatpak/flatpak-builder-tools/tree/master/node):

```sh
sudo apt install pipx python3
git clone git@github.com:flatpak/flatpak-builder-tools.git
cd flatpak-builder-tools/node
pipx install .
flatpak-node-generator yarn ../../axolotl-web/yarn.lock -o ../../flatpak/node-sources.json
```

### axolotl

Generate cargo dependencies via [flatpak-cargo-generator](https://github.com/flatpak/flatpak-builder-tools/tree/master/cargo):

```sh
sudo apt install python3 python3-aiohttp python3-toml python3-yaml
git clone git@github.com:flatpak/flatpak-builder-tools.git
cd flatpak-builder-tools/cargo
python3 ./flatpak-cargo-generator.py ../../Cargo.lock -o ../../flatpak/cargo-sources.json
```
