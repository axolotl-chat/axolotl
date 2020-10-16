# Installing

`axolotl` has a few different installation options in place.
Below is a list describing the tooling and dependencies required to use them.

## Clickable

**Tooling**

This requires `clickable` to be installed locally.
Installation instructions can be found [here](https://clickable-ut.dev/en/latest/install.html#install).

**Dependencies**

The following build dependencies are required:
* Docker
* Go

The following go-qml dependencies are required:
```
sudo add-apt-repository ppa:ubuntu-sdk-team/ppa
sudo apt-get update
sudo apt-get install qtdeclarative5-dev qtbase5-private-dev qtdeclarative5-private-dev libqt5opengl5-dev qtdeclarative5-qtquick2-plugin
sudo ln -s /usr/include/x86_64-linux-gnu/qt5/QtCore/5.9.1/QtCore /usr/include/
```

To install all go dependencies, use `go mod download`.

**Build and Install**

To run the default set of sub-commands, simply run clickable in the root directory.
Clickable will attempt to auto detect the build template and other configuration options.

`clickable`

**Run**

`clickable launch`

For a full list of available clickable commands, see [here](https://clickable-ut.dev/en/latest/commands.html).

## Snap

**Tooling**

This requires `snapcraft` to be installed locally.
Installation instructions can be found [here](https://snapcraft.io/docs/getting-started).

**Dependencies**

TODO: Add more info

**Build and Install**

The Snap template used for the installation can be found
in the /snap subdirectory.

TODO: Add more info

**Run**

TODO: Add more info

## Flatpak

**Tooling**

This requires `flatpak` and `flatpak-builder` to be installed locally.
Installation instructions can be found [here](https://flatpak.org/setup/)

**Dependencies**

The following Flatpak SDKs are required:
```
flatpak install org.freedesktop.Platform//20.08
flatpak install org.freedesktop.Sdk//20.08
flatpak install org.freedesktop.Sdk.Extension.golang//20.08
flatpak install org.freedesktop.Sdk.Extension.node12//20.08
```

**Build and Install**

Installation can be done user-level or system-wide.
To list installed applications and/or runtimes, use `flatpak list`.

The Flatpak [manifest](https://docs.flatpak.org/en/latest/manifests.html) used for the installation can be found
in the /flatpak subdirectory.

User-level: 

```flatpak-builder --user --install build ./flatpak/org.nanuc.Axolotl.yml```

System-wide:

Note that this requires root.

```sudo flatpak-builder --install build ./flatpak/org.nanuc.Axolotl.yml```

**Run**

To start the application, either search for "Axolotl" in your app drawer or start it with the below command.

`flatpak run org.nanuc.Axolotl`

## AppImage

**Tooling**

This requires `appimagetool`, `go` and `npm` to be installed locally.
Installation instructions for `appimagetool` can be found [here](https://github.com/AppImage/AppImageKit#appimagetool-usage)

**Build and Install**

AppImage does not really have a concept of install, just execute the build script to compile and put all files in place.
The build files are stored in /build/AppDir.

```
cd appimage
./build.sh
```

**Run**

To start the application, execute the AppImage binary directly:

`./Axolotl-x86_64.AppImage`
