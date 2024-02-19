# axolotl-web

This is the frontend half of the axolotl project.
Axolotl is a complete cross-platform Signal client.

The Axolotl backend is running a web server, and with it serving the frontend bundle.

## Setup

This (sub)project is set up to support Node Version Manager (nvm).
To install, see [here](https://github.com/nvm-sh/nvm#installing-and-updating).

Once installed, the node and yarn version used by this project can be installed as follows.

```
nvm install
nvm use
```

Lastly, the dependencies needs to be downloaded.

```
yarn install
```

## Run

To start just the frontend, use the following command.

Note though, that the intended use of the frontend is generally to be started and used by the backend.

```
yarn run serve
```

## Build

To create the bundle, which the backend is serving, a bundle is required.
The bundle contains HTML, javascript and CSS - see `axolotl-web/dist` once finished.

```
yarn run build
```
