import QtQuick 2.4
import Ubuntu.Components 1.3
import "../components"

TelegramPage {
    id: page

    head.backAction.visible: false

    body: Item {
        anchors {
            fill: parent
            margins: units.gu(2)
        }

        Text {
            id: infoText
            elide: Text.ElideRight
            anchors {
                top: parent.top
                margins: units.gu(1)
            }
            width: parent.width
            wrapMode: Text.WordWrap
            text: "<h3>Thanks for trying out Signal!</h3><br><br> \
            File bugs and feature requests on github:<br>\
            <a href='https://github.com/nanu-c/textsecure-qml/issues'>https://github.com/nanu-c/textsecure-qml/issues</a><br>"
            onLinkActivated:Qt.openUrlExternally(link)
        }

        TelegramButton {
            anchors {
                top: infoText.bottom
                topMargin: units.gu(1)
                right: parent.right
                left: parent.left
            }
            width: parent.width

            text: i18n.tr("OK")
            onClicked: pageStack.push(dialogsPage)
        }
    }
}
