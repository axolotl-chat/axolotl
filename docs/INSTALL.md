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

The following translation dependencies are required:
```
sudo apt-get install gettext
```

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

This also transfers the click package to the Ubuntu Touch Phone.

`clickable`

**Run**

`clickable launch`

Clickable supports a few different parameters.
For example, `clickable launch logs` to start signal and get logging output.

For a full list of available clickable commands, see [here](https://clickable-ut.dev/en/latest/commands.html).

## Snap

**Tooling**

This requires `snap` and `snapcraft` to be installed locally.
Installation instructions for snapcraft can be found [here](https://snapcraft.io/docs/getting-started).

**Dependencies**

Snapcraft manages its own dependencies.

**Build and Install**

The Snap template used for the installation can be found
in the /snap subdirectory.

To build the application, use the following command from the root of this repository.

`sudo snapcraft`

To install the built snap, use snap:

`sudo snap install axolotl_0.9.8_amd64.snap --dangerous`

**Run**

To start the application, either search for "Axolotl" in your app drawer or start it with the below command.

`snap run axolotl`

## Flatpak

**Tooling**

This requires `flatpak` and `flatpak-builder` to be installed locally.
Installation instructions can be found [here](https://flatpak.org/setup/)

### Web Version

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

```flatpak-builder --user --install build ./flatpak/web/org.nanuc.Axolotl.yml```

System-wide:

Note that this requires root.

```sudo flatpak-builder --install build ./flatpak/web/org.nanuc.Axolotl.yml```

**Run**

To start the application, either search for "Axolotl" in your app drawer or start it with the below command.

`flatpak run org.nanuc.Axolotl`

### QT Version

**Dependencies**

The following Flatpak SDKs are required:
```
flatpak install org.kde.Platform//5.15
flatpak install org.kde.Sdk//5.15
flatpak install org.freedesktop.Sdk.Extension.golang//20.08
flatpak install org.freedesktop.Sdk.Extension.node12//20.08
flatpak install io.qt.qtwebengine.BaseApp//5.15
```

**Build and Install**

Installation can be done user-level or system-wide.
To list installed applications and/or runtimes, use `flatpak list`.

The Flatpak [manifest](https://docs.flatpak.org/en/latest/manifests.html) used for the installation can be found
in the /flatpak subdirectory.

User-level:

```flatpak-builder --user --install build ./flatpak/qt/org.nanuc.Axolotl.yml```

System-wide:

Note that this requires root.

```sudo flatpak-builder --install build ./flatpak/qt/org.nanuc.Axolotl.yml```

**Run**

To start the application, either search for "Axolotl" in your app drawer or start it with the below command.

`flatpak run org.nanuc.Axolotl -e=qt`

### Create a Flatpak bundle

Flatpak supports creating a [bundle](https://docs.flatpak.org/en/latest/single-file-bundles.html), which is a single
binary which can be used to distribute the application using removable media, or to send the application as an email
attachment.

To create a bundle, use the following steps.

**Dependencies**

During the build process, a gpg key is needed.
To generate one, install [gpg](https://www.gnupg.org/download/) and use it to generate a key (if you dont have one
already) with `gpg --gen-key`.

Then find and take note what your gpg key id is by looking for your key with `gpg --list-keys`.

**Build and Sign**

```
flatpak-builder --default-branch=main --disable-cache --force-clean --gpg-sign=mQINBFlD2sABEADsiUZUO... --repo=axolotl.repo axolotl.build ./flatpak/web/org.nanuc.Axolotl.yml
```

To then create the bundle, use the following.
Note that they should be executed from the same location, as the folder "axolotl.repo" is first generated, and then used.

```
flatpak build-bundle axolotl.repo axolotl.flatpak org.nanuc.Axolotl main --runtime-repo=https://flathub.org/repo/flathub.flatpakrepo
```

The end result is a binary file called `axolotl.flatpak`.

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
If needed, set the file as executable with `chmod +x Axolotl-x86_64.AppImage` first.

`./Axolotl-x86_64.AppImage`

## Build Axolotl for all arches with clickable and snap

This requires clickable and snapcraft to be installed.
It also requires the axolotl-web bundle to already be built.
see [build.sh](../scripts/build.sh)

## Mobian

If you want to run Axolotl on Mobian, you can download the installer, which is a simple script with all neccessary commands included. It can be found [here](../scripts/axolotl_installer_mobian_1-1.sh) - right click "Safe Link As...".
Execute it from your Download folder with

```
sh axolotl_installer_mobian.sh
```

To check out, what the script does exactly or to execute commands separately, visit the [Mobian wiki page for Axolotl](https://wiki.mobian-project.org/doku.php?id=axolotl) - section Manual Installation.
