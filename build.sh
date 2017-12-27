#!/bin/bash

mode=${1:-dev}
echo "Building package in $mode mode."

rm -Rf build

rm textsecure.nanu*click

cp -a click build
#sudo docker run --rm -it -v $(pwd):/home/developer -v /home/nanu/src/go:/home/developer/gopath -w /home/developer nanu-c/goqml-cross get github.com/amlwwalker/qml
sudo docker run --rm -it -v $(pwd):/home/developer -v $GOPATH:/home/developer/gopath -w /home/developer nanu-c/goqml-cross build -i -o build/textsecure .
mkdir -p build/qml
cp -a qml/phoneui build/qml
cp -a CHANGELOG.md build
# if [ $mode = "dev" ];then
# 	#copy config.yml or rootCA.pem
# 	cp -a dev/* build/
# fi

# Build and include translations
for po in po/*.po; do
	loc=$(echo $(basename $po)|cut -d'.' -f1)
	dir=build/share/locale/$loc/LC_MESSAGES
	mkdir -p $dir
	msgfmt $po -o $dir/textsecure.jani.mo
done

click build build
