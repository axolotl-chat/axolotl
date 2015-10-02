import QtQuick 2.0
import Ubuntu.Components 0.1
import "../components"

TelegramPage {
    id: page

    head.backAction.visible: false

    pageTitle: i18n.tr("Welcome!")

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
            text: "<h3>Thanks for trying out TextSecure!</h3><br><br> \
            This app is not production ready yet, in particular conversation storage is not encrypted \
            and its format is not finalized, support for groups is not complete, some corner case errors are not handled,\
            and there are other bugs and UI issues.<br><br>\
            However it is already quite usable and can get even better with your help.\
            Read the changelog, file bugs and contribute patches on github :)<br>\
            <a href='https://github.com/janimo/textsecure-qml'>https://github.com/janimo/textsecure-qml</a><br>"
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
