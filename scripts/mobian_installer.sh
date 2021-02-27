#!/bin/sh
# This script installs or updates Axolotl to the latest version from https://github.com/nanu-c/axolotl and should only be used on Mobian devices by "sh mobian_installer.sh". Please restart before executing. The update is quite fast but the first installation will take up to 45 min. So please be patient... And disable 'Settings - Power - Automatic Suspend' for that period of time.
# Created by arno_nuehm
# Version 1.5 - 19.02.2021
# Inspired by https://wiki.mobian-project.org/doku.php?id=axolotl

echo "This script installs or updates Axolotl\nto the latest version from\nhttps://github.com/nanu-c/axolotl.\nPlease restart before executing.\nThe update is quite fast but the first\ninstallation will take up to 45 min.\nSo please be patient...\nAnd disable\n 'Settings - Power - Automatic Suspend'\nfor that period of time."
echo -n "Do you want to continue? (y/n) --> "
read yesno < /dev/tty
if [ "x$yesno" = "xy" ];
then
  :
else
  echo "Aborting..." & exit 1
fi
echo "Installing dependencies..."
sudo apt update && sudo apt install golang nodejs npm mercurial python qmlscene qml-module-qtwebsockets qml-module-qtmultimedia qml-module-qtwebengine || {
  echo "Installing dependencies failed." ;
  exit 1; 
}
#The following qml modules have to be installed separately (regex issue).
sudo apt-get install qml-module-qtquick.controls || {
  echo "Installing special dependency qml-module-qtquick.controls failed." ;
  exit 1; 
}
sudo apt-get install qml-module-qtquick.dialogs || {
  echo "Installing special dependency qml-module-qtquick.dialogs failed." ;
  exit 1; 
}
echo "Cloning..."
go get -d -u github.com/nanu-c/axolotl/ || {
  echo "Cloning failed" ;
  exit 1; 
}
cd $(go env GOPATH)/src/github.com/nanu-c/axolotl && go mod download || {
  echo "Downloading (go) failed" ;
  exit 1; 
}
echo "Installing..."
cd axolotl-web && npm install || {
  echo "Installing (npm) failed" ;
  exit 1; 
}
#node-sass does not support arm64 so it has to be rebuilt
echo "Rebuilding of npm-sass..."
npm rebuild node-sass || {
  echo "Rebuilding failed" ;
  exit 1; 
}
echo "Building (npm)..."
npm run build || {
  echo "Building (npm) failed" ;
  exit 1; 
}
cd .. && mkdir -p build/linux-arm64/axolotl-web
echo "Building (go)..."
env GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/axolotl . || {
  echo "Building (go) failed" ;
  exit 1; 
}
cp -r axolotl-web/dist build/linux-arm64/axolotl-web && cp -r guis build/linux-arm64
if [ -f /usr/share/applications/axolotl.desktop ];
then
  :
else
  echo "[Desktop Entry]\nType=Application\nName=Axolotl\nGenericName=Signal Chat Client\nPath=$HOME/go/src/github.com/nanu-c/axolotl/build/linux-arm64/\nExec=$HOME/go/src/github.com/nanu-c/axolotl/build/linux-arm64/axolotl\n#Exec=$HOME/go/src/github.com/nanu-c/axolotl/build/linux-arm64/axolotl -e qt\nIcon=$HOME/go/src/github.com/nanu-c/axolotl/build/linux-arm64/axolotl-web/dist/axolotl.png\nTerminal=false\nCategories=Network;Chat;InstantMessaging;Qt;\nStartupWMClass=axolotl" | sudo tee -a /usr/share/applications/axolotl.desktop
fi
echo "Congratulations! You should now see an Axolotl smiling in your app menu."

