import QtQuick 2.4
import "TelegramColors.js" as TelegramColors

Rectangle {
    property alias text: buttonLabel.text
    property alias textColor: buttonLabel.color

    signal clicked

    width: buttonLabel.width + units.gu(10)
    height: buttonLabel.height + units.gu(2)
    color: enabled ? TelegramColors.blue : TelegramColors.grey
    radius: 10

    Behavior on color {
        ColorAnimation {
            duration: 300
        }
    }

    Text {
        id: buttonLabel
        anchors.centerIn: parent
        horizontalAlignment: TextInput.AlignHCenter
        verticalAlignment: TextInput.AlignVCenter

        color: TelegramColors.white
        font.pixelSize: FontUtils.sizeToPixels("medium")
    }

    MouseArea {
        id: startMessaging
        anchors.fill: parent

        onPressed: parent.onPressed()
        onReleased: parent.onReleased()
        onClicked: parent.clicked()
    }

    function onPressed() {
        color = TelegramColors.dark_blue;
    }

    function onReleased() {
        color = TelegramColors.blue
    }
}
