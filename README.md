# TextSecure client for the Ubuntu Phone

This is a Signal compatible client for the Ubuntu Phone, written in Go and QML.
It builds upon the [Go textsecure package] (https://github.com/nanu-c/textsecure/) and modified versions of the
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

For more details check the [INSTALL.md] (https://github.com/nanu-c/textsecure-qml/INSTALL.md)

Contributing
-----------

User and developer discussions happen on the [mailing list] (https://groups.google.com/forum/#!forum/textsecure-go), everything else
is on github.

Migrating from `janimo/textsecure-qml`
--------------------------------------
1. Download and install the app from the OpenStore; do not launch the app!
2. Copy the directory `/home/phablet/.local/share/textsecure.jani/.storage` to
   `/home/phablet/.local/share/textsecure.nanuc/.storage`
3. Copy the file `/home/phablet/.config/textsecure.jani/config.yml` to
   `/home/phablet/.config/textsecure.nanuc/config.yml`.
   Edit the copied file by changing `storageDir: /home/phablet/.local/share/textsecure.nanuc/.storage` (not strictly required: also
   update `userAgent: TextSecure 0.3.18 for Ubuntu Phone` to reflect the current version).
4. _Not strictly required._
   Copy your conversation history by copying the file `/home/phablet/.local/share/textsecure.jani/db/db.sql` to
   `/home/phablet/.local/share/textsecure.nanuc/db/db.sql`
5. _Not strictly required._
   Copy the attachments by copying the directory `/home/phablet/.local/share/textsecure.jani/attachments` to
   `/home/phablet/.local/share/textsecure.nanuc/attachments`.
   Download the `db.sql` to your computer and run `sqlite3 db.sql "UPDATE messages SET attachment = REPLACE(attachment,
   '/home/phablet/.local/share/textsecure.jani/attachments/', '/home/phablet/.local/share/textsecure.nanuc/attachments/') WHERE
   attachment LIKE '/home/phablet/.local/share/textsecure.jani/attachments/%';"`.
   Upload the now updated `db.sql` back to your phone.
6. **Remove the old app!**
   If you do not remove the old app and you send or receive new messages with the other app you, conflicts may occur.
