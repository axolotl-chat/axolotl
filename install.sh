#!/bin/bash

CLICK_NAME=textsecure.jani*click

adb push $CLICK_NAME /home/phablet
adb shell pkcon install-local $CLICK_NAME --allow-untrusted
