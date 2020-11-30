# Translations

Axolotl uses gettext for translations. Use the `.po` files in `/po/` for translations.
Instructions on how to translate using `.po` files are available here: http://docs.ubports.com/en/latest/contribute/translations.html#po-ts-file-editor

The following translation dependencies are required:
```
sudo apt-get install gettext
```

Once you finished translating, test the strings. For testing set up  development enviroment as above and run
* `npm run translate-extract` extracting the language strings. This updates only the pot file
* `npm run translate-update` for updating all the translation files
* `npm run translate-compile` for updating the json file used by axolotl-web. Without that you don't see any results
* or `npm run translate` for all of the 3 commands at the same time. This should be run befor commiting any changes on axolotl-web
