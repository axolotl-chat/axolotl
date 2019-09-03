import QtQuick 2.4
// import Ubuntu.Components 1.3

import QtWebView 1.1

Rectangle{
    id: root
    objectName: 'root'
    // applicationName: 'textsecure.nanuc'

    width: 320
    height: 480

    Item {
        anchors.fill: parent
            WebView  {
                id: webViewBlub
                width: parent.width
                height: parent.height
                anchors.fill: parent
                url: "http://localhost:9080/"
            }
    }
    // PushClient {
    //   id: pushClient
    //   appId: "textsecure.nanuc_textsecure"
    //   onTokenChanged: {
    //     //console.log("Push client token is", token)
    //   }
    // }
}
