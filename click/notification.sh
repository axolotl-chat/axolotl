#!/bin/sh
echo "Notification"
gdbus call --session --dest com.ubuntu.Postal --object-path /com/ubuntu/Postal/textsecure_2Enanuc2 \
 --method com.ubuntu.Postal.Post la\
 '"{\"message\": \"foobar\", \"notification\":{\"card\": {\"summary\": \"Sofla\", \"body\": \"hello\", \"popup\": true, \"persist\": true, \"tag\":\"chat\",\"sound\":\"buzz.mp3\", \"vibrate\":{\"pattern\":[200,100],\"duration\":200,\"repeat\":2}}}}"'
