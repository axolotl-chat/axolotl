#!/bin/bash

# make sure the script fails on errors
set -Eeuo pipefail
echo "Running postbuild script"
case $ARCH in
	amd64)
		ARCH_NAME=x86_64
		;;
	arm64)
		ARCH_NAME=aarch64
		;;
	armhf)
		ARCH_NAME=armv7
		;;
esac

readonly FILENAME=libzkgroup_linux_$ARCH_NAME.so

if [ -v $GOPATH ]
then
	# Github
	readonly ZKGROUP_GITHUB=/github/workspace/go/src/github.com/nanu-c/zkgroup/lib/$FILENAME
	cp $ZKGROUP_GITHUB $@/lib/;
	rm -f ${INSTALL_DIR}/\\${GITHUB_WORKSPACE}
else
	# Clickable
	readonly ZKGROUP_OPTION1=$GOPATH/pkg/mod/github.com/nanu-c/zkgroup@v0.9.0/lib/$FILENAME
	readonly ZKGROUP_OPTION2=$GOPATH/src/github.com/nanu-c/zkgroup/lib/$FILENAME

	mkdir -p $CLICK_LD_LIBRARY_PATH

	[ -f $ZKGROUP_OPTION1 ] && cp $ZKGROUP_OPTION1 $CLICK_LD_LIBRARY_PATH/;
	# [ -f $ZKGROUP_OPTION2 ] && cp $ZKGROUP_OPTION2 $CLICK_LD_LIBRARY_PATH/;

	if [ ! -f $CLICK_LD_LIBRARY_PATH/$FILENAME ]
	then
		echo "Didn't find $FILENAME which is required"
		exit 1
	fi
fi
