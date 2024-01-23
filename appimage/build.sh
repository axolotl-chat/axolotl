#!/usr/bin/env bash

# https://sipb.mit.edu/doc/safe-shell/
set -eu -o pipefail

for build_dependency in go npm appimagetool
do
  if [ ! -f "$(command -v "${build_dependency}")" ]; then
    echo "${dependency} is required!"
    echo "Please install all build dependencies first."
    exit 1
  fi
done

echo "build axolotl-web"
pushd ../axolotl-web
  npm ci
  npm run build
popd

echo "packaging it all up"
pushd ../
  rm -rf build/AppDir

  mkdir -p build/AppDir/usr/bin
  cp -f axolotl build/AppDir/usr/bin/axolotl

  mkdir -p build/AppDir/usr/bin/axolotl-web
  cp -rf axolotl-web/dist build/AppDir/usr/bin/axolotl-web

  cp -f appimage/AppDir/AppRun build/AppDir/AppRun
  cp -f appimage/AppDir/axolotl.desktop build/AppDir/axolotl.desktop
  cp -f appimage/AppDir/axolotl.png build/AppDir/axolotl.png

  mkdir -p build/AppDir/usr/share/metainfo
  cp -f appimage/AppDir/axolotl.appdata.xml build/AppDir/usr/share/metainfo/axolotl.appdata.xml

  appimagetool build/AppDir
popd
