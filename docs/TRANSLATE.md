# Translations

Axolotl uses gettext for translations. For every language a `.po` file exists in the apps `/po/` subfolder. These `.po` files need to be translated.

Instructions on how to translate using `.po` files are available here: http://docs.ubports.com/en/latest/contribute/translations.html#po-ts-file-editor

Once you finished translating, test the strings. This should be done before commiting any changes on axolotl-web.

The following dependencies are required and need to be installed:
```
sudo apt-get install gettext

```
Set up development environment as described under [README.md](README.md).

Change into the `axolotl-web` subfolder and run:
`npm run translate`

This command combines the following three single steps into one. Each of them can of course be run separately.
* `npm run translate-extract` extracting the language strings. This updates only the pot file.
* `npm run translate-update` for updating all the translation files.
* `npm run translate-compile` for updating the json file used by axolotl-web. Without that you don't see any results.

Then open Axolotl and have a look at your strings by either installing your package (click, snap, flatpak or appImage) or as described in [README.md](README.md) under "Run development".
