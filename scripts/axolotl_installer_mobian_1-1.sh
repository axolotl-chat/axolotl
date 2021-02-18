#!/bin/sh
# This script installs the latest version of Axolotl from https://github.com/nanu-c/axolotl and should only be used on Mobian devices by "sh axolotl_installer_mobian_1-1.sh". Please do a restart before executing the script. The installation will take up to 45 min. And please be patient...
# Created by arno_nuehm
# Version 1.1 - 10.02.2021
# Inspired by https://wiki.mobian-project.org/doku.php?id=axolotl

echo "This script installs the latest version of Axolotl from https://github.com/nanu-c/axolotl. Please do a restart before executing the script. The installation will take up to 45 min. And please be patient..."
echo "installing dependencies..."
sudo apt-get update && sudo apt-get install golang nodejs npm mercurial python qmlscene qml-module-qtwebsockets qml-module-qtmultimedia qml-module-qtwebengine
#the following qml modules have to be installed separately (regex issue)
sudo apt-get install qml-module-qtquick.controls
sudo apt-get install qml-module-qtquick.dialogs
echo "cloning..."
go get -d github.com/nanu-c/axolotl
cd $(go env GOPATH)/src/github.com/nanu-c/axolotl && go mod download
echo "installing..."
cd axolotl-web && npm install
#node-sass does not support arm64 so we have to rebuild it
echo "rebuilding of npm-sass..."
npm rebuild node-sass
echo "building (npm)..."
npm run build
cd .. && mkdir -p build/linux-arm64/axolotl-web
echo "building (go)..."
env GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/axolotl .
cp -r axolotl-web/dist build/linux-arm64/axolotl-web && cp -r guis build/linux-arm64
echo "[Desktop Entry]\nType=Application\nName=Axolotl\nGenericName=Signal Chat Client\nPath=/home/mobian/go/src/github.com/nanu-c/axolotl/build/linux-arm64/\nExec=/home/mobian/go/src/github.com/nanu-c/axolotl/build/linux-arm64/axolotl\n#Exec=/home/mobian/go/src/github.com/nanu-c/axolotl/build/linux-arm64/axolotl -e qt\nIcon=/home/mobian/go/src/github.com/nanu-c/axolotl/build/linux-arm64/axolotl-web/dist/axolotl.png\nTerminal=false\nCategories=Network;Chat;InstantMessaging;Qt;\nStartupWMClass=axolotl" | sudo tee -a /usr/share/applications/axolotl.desktop
echo "Congratulations! You should now see an Axolotl smiling in your app menu."
