#!/bin/bash

# copy click files
echo "copy click files $@"

cp -a ../../../click/* $@
# Build axolotl-web
echo "update translations and build axolotl-web $@"
cd ../../../axolotl-web&&npm run translate &&npm run build && mkdir $@/axolotl-web&&cp dist $@/axolotl-web/ -r

echo $ARCH
if [ $ARCH == "arm64" ]
then
    [ -z '$GOPATH' ] && cp /github/workspace/go/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_aarch64.so $@/lib/libzkgroup_linux_aarch64.so; 
    cp $GOPATH/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_aarch64.so $@/lib/libzkgroup_linux_aarch64.so ||true; 
    rm ${INSTALL_DIR}/\\$GITHUB_WORKSPACE||true
fi
if [ $ARCH == "armhf" ]
then
    [ -z '$GOPATH' ] && cp /github/workspace/go/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_armhf.so $@/lib/libzkgroup_linux_armhf.so; 
    cp $GOPATH/src/github.com/nanu-c/zkgroup/lib/libzkgroup_linux_armhf.so $@/lib/libzkgroup_linux_armhf.so ||true; 
    rm ${INSTALL_DIR}/\\$GITHUB_WORKSPACE||true
fi