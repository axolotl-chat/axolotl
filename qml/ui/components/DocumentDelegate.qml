import QtQuick 2.4
import Ubuntu.Components 1.3

Item {
    id: documentBox

    signal clicked(var mouse)

    width: units.gu(28)
    height: units.gu(8)

    Rectangle {
        id: documentBoxBackground
        anchors {
            left: parent.left
            top: parent.top
            bottom: parent.bottom
        }
        width: height
        color: Qt.rgba(0, 0, 0, 0.1)
    }

    Icon {
        id: attachmentIcon
        anchors {
            centerIn: documentBoxBackground
        }
        color: textColor
        height: parent.height - units.gu(3)
        width: height
        name: "attachment"
    }

    Image {
        anchors {
            fill: documentBoxBackground
        }
        source: Qt.resolvedUrl(thumbnail)
        asynchronous: true
        fillMode: Image.PreserveAspectCrop
    }

    Image {
        anchors {
            fill: documentBoxBackground
            margins: units.gu(1)
        }
        visible: needsDownload
        source: {
            if (needsDownload) {
                if (isDownloading) {
                    return Qt.resolvedUrl("../images/photocancel.png");
                } else {
                    return Qt.resolvedUrl("../images/photoload.png");
                }
            }
            return ""
        }
    }

    Label {
        id: documentFileNameLabel
        anchors {
            top: parent.top
            left: documentBoxBackground.right
            right: parent.right
            topMargin: units.dp(4)
            leftMargin: units.gu(1)
        }
        elide: Text.ElideRight
        fontSize: userSettings.contents.fontSize
        text: documentFileName
        color: textColor
    }

    Label {
        anchors {
            top: documentFileNameLabel.bottom
            left: documentFileNameLabel.left
            topMargin: units.gu(1)
        }
        fontSize: userSettings.contents.fontSize
        text: delegateUtils.humanFileSize(documentSize, true)
        color: textColor
    }

    MouseArea {
        anchors.fill: parent
        onClicked: documentBox.clicked(mouse)
    }
}
