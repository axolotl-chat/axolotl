#!/bin/bash

# Translations
# generate pot file
echo "update translations"
find ../qml/* -iname "*.qml" | xargs xgettext --from-code utf-8 -C --qt --keyword=tr -o ../po/textsecure.nanuc.pot
find ../po/* -exec msgmerge --update {} ../po/textsecure.nanuc.pot \;
rm ../po/*~~

# mkdir -p ../build/tmp/qml
# cp -a ../qml/phoneui ../build/tmp/qml
# cp -a ../docs/CHANGELOG.md ../build/tmp
cp -a ../click/* ../build/install
# cp -a ../lib ../build/tmp
# if [ $mode = "dev" ];then
# 	#copy config.yml or rootCA.pem
# 	cp -a dev/* build/
# fi

# Build and include translations
cp ../po/textsecure.nanuc.pot ../build/tmp/
for po in ../po/*.po; do
	loc=$(echo $(basename $po)|cut -d'.' -f1)
	dir=../build/tmp/share/locale/$loc/LC_MESSAGES
	mkdir -p $dir
	msgfmt $po -o $dir/textsecure.nanuc.mo
done

# Build axolotl-web
cd ../axolotl-web&&npm run build && mkdir ../build/install/axolotl-web&&cp dist/* ../build/install/axolotl-web -r
