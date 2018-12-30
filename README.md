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

The build-system is now integrated in the `clickable` Version 3.2.0.
* Install [Golang] (https://golang.org/doc/install
* Add gopath to ~/.bashrc https://github.com/golang/go/wiki/SettingGOPATH
* Check out this git `go get -d https://github.com/nanu-c/textsecure-qml`
* `cd $GOPATH/src/github.com/nanu-c/textsecure-qml`
* install dependencies `sudo apt install mercurial bzr`
* Get dependencies `go get -d ./...`
* Get [clickable](https://github.com/bhdouglass/clickable#install)
* Build the modified docker container with `cd docker&& docker build -t nanuc/ut-textsecure-sdk:16.04 .`
* Back to main dir, then
* Run clickable `clickable`, this also transfers the click package to the Ubuntu Touch Phone
* Run `clickable launch logs` to start signal and watch the log

Contributing
-----------

Please fill issues here on github https://github.com/nanu-c/textsecure-qml/issues .

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
