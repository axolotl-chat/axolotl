# Migrating from `janimo/axolotl`

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
   If you do not remove the old app and you send or receive new messages with the other app, conflicts may occur.