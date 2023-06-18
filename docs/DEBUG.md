# Debugging

This document is intended as a place to list some tips and tricks, mainly aimed for axolotl
development.

## Browser access

The Axolotl application can be accessed through the application window, but additionally
it is possible to also access the application through any browser.

To do so, just point a browser to `http://localhost:9080`.

## Run only frontend and connect to phone backend

Per default, both the Axolotl frontend and backend is started on the same device.

It is however also possible to connect the axolotl frontend to a backend running on another system,
for example a phone.

That way the Signal registration on your phone is used.

- `cd axolotl-web`
- `VITE_WS_ADDRESS=10.0.0.2 npm run serve` (replace 10.0.0.2 with the IP of your phone)
