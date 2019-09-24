import QtQuick 2.4
import Ubuntu.Components 1.3
import "TelegramColors.js" as TelegramColors

// This component is intented for use in Ubuntu Header content.
PageHeader {
    id: telegramheader
    property int chatId: 0
    property string chatPhoto: ""
    property string title: ""
    property string subtitle: ""

    property bool isSecretChat: false
    property bool isConnecting: false
    clip: true

    signal clicked()
    height: units.gu(7)

    MouseArea {
        anchors.fill: parent
        onClicked: {
        //     mouse.accepted = true;
        //     telegramheader.clicked();
        }
    }
}
