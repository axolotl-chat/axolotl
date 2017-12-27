#!/bin/bash

CLICK_NAME=textsecure.nanuc*click

adb push $CLICK_NAME /home/phablet
adb shell pkcon install-local $CLICK_NAME --allow-untrusted
clickable launch logs
