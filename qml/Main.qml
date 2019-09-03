import QtQuick 2.4
import Ubuntu.Components 1.3
//import QtQuick.Controls 2.2
import Morph.Web 0.1

MainView {
    id: root
    objectName: 'root'
    applicationName: 'textsecure.nanuc'

    width: units.gu(45)
    height: units.gu(75)

    Item {
        anchors.fill: parent
            WebView  {
                id: webViewBlub
                width: parent.width
                height: parent.height
                anchors.fill: parent
                url: "http://localhost:8080/"
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
