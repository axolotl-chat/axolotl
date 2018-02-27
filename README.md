# TextSecure client for the Ubuntu Phone

This is a Signal compatible client for the Ubuntu Phone, written in Go and QML.
It builds upon the [Go textsecure package](https://github.com/nanu-c/textsecure/) and modified versions of the
Telegram for Ubuntu Phone QML interface.

What works
-----------

 * Phone registration
 * Contact discovery
 * Direct and group messages
 * Photo, video, audio and contact attachments in both direct and group mode
 * Preview for photo and audio attachments
 * Storing conversations
 * Encrypted message store
 * Desktop client provisioning/syncing

What is missing
---------------

 * Push notifications
 * Most settings that are available in the Android app
 * Encrypted phone calls

There are still bugs and UI/UX quirks.

Installation
------------

Download the latest release from the app store or build it yourself (you'll need docker running)

    ./build.sh rel

Install on a phone connected via adb

    ./install.sh

For more details check the [INSTALL.md](https://github.com/nanu-c/textsecure-qml/blob/master/docs/INSTALL.md).

Contributing
-----------

User and developer discussions happen on the [mailing list] (https://groups.google.com/forum/#!forum/textsecure-go), everything else
is on github.
