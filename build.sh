#!/bin/bash

mode=${1:-dev}
echo "Building package in $mode mode."

rm -Rf builddir

rm textsecure.jani*click

cp -a click builddir

docker run --rm -it -v $(pwd):/home/developer -v $GOPATH:/home/developer/gopath -w $(pwd|sed "s,$GOPATH,/home/developer/gopath,") janimo/goqml-cross build -i -o builddir/textsecure
mkdir -p builddir/qml
cp -a qml/phoneui builddir/qml
if [ $mode = "dev" ];then
	#copy config.yml or rootCA.pem
	cp -a dev/* builddir/
fi

click build builddir
