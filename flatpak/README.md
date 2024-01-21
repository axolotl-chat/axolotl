# Flathub publishing

Flathub is the largest, de facto standard location for Flatpak software.

To publish an application there, a list of [App Requirements](https://github.com/flathub/flathub/wiki/App-Requirements)
do all need to be fulfilled.

One of these requirements is that of "Stable releases, reproducible builds".

To be published, all dependencies of the application needs to be listed in the Flatpak manifest.

There is a set of [flatpak builder tools](https://github.com/flatpak/flatpak-builder-tools) provided as to assist with
this dependency listing.

## axolotl-web

The big issue is for the `axolotl-web` dependencies.

Also, for this there is also a tool available,
[flatpak-node-generator](https://github.com/flatpak/flatpak-builder-tools/tree/master/node).
I have however never been able to successfully use it to parse the dependencies, and output it to the required list.

I suspect this is due to the complexity of the relations between the dependencies in node_modules, as I have waited for
several hours without any noticeable change.

Either way, to work around this, the dependencies for a specific version are bundled together and put in this repository.

### Create dependency archive

First, make sure to pull all the git tags.

```
git fetch --all --tags
```

Then check out the published tag. In our case, `v1.0.1`

```
git checkout tags/v1.0.1
```

Change to the axolotl-web directory, and make sure to use the npm version specified in the .nvmrc file.

```
cd axolotl-web/
nvm use
```

Then, from the axolotl-web directory, install all npm dependencies listed in
[package-lock.json](../axolotl-web/package-lock.json).
Note that `python` is required (!) for the node-sass installation to complete.

```
npm ci
```

Lastly, create the archive we want, naming it after the tag we checked out before.

```
tar cfvJ ../flatpak/archives/axolotl-web-dependencies-x86_64-v1.0.1.tar.xz node_modules
```

To verify, the archive can be extracted by using `tar xvJf axolotl-web-dependencies-x86_64-v1.0.1.tar.xz`.

## Flatpak details

To enter a shell into a built flatpak, use `make debug-flatpak-web`.

### Build directories

* axolotl-electron-bundle: `/run/build/axolotl-electron-bundle`

### Install directories

* axolotl-electron-bundle: `/app/bin/axolotl-electron-bundle`
