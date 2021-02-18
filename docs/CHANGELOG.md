0.9.8 (Feb 10 2021)
------------------------------------
* Fix recieving messages because of Signal-API changes(nanu-c)
* Delviver call messages in the appropiate Chats (nanu-c)


0.9.7 (Feb 8 2021)
------------------------------------
* Add support for registration captchas (nanu-c)
* Deduplicate chat's that exists twice (nanu-c)
* German translation update (danfro)
* Add complete translation for PT-BR (yds12)
* Don't show the currently broken warning everywhere (flaburgan)
* Fix messages sent by signal desktop appearing in the wrong chat (nanu-c)

0.9.6 (Jan 31 2021)
------------------------------------
* Add a message to warn about current broken state in Axolotl due to upstream Signal changes (flaburgan)
* Fix messages sometimes going to the wrong conversation (nanu-c)
* Fix duplicate conversations for the same group (nanu-c)
* Make the debug screen harder to reach and allow users to exit if they went there by mistake (flaburgan)

0.9.5 (Jan 15 2021)
------------------------------------
* Support for empty prekeys - solves message sending errors (nanu-c)
* Support for uuids (nanu-c)
* Add a more telling error message when something isn't working (flaburgan)
* Fix sending messages if a signal desktop is linked (nanu-c)

0.9.4 (Jan 15 2021)
------------------------------------
* Warn users that due to upstream changes in Signal, importing contacts is currently broken. (flaburgan)
* Allow to manually add contacts as a workaround by disabling discovery (nanu-c)
* Update Dutch translations (Vistaus)

0.9.3 (Jan 13 2021)
------------------------------------
* Add an indication in the "Settings" page about how to enable notifications in Axolotl (flaburgan)
* Fix the desktop shortcut in Flatpak installation (joshbowyer)
* When quoting a message and answer with an image, display the quoted message above the image (flaburgan)
* Add a link to the "Contacts" page in the main menu (flaburgan)
* Add the sender name of a quoting message in groups (flaburgan)
* Add the complete date and time of a message when tapping the "From now" time (flaburgan)
* Add a "Loading" spinner on the picture when sending (flaburgan)

0.9.2 (Dec 21 2020)
------------------------------------
* Hotfix - Fix XSS injection in messages

0.9.0.1 (Dec 9 2020)
------------------------------------
* Hotfix - disappearing messages

0.9.0 (Nov 30 2020)
------------------------------------
* Improve ci-scripts, appimage and flat-package (olof-nord)
* Add support for displaying quoted messages (flaburgan & nanu-c)
* Add a loading screen until Axolotl is ready (flaburgan)
* Fix avatar handling (blackoverflow)
* Add support for message read status ticks (nanu-c)
* Improve documentation (olof-nord)
* Fix message syncing between signal desktop and axolotl (nanu-c)
* Indicate if the attachment is a video and handle video donwloads correctly (nanu-c)
* Update translations (Anne017 + danfro)

0.8.9 (Oct 09 2020)
------------------------------------
* Fix scrolling to old messages display duplicates, by flaburgan
* Fix regression, reception of vocal messages made Axolotl crashes, by nanu-c
* Fix linked desktop clients not displayed, by flaburgan
* Improve snap packaging, by olof-nord
* Translations update by DanFro and Anne017

0.8.8 (Oct 04 2020)
------------------------------------
* Display every pictures of received messages, not several time the first one, by nanu-c
* Send message immediatly even when it has more than one line (no need to tap a second time), by flaburgan
* Fix avatar icons looking strange after navigation to linked devices list, by flaburgan
* Fix snap package builds, by nanu-c
* And many small fixes and cleanups, by nanu-c, flaburgan & olof-nord

0.8.7 (Sept 27 2020)
------------------------------------
* support reciving quotes, reactions(somehow), contacts (fixes the flood with the timelimit notice)
* fix attachments sometimes not downloaded (gifs)
* respect newlines in messages

0.8.6 (Aug 30 2020)
------------------------------------
* fix contact import to signal desktop
* fix group import to signal desktop
* fix duplicate messages in signal desktop

0.8.5 (Aug 29 2020)
------------------------------------
* fix attachments due to changed signal api

0.8.4 (Aug 14 2020)
------------------------------------
* Update french translations by Anne017
* Minor UX/UI glitches
* Don't expose the whole filesystem, see #140
* Long press on the message input field doesn't delete the written text, but appends
* Update all dependencies
* Save images and videos via content hub



0.8.3 (Jul 26 2020)
------------------------------------
* Support for QR-Code fingerprints by blackoverflow
* Dutch translation by Heimen Stoffels
* Only push to dbus if available by Björn Tantau
* Server only mode by Björn Tantau

0.8.2 (May 5 2020)
------------------------------------
* Show number of contact person in the chat menu and make it callable by arthur
* Refactor Layout with a new design by Flaburgan
* Update French translation by Anne017
* Improve backend message worker, fixes some crashes

0.8.1 (Apr 16 2020) hotfix and translations
------------------------------------
* Portugese translations by rubencarneiro
* fix message sending states
* fix contact imports
* fix building snap package

0.8.0 (Apr 13 2020)
------------------------------------
* Support for DarkMode by Flaburgan (and a little bit nanu-c)
* UI/UX improvements by Flaburgan
* Translations (help needed, see github for how to translate)
* French translation by Flaburgan
* Norwegian translation by Ari
* German translation
* Basic support for self destroying messages
* more Error handling on registration failures

Bugfixes:
* Members of groups are now correctly shown
* fix set password page

0.7.7.2 (Mar 05 2020) Hotfix
------------------------------------
* Don't display always "file" on pure text messages
* Fix sending messages
* fix sending attachments
* fix sendind attachments from snap

0.7.7.1 (Mar 03 2020) Hotfix
------------------------------------
* Don't display always "file" on pure text messages

0.7.7 (Mar 03 2020) Stability
------------------------------------
* Support for multi attachments
* Message Input box resizes as it should
* catch websocket panics leads to more stability

0.7.6 (Feb 27 2020) Stability
------------------------------------
* Autofocus in lots of places
* mark unsend Messages as error (still at the beginning of the chat after reentering)
* Notifications are delted after entering the coresponding chat
* Images/videos have now a fullscreen mode

0.7.5 (Dez 21 2019) Axolotl-Beta: second beta version
--------------------
* [ut] Fix attachment sending
* [snap] add snap support

0.7.4 (Dez 16 2019) Axolotl-Beta: second beta version
---------------------
* fix editing groups
* Handle urls in messages
* fix deleting contacts
* update libraries + fix linter errors
* first step into snap's
* Verification code input is visible again
* longpress pastes the clipboard

0.7.2.1 (Dez 3 2019)
--------------------
* qUICK FIX FOr handling recieved messages

0.7.2 (Dez 2 2019) Axolotl-Beta: second beta version
--------------------
* Show group avatars
* Update group if group is corrupted
* Improve deleting/editing Chats/Contacts on longpress
* Update to next signal API revision
* Complete renaming to axolotl
* Show phone number/group memebers in the header
* update axolotl-web packages
* some ux improvements


0.7.0 (Okt 19 2019) Axolotl-Beta: first beta version
--------------------
* refactor message input field alittle bit more
* cleanup log

0.6.15 (Okt 19 2019) Axolotl-Alpha
--------------------
* refactor message input field
* fix Unknown groups
* minor fixes in dates and menu

0.6.14.1 (Okt 18 2019) Axolotl-Alpha
* fix typo
0.6.14 (Okt 18 2019) Axolotl-Alpha
--------------------
* jump to top on showing chat list
* fix app not focus on canceling desktop sync
* fix incomming group message added to wrong chat
* scroll down on enter chat is now faster
* support for latest clickable
* handle external urls externally

0.6.13 (Okt 15 2019) Axolotl-Alpha
--------------------
* Add Info page before registering
* About and Settings page
* creating a chat shows the correct title
* mention tagger on ut for scanning the desktop qr-code
* indicate that importing contacts takes times
* add reset session and show identity
* adding yourself to a group is not possible anymore

0.6.12 (Okt 13 2019) Axolotl-Alpha
--------------------
* support notifications on non ut os
* unread messages counter
* enable/disable Notifications
* show sender of message in group chats
* support of contact imports with multiple numbers


0.6.11 (Okt 11 2019) Axolotl-Alpha
--------------------
* indicate wrong password + unregister when password forgotten
* fix empty contacts list, edit/delete contacts, hopefully also show always names in the chatList
* show ratelimiting error in contact list
* improve logging

0.6.10 (Okt 06 2019) Axolotl-Alpha -rc1
--------------------
* support for encrypted db's


0.6.9 (Okt 06 2019) Axolotl-Alpha
--------------------
* send attachments ut
* fix import contacts on ut
* Create groups: show contact list also the first time the modal is opened
* Clean logs and 2 typos thanks @TotalSonic

0.6.8 (Okt 2 2019) Axolotl-Alpha
--------------------
* creat Group Chats
* sort chat list
* cleanup log

0.6.7 (Sept 30 2019) Axolotl-Alpha
--------------------
* fix entering new chat and show first sent message

0.6.6 (Sept 30 2019) Axolotl-Alpha
--------------------
* rewrite vcf parsing

0.6.5 (Sept 24 2019) Axolotl-Alpha
--------------------
* versions for raspberryPi and windows
* fix contacts import in ut


0.6.4 (Sept 23 2019) Axolotl-Alpha
--------------------
* delete/edit contacts
* dynamicaly growing messagefield
* chat/contacts action headers

0.6.3 (Sept 21 2019) Axolotl-Alpha
--------------------
* remove chats,
* crossplattform webview,
* contacts import on non ut devices,
* snaps are building (but not running)
* remove zbar dependencies

0.6.2 (Sept 17 2019) Axolotl-Alpha
--------------------
* More stable ux, dialogs/menu closing, show phone numbers in contact list

0.6.1 (Sept 14 2019) Axolotl-Alpha
--------------------
* Import contacts from content hub

0.6.0 (Sept 14 2019) Axolotl-Alpha
--------------------
* remove go-qml bindings
* new ux
* display attachments inline
* know bugs because of the complete rewrite of the client:
** no group creation, no notification config, no content hub integration(no contact import, no saving)
** no support of encrypted db

0.4.6 (Sep 04 2019) Hotfix
--------------------
* fix Contact import working properly
* fix recieving messages
* still not fixed: open new chat

0.4.5 (Aug 31 2019)
--------------------
* import a singel contact thats not in the contact Book
* UX improvements
0.4.4 (Aug 30 2019)
--------------------
* fix contacts import, old contacts are not replaced anymore
* fix leading actions on dialog page
0.4.3 (Aug 29 2019)
--------------------
* fix unregistering
0.4.2 (Aug 30 2019)
--------------------
* Update protobufs and wrong imports
* Registration Failures like RateLimiting are shown
* support for clickable desktop

0.4.1
--------------------
* Update translations in German, Spanish, French and English typos -&gt; thanks to advocatux, Anne017, Danfro and RobertZenz
* Debuglog switch works now
0.4.0
--------------------
* Fix Build
* Fix language creations
* UX improvements
0.3.261 (May 12 2018)
--------------------
* Manifest.json changes to pass click-review
* Add clickable.json for 16.04

0.3.26 (May 10 2018)
--------------------
* Fix Apparmor policy (thanks @DanChapman and @DanChapman)
* Clean up
* Disable Notifications in chatoptions

0.3.25 (May 10 2018)
--------------------
* local Push Notifications as long as the app is running
* Enable Dbug Log in Options

0.3.24 (Apr 22 2018)
--------------------
* starting a new Chat is working again

0.3.23 (Apr 21 2018)
--------------------
* Reconnect correctly after disconection

0.3.22 (Apr 17 2018)
--------------------
* Update to Ubuntu Components 1.3
* Build with clickable
* Make the search functional
* Avatars in GroupChats
* Disabling the Password works now
* Sending a contact includes now the name
* Set default loglevel to Info
* Fix crash on startup, when db is still encrypted but incoming messages are waiting



0.3.21 (Mar 03 2018)
--------------------
* late night typo, preventing showing a dialog

0.3.20 (Mar 03 2018)
--------------------
* fix PasswordInput page preventing Start of app, fix Advancedsettings Page, fix appearing keyboard on Startup, more elegant solution for fix in 0.3.19

0.3.19 (Feb 28 2018)
--------------------
* fix register process

0.3.18 (Feb 14 2018)
--------------------
* Link devices

0.3.17 (Jan  11 2018)
--------------------
* Secure Message store

0.3.16 (Jan  03 2018)
--------------------
* Get sending attachments working again

0.3.15 (Jan  03 2018)
--------------------
* fix Contact Book

0.3.14 (Jan  01 2018)
--------------------
* rename to signal
* hide Send Attachment until it's working again
* faster load of Dialogs
* structure code in go Packages


 0.3.13 (Dez  27 2017)
 --------------------
 * replace qml-go with newer version
 * move it to the open store
 * replace maintainerpaths


0.3.12 (Apr 29 2017)
--------------------

* Fix crash on receiving empty messages.

0.3.11 (Feb 29 2016)
--------------------

* Fix importing contact numbers with weird characters.
* Improve attachment UI.
* Log app messages to ~/.cache/textsecure.jani/log.
* Allow uploading application debug logs to github.

0.3.10 (Jan 20 2016)
--------------------

* Add unregistration UI to allow fixing broken setups.
* Make sure to use contact names not numbers in session titles.
* UI to double check phone number when registering.
* Try to fix incorrectly entered phone numbers when registering (i.e. dropping extra 0s)

0.3.9 (Jan 08 2016)
-------------------

* Reduce excessive CPU usage even when idle.
* Show video thumbnails.
* Do not cover text with the thumbnail when a message has both text and media content.
* Remove intro page, add a 'Help' menu item instead.
* Small UI cleanups.

0.3.8 (Dec 02 2015)
-------------------

* Always show most recent conversations at the top.
* Allow deleting single messages.
* Allow creating a group with only one other member besides us.
* On startup try sending out previously unsent messages.
* Remove old-style storage dir if detected instead of asking the user to remove it manually.
* Rework signup page wording and layout and localize it.

0.3.7 (Nov 29 2015)
-------------------

 * Add translations for 35 languages, imported from Signal for Android.
 * Allow deleting a conversation.
 * Emphasize conversations with unread messages.
 * Add 'Mark all read' menu item.
 * Show the phone number of the contact in the conversation page.
 * Use a distinct style for group update messages.

0.3.6 (Nov 20 2015)
-------------------

 * Allow placing a phone call (regular, non-secure) to the contact we're messaging.
 * Fix message info action.
 * Rearrange/remove/add some menu actions to match Signal on Android.
 * Simplify welcome page.

0.3.5 (Nov 04 2015)
-------------------

* Also import contacts that have no international prefix in the address book.
* Make sure outgoing attachments are stored outside the ephemeral content-hub cache.
* Handle peers that have more than one device registered.
* Handle peers that changed identity keys via reregistering.
* Fix group updates so they do not deactivate the session.

0.3.4 (Oct 27 2015)
-------------------

* Add UI for verifying peer identity.
* Save settings to file.
* Do not allow sending to groups we left.
* Better detection of incoming attachment mimetypes.

0.3.3 (Oct 21 2015)
-------------------

* Show contact avatars if present in the system address book.
* Show group avatars when available.
* Prevent video preview crash when viewing more than one video.
* Allow viewing incoming video attachments.
* Do not block the UI at all when sending attachments.
* Clearer group update messages.

0.3.2 (Oct 17 2015)
-------------------

* Backend robustness fixes, drop duplicates and other invalid messages.
* Fix lockup in sign-up page if button is pressed more than once.
* Small UI fixes.

0.3.1 (Oct 12 2015)
------------------

* Add group update and group leave functionality.
* Use multiple dialog bubble colors.
* Stop offering sending random files as attachments, Android does not support it.
* Allow sending contact phone numbers as attachments.
* Bugfixes.

0.3.0 (Oct 5 2015)
-------------------

* Persist conversations by saving sessions to SQLite and attachments to files.
  Storage is bound to change format and is currently unencrypted.
  Things may get lost on upgrades until it stabilizes.
* Improved group messaging.
* Allow resetting the session (to debug or to get rid of corrupted sessions).
* Support sending and playback of audio attachments.
* Add session search.
* Show send timestamps instead of receive timestamps on messages and sessions.

0.2.6 (Sep 25 2015)
------------------

 * Errors no longer require an app restart but are presented in a dialog.
 * Fix outgoing timestamp that was in the very distant future.
 * Fix the contact search field.
 * Fix group member selection.
 * Fix avatars in group chats.
 * Do not allow sending attachments larger than 100 Mib to prevent OOM.
 * Add group info dialog.
 * Add message info dialog when one message is selected to show timestamps.
 * Add welcome page to run on app start.

0.2.5 (Sep 22 2015)
------------------

 * Make sending messages asynchronous, so they don't block the UI.
 * Show sending/sent/delivered status on outgoing messages.
 * Handle network disconnections.

0.2.4 (Sep 18 2015)
------------------

 * Allow sending attachments in group messages.
 * Show thumbnails for image attachments.

0.2.3 (Sep 11 2015)
------------------

 * Stop encrypting the session metadata and asking for an encryption password.
   This follows the Android app's decision, and is needed for cross-device sync
   in the future.

0.2.2 (Sep 8 2015)
------------------

 * Initial incomplete group messaging support.
 * Show app version in settings.
 * Fix crash introduced in 0.2.1

0.2.1 (Sep 8 2015)
------------------

 * Handle contacts with multiple phone numbers.

0.2.0 (Sep 4 2015)
------------------

 * Get address book contacts via the content hub instead of the DBus service
   so it can be published in the app store.
 * Show per user color avatars.
 * Fix app icon size to look better in the app scope.

0.1 (Aug 13 2015)
---

 * Initial public release.
