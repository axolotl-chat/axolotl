import QtQuick 2.4
import Ubuntu.Components 0.1
import "TelegramColors.js" as Color

Item {
    id: item
    property alias text: sectionLabel.text
    property string image: ""

    width: parent.width
    height: labelRect.height + sectionImage.height
            + (image !== "" ? units.gu(1) : 0)

    Rectangle {
        id: labelRect
        anchors.horizontalCenter: parent.horizontalCenter
        width: sectionLabel.paintedWidth + units.gu(2)
        height: sectionLabel.height + units.gu(1)

        color: Color.chat_section
        radius: 3

        Text {
            id: sectionLabel
            anchors.centerIn: parent
            horizontalAlignment: Text.AlignHCenter
            width: item.width - units.gu(4)

            color: Color.white
            font.weight: Font.DemiBold
            font.pixelSize: FontUtils.sizeToPixels("medium")
            elide: Text.ElideRight
            wrapMode: Text.WordWrap
            text: section
        }

    }

    Image {
        id: sectionImage
        anchors {
            top: labelRect.bottom
            topMargin: units.gu(1)
            horizontalCenter: parent.horizontalCenter
        }
        width: visible ? units.gu(12) : 0
        height: width
        fillMode: Image.PreserveAspectCrop
        visible: source != ""
        source: image
    }
}
