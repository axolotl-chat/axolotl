# Installing

`axolotl` has a few different installation options in place.
Below is a list describing the tooling and dependencies required to use them.

**Note**: Be aware of the Crayfish Backend section if you are not using
Clickable.

## Checkout git submodules

Axolotl uses [crayfish](https://github.com/nanu-c/crayfish) in combination with [textsecure](https://github.com/signal-golang/textsecure) as backend to decipher the signal messages. In order to checkout the code run `git submodule update`.

## Clickable

**Tooling**

This requires `clickable` to be installed locally (version 7 or above).
Installation instructions can be found [here](https://clickable-ut.dev/en/dev/install.html).

**Build and Install**

**Note**: For the next three commands add `--arch <arch_of_your_mobile>` (i.e. `--arch arm64`) to the command when building for a mobile device.

1. In order to build axolotl you need to get its nodejs dependencies once:

    `clickable build --libs axolotlweb`

2. Finally the app is built by running:

    `clickable`

    This will build the app, install it onto a device connected via usb and run the app on the device.

    All steps can be done with individual clickable commands `clickable build`, `clickable install` and `clickable launch`. To build and run Axolotl on your pc run `clickable desktop`.

Clickable supports a few different parameters. Those can be set via command line or in the `clickable.yaml` file. For example run `clickable launch logs` to start signal and get logging output.

For a full list of available clickable commands, see [here](https://clickable-ut.dev/en/latest/commands.html).

## Native build

### Rust

Install Rust using [rustup](https://www.rust-lang.org/tools/install).

### Build Instructions

Build axolotl

```bash
make build
```

Building should work using both `stable` and `nightly` toolchains.


### Cross compile build

#### cross

To cross-compile for other targets, one approach is to use `cross` and specify the target flag.
[Cross](https://github.com/rust-embedded/cross) provides an environment, cross toolchain and cross
compiled libraries for building, without needing to install them separately.

To install, use `cargo install cross`.

To do a cross-compile build, use the following:

```bash
cross build --release --target aarch64-unknown-linux-gnu
cross build --release --target armv7-unknown-linux-gnueabihf
```

#### Natively

Another approach of cross-compiling is to set up the dependencies yourself.

For that two things are required. First install the required dependencies.
For Ubuntu, the following packages are required.

```bash
sudo apt install gcc-aarch64-linux-gnu gcc-arm-linux-gnueabihf
```

Then install the rust targets, e.g.:

```bash
rustup target add aarch64-unknown-linux-gnu
rustup target add armv7-unknown-linux-gnueabihf
```

Configure cargo with the cross-linker. For gcc:

```bash
export CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
export CARGO_TARGET_ARMV7_UNKNOWN_LINUX_GNUEABIHF_LINKER=armv7-unknown-linux-gnueabihf-gcc
```

To do a cross-compile build, use the following:

```bash
cargo build --release --target aarch64-unknown-linux-gnu
cargo build --release --target armv7-unknown-linux-gnueabihf
```

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

`sudo snap install axolotl_1.6.0_amd64.snap --dangerous`

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
flatpak install org.freedesktop.Platform//22.08
flatpak install org.freedesktop.Sdk//22.08
flatpak install org.freedesktop.Sdk.Extension.golang//22.08
flatpak install org.freedesktop.Sdk.Extension.node18//22.08
flatpak install org.freedesktop.Sdk.Extension.rust-stable//22.08
flatpak install org.electronjs.Electron2.BaseApp//22.08
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
flatpak install org.kde.Platform//5.15-22.08
flatpak install org.kde.Sdk//5.15-22.08
flatpak install org.freedesktop.Sdk.Extension.golang//22.08
flatpak install org.freedesktop.Sdk.Extension.node18//22.08
flatpak install org.freedesktop.Sdk.Extension.rust-stable//22.08
flatpak install io.qt.qtwebengine.BaseApp//5.15-22.08
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

## Mobian (or Debian arm64 systems)

**Build and Install**

Building Axolotl on Mobian (or other Debian arm64 systems) can be done by installing the building and packaging dependencies via
```
make dependencies-deb-arm64
```
and getting the source afterwards via

```
env GO111MODULE=off go get -d -u github.com/nanu-c/axolotl/
```
Executing the following Makefile command in the source folder will finally build Axolotl.
```
make build-deb-arm64
```

Installation can be done afterwards via
```
make install-deb-arm64
```

**Debian packaging**

The Debian arm64 packages [here](https://github.com/nanu-c/axolotl/releases) were fully autmatically created using the Github workflow.

Packaging is still under improvement to comply with official Debian packaging rules.

**Debian cross-compiling and packaging**

Axolotl can be cross-compiled and packaged for Mobian (Debian arm64 systems) on a Debian x86_64/amd64 system using
```
make dependencies-deb-arm64-cc
```
followed by
```
make build-deb-arm64-cc
```
and
```
make prebuild-package-deb-arm64-cc build-package-deb-arm64-cc
```
The resulting deb-file can be found in the source folder.


## Fedora

We are not activly supporting building and installing Axolotl nativly on Fedora but the following hint from [#502](https://github.com/nanu-c/axolotl/issues/502) might help.

```
GOVCS='*:all' does the trick though.
I also had to install the breezy package to get the bzr command.
```

## Bare

**Tooling**

This requires `make`, `go`, `nodejs`, `cargo`,  `rust` and `npm` to be installed locally.
For the required versions, see [go.mod](../go.mod) and [package.json](../axolotl-web/package.json)

**Build and Install**

To install, simply use the makefile target `install`.

`make install`
