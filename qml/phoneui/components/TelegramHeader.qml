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
    // Avatar {
    //     id: headerImage
    //
    //     chatId: telegramheader.chatId
    //     chatTitle: telegramheader.title
    //     chatPhoto: telegramheader.chatPhoto
    //     // isLogo: false
    //
    //     width: height
    //     height: parent.height * 4 / 5.0
    //     anchors {
    //         left: parent.left
    //         verticalCenter: parent.verticalCenter
    //     }
    //
    //     RotationAnimation {
    //         id: connectingAnimation
    //         target: headerImage
    //         direction: RotationAnimation.Clockwise
    //         from: 0
    //         to: 359
    //         loops: Animation.Infinite
    //         duration: 5000
    //         alwaysRunToEnd: false
    //         // running: isConnecting && headerImage.isLogo
    //         running: isConnecting
    //         properties: "rotation"
    //
    //         onRunningChanged: {
    //             if (!running) {
    //                 connectingAnimation.stop();
    //                 headerImage.rotation = 0;
    //             }
    //         }
    //     }
    // }

    // Image {
    //     id: secretChatImage
    //     source: Qt.resolvedUrl("../images/ic_lock_green.png");
    //     sourceSize.width: units.gu(1)
    //     anchors {
    //         left: parent.left
    //         leftMargin: visible ? units.gu(1) : 0
    //         verticalCenter: parent.verticalCenter
    //     }
    //     fillMode: Image.PreserveAspectFit
    //     visible: isSecretChat
    //     height: isSecretChat ? units.gu(1.8) : 0
    // }

    // TelegramLabel {
    //     id: titleText
    //     // We need fixed width. Otherwise, we overflow action icons.
    //     width: Math.min(implicitWidth, parent.width - headerImage.width - secretChatImage.width - anchors.leftMargin)
    //     anchors {
    //         top: parent.top
    //         topMargin: units.gu(1)
    //         left: secretChatImage.right
    //         leftMargin: units.gu(1)
    //     }
    //     verticalAlignment: Text.AlignVCenter
    //
    //     font.pixelSize: FontUtils.sizeToPixels("large")
    //     elide: Text.ElideRight
    //     text: title.length === 0 ? i18n.tr("Signal") : title
    //
    //     state: header.subtitle.length > 0 ? "subtitle" : "default"
    //     states: [
    //         State {
    //             name: "default"
    //             AnchorChanges {
    //                 target: titleText
    //                 anchors.verticalCenter: titleText.parent.verticalCenter
    //             }
    //             PropertyChanges {
    //                 target: titleText
    //                 height: titleText.implicitHeight
    //             }
    //         },
    //         State {
    //             name: "subtitle"
    //             PropertyChanges {
    //                 target: titleText
    //                 height: titleText.parent.height / 2
    //             }
    //         }
    //     ]
    //
    //     transitions: [
    //         Transition {
    //             AnchorAnimation {
    //                 duration: UbuntuAnimation.FastDuration
    //             }
    //         }
    //     ]
    // }
    //
    // Label {
    //     id: subtitleText
    //     width: Math.min(implicitWidth, parent.width - headerImage.width - anchors.leftMargin)
    //     anchors {
    //         left: headerImage.right
    //         leftMargin: units.gu(1)
    //         bottom: parent.bottom
    //         bottomMargin: units.gu(0.5)
    //     }
    //     verticalAlignment: Text.AlignVCenter
    //     height: parent.height / 2
    //
    // //    color: TelegramColors.blue
    //     fontSize: "small"
    //     elide: Text.ElideRight
    //     text: subtitle
    //
    //     Connections {
    //         target: header
    //         onSubtitleChanged: {
    //             subtitleText.opacity = 0;
    //             subtitleTextTimer.start();
    //         }
    //     }
    //
    //     Timer {
    //         id: subtitleTextTimer
    //         interval: UbuntuAnimation.FastDuration
    //         onTriggered: {
    //             subtitleText.text = header.subtitle;
    //             subtitleText.opacity = 1;
    //         }
    //     }
    //
    //     Behavior on opacity {
    //         NumberAnimation {
    //             duration: UbuntuAnimation.FastDuration
    //         }
    //     }
    // }

    MouseArea {
        anchors.fill: parent
        onClicked: {
        //     mouse.accepted = true;
        //     telegramheader.clicked();
        }
    }
}
