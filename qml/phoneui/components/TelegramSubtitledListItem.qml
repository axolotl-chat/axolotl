import QtQuick 2.4
import Ubuntu.Components 1.3
import Ubuntu.Components.ListItems 0.1 as ListItem

import "TelegramColors.js" as TelegramColors

ListItem.Empty {
    id: item
    width: parent.width
    height: units.gu(8)

    property string title: ""
    property string titleColor: TelegramColors.black
    property bool titleIsBold: false
    property int titleMaxLineCount: 1
    property string subtitle: ""
    property string subtitleColor: TelegramColors.grey
    property bool subtitleIsBold: false

    Rectangle {
        id: background
        anchors.fill: parent
        color: item.selected ? TelegramColors.blue : TelegramColors.transparent
    }

    Text {
        id: title
        anchors {
            top: parent.top
            topMargin: units.gu(1)
            left: parent.left
            leftMargin: units.gu(2)
            right: parent.right
            rightMargin: units.gu(2)
        }
        width: parent.width
        verticalAlignment: TextInput.AlignVCenter

        maximumLineCount: item.titleMaxLineCount
        wrapMode: "WordWrap"

        elide: Text.ElideRight
        font.pixelSize: FontUtils.sizeToPixels("large")
        font.weight: item.titleIsBold ? Font.Bold : Font.Light
        color: item.titleColor
        text: item.title
    }

    Text {
        id: subtitle
        anchors {
            top: title.bottom
            topMargin: units.dp(4)
            left: title.left
            bottom: parent.bottom
            bottomMargin: units.gu(1)
            right: title.right
        }
        width: parent.width
        verticalAlignment: TextInput.AlignVCenter

        elide: Text.ElideRight
        font.pixelSize: FontUtils.sizeToPixels("medium")
        font.weight: item.subtitleIsBold ? Font.Bold : Font.Light
        color: item.subtitleColor
        text: item.subtitle
    }
}
