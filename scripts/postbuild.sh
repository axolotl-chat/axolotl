#!/bin/bash

# Translations
# generate pot file
echo "update translations $@"
find ../../../qml/* -iname "*.qml" | xargs xgettext --from-code utf-8 -C --qt --keyword=tr -o ../../../po/textsecure.nanuc.pot
find ../../../po/* -exec msgmerge --update {} ../../../po/textsecure.nanuc.pot \;
# rm ../../../po/*~~
cp -a ../../../click/* $@
# cp -a ../lib ../build/tmp
# if [ $mode = "dev" ];then
# 	#copy config.yml or rootCA.pem
# 	cp -a dev/* build/
# fi

# Build and include translations
cp ../../../po/textsecure.nanuc.pot $@
for po in ../../../po/*.po; do
	loc=$(echo $(basename $po)|cut -d'.' -f1)
	dir=$@/share/locale/$loc/LC_MESSAGES
	mkdir -p $dir
	msgfmt $po -o $dir/textsecure.nanuc.mo
done
# Build axolotl-web
$a = pwd
cd ../../../axolotl-web&&npm run build && mkdir $@/axolotl-web&&ls -lah dist &&cp dist $@/axolotl-web/ -r && ls -lah $@
# cd ../../../axolotl-web&&npm run build && mkdir $@/axolotl-web&&rm -rf $@/axolotl-web/dist &&cp dist $@/axolotl-web/dist -r
cd $a
