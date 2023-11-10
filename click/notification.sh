#!/bin/sh
echo "Notification"

# for sending
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/axolotl_2Enanuc --method com.ubuntu.Postal.Post axolotl.nanuc_axolotl '"{\"message\": \"foobar\", \"notification\":{\"card\": {\"summary\": \"Sofla\", \"body\": \"hello\", \"popup\": true, \"persist\": true}, \"tag\":\"chat\",\"sound\":\"buzz.mp3\", \"vibrate\":{\"pattern\":[200,100],\"duration\":200,\"repeat\":2}}}"'
# show my persitent messages with the respective tag
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/axolotl_2Enanuc --method com.ubuntu.Postal.ListPersistent axolotl.nanuc_axolotl
# clear all messages with tag chat
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/axolotl_2Enanuc --method com.ubuntu.Postal.ClearPersistent axolotl.nanuc_axolotl chat
# get all my messages
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/axolotl_2Enanuc --method com.ubuntu.Postal.PopAll axolotl.nanuc_axolotl
