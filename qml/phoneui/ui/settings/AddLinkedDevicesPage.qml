import QtQuick 2.4
import Ubuntu.Components 1.3
import "../../components"
import Ubuntu.Content 1.1

TelegramPage {
    id: addlinkdevice
    visible: true
    header: PageHeader {
      title: i18n.tr("Link a Device")
        id: pageHeader
        width: parent.width
        leadingActionBar.actions:[
          Action {
            id: backAction
            iconName: "back"
            onTriggered:{
              back();
            }
          }
        ]
    }

    Rectangle {
        width: parent.width -100
        height: parent.height -100
        color: "white"
        border.color: "black"
        border.width: 5
        radius: 10
        anchors.top: pageHeader.bottom
        anchors.verticalCenter: parent.verticalCenter
        anchors.horizontalCenter: parent.horizontalCenter
        Camera {

        }
    }


}
