import QtQuick 2.0
import Ubuntu.Components 1.1
import Ubuntu.Components.ListItems 1.0 as ListItem
import "listitems"
import "../js/avatar.js" as Avatar
import "../js/time.js" as Time
import "TelegramColors.js" as TelegramColors

ListItemWithActions {
    id: listitem

    property int userId: 0 // to get avatar if thumbnail not set
    property string photo: ""
    property string title: ""
    property bool isOnline: false
    property int lastSeen: 0

    width: parent.width
    height: units.gu(8)
    showDivider: true
    color: listitem.pressed ? TelegramColors.list_pressed : TelegramColors.white

    Avatar {
        id: imageShape
        anchors {
            top: parent.top
            topMargin: units.dp(4)
            left: parent.left
            leftMargin: units.gu(2)
            bottom: parent.bottom
            bottomMargin: units.dp(4)
            rightMargin: units.gu(1)
        }
        width: height

        chatId: listitem.userId
        chatTitle: listitem.title
        chatPhoto: listitem.photo
    }

    Text {
        id: titleText
        width: implicitWidth
        anchors {
            top: parent.top
            topMargin: units.gu(1)
            left: imageShape.right
            leftMargin: units.gu(1.5)
            right: parent.right
            rightMargin: units.gu(1.5)
        }
        verticalAlignment: TextInput.AlignVCenter

        font.pixelSize: FontUtils.sizeToPixels("large")
        color: TelegramColors.black
        elide: Text.ElideRight
        text: title
    }

}
