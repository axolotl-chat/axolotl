# TextSecure client for the Ubuntu Phone

This is a preliminary source code release of a TextSecure compatible client for the Ubuntu Phone, written in Go and QML.
It builds upon the [Go textsecure package] (https://github.com/janimo/textsecure) and modified/hacked versions of the
Telegram for Ubuntu Phone QML interface.

Currently it is very basic and recommended for curious users and developers only.

What works
-----------

Phone registration, contact discovery, direct and group text messages, photo and video attachments.

What is missing
---------------

Persistent storing of conversations and many other features of the Android app. There are bugs and UI/UX quirks.

Installation
------------

Download the latest release from the app store or build it yourself

    ./build.sh rel

Install on a phone connected via adb

    ./install.sh

For more details check the [wiki] (https://github.com/janimo/textsecure-qml/wiki/Installation)

Contributing
-----------

User and developer discussions happen on the [mailing list] (https://groups.google.com/forum/#!forum/textsecure-go), everything else
is on github.
