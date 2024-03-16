# Installing

Build the frontend first. See [instructions](./axolotl-web/README.md).

## Clickable

### Tooling

This requires `clickable` to be installed locally (version 7 or above).
Installation instructions can be found [here](https://clickable-ut.dev/en/dev/install.html).

### Build and Install

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

### Tooling

This requires `snap` and `snapcraft` to be installed locally.
Installation instructions for snapcraft can be found [here](https://snapcraft.io/docs/getting-started).

### Dependencies

Snapcraft manages its own dependencies.

### Build and Install

The Snap template used for the installation can be found
in the /snap subdirectory.

To build the application, use the following command from the root of this repository.

`sudo snapcraft`

To install the built snap, use snap:

`sudo snap install axolotl_1.6.0_amd64.snap --dangerous`

### Run

To start the application, either search for "Axolotl" in your app drawer or start it with the below command.

`snap run axolotl`

## Flatpak

### Tooling

This requires `flatpak` and `flatpak-builder` to be installed locally.
Installation instructions can be found [here](https://flatpak.org/setup/)

### Dependencies

The following Flatpak SDKs are required:
```
flatpak install install org.gnome.Platform//45
flatpak install install org.gnome.Sdk//45
flatpak install install org.freedesktop.Sdk.Extension.node18//22.08
flatpak install install org.freedesktop.Sdk.Extension.rust-stable//22.08
```

### Build and Install

```
cd flatpak
flatpak-builder build org.nanuc.Axolotl.yml --force-clean --keep-build-dirs --ccache --user --install
```

### Run

To start the application, either search for "Axolotl" in your app drawer or start it with the below command.

`flatpak run org.nanuc.Axolotl --mode tauri`

## AppImage

### Tooling

This requires `yarn`, `cargo` and `tauri-cli` to be installed locally.

### Build

```
cargo tauri build --features tauri --bundles appimage
```

### Run

To start the application, execute the AppImage binary directly

`target/release/bundle/appimage/*.AppImage`

## Debian

### Tooling

This requires `yarn`, `cargo` and `tauri-cli` to be installed locally.

### Build and Install

```
cargo tauri build --features tauri --bundles deb
sudo apt install ./target/release/bundle/deb/*.deb
```

## Bare

### Tooling

This requires `make` and `cargo` to be installed locally.

### Build and Install

```
make install
```
