#!/bin/sh
echo "Notification"

# for sending
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/textsecure_2Enanuc --method com.ubuntu.Postal.Post textsecure.nanuc_textsecure '"{\"message\": \"foobar\", \"notification\":{\"card\": {\"summary\": \"Sofla\", \"body\": \"hello\", \"popup\": true, \"persist\": true, \"tag\":\"chat\",\"sound\":\"buzz.mp3\", \"vibrate\":{\"pattern\":[200,100],\"duration\":200,\"repeat\":2}}}}"'
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/textsecure_2Enanuc --method com.ubuntu.Postal.ClearPersistent textsecure.nanuc_textsecure chat
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/textsecure_2Enanuc --method com.ubuntu.Postal.PopAll textsecure.nanuc_textsecure
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/textsecure_2Enanuc --method com.ubuntu.Postal.ListPersistent textsecure.nanuc_textsecure


gdbus introspect --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/textsecure_2Enanuc --xml
