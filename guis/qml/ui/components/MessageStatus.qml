import QtQuick 2.4
import Ubuntu.Components 1.3
import "TelegramColors.js" as Color

Item {
    property string time

    Rectangle {
        visible: isPhotoOrVideo
        anchors {
            leftMargin: units.gu(-1)
            rightMargin: units.gu(-1)
            topMargin: units.dp(-2)
            bottomMargin: units.dp(-2)
            fill: messageStatusRow
        }
        color: Qt.rgba(0, 0, 0, 0.4)
        radius: units.dp(3)
    }

    height: messageStatusRow.height
    width: messageStatusRow.width

    Row {
        id: messageStatusRow
        spacing: units.dp(4)

        Label {
            id: timeLabel
            anchors.verticalCenter: parent.verticalCenter
            font.weight: Font.DemiBold
            fontSize: "x-small"
            color: {
                if (isPhotoOrVideo) {
                    return Color.white;
                }
                return outgoing ? "black" : "white"
            }
            text: time
        }

        Image {
            id: messageSentStatus
            anchors.verticalCenter: parent.verticalCenter
            width: units.gu(2)
            height: width
            visible: outgoing
            z: 1
            fillMode: Image.PreserveAspectFit
            source: {
                var icon=""
                var c = ""

                if (!isSent) {
                    icon = "msg_clock"
                } else if (!isRead) {
                    icon="Checks1_2x"
                } else {
                    icon="Checks2_2x"
                }
                if (isPhotoOrVideo) {
                    c = "_white"
                }
                return Qt.resolvedUrl("../images/"+icon+c+".png");
            }
        }
    }
}
