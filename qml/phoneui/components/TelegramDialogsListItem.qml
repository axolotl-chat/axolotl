import QtQuick 2.0
import Ubuntu.Content 0.1
import Ubuntu.Components 1.1
import Ubuntu.Components.ListItems 1.0 as ListItem
import "listitems"
import "TelegramColors.js" as TelegramColors
import "../js/time.js" as Time

ListItemWithActions {
    property int dialogId: 0
    property int peerId: 0
    property int peerType: 0

    property bool isOutgoing: false
    property bool isSent: false
    property bool isRead: false

    property bool isAction: false
    property string actionType: "a" //TLMessageAction.TypeMessageActionEmpty
    property string actionTitle: ""
    property string actionUser: ""
    property int topMessageId: 0
    property int state: 0

    property bool isGroupChat: false
    property bool isMedia: ContentType.Unknown != mediaType

    property bool isTyping: false
    property string whoIsTyping : ""
    property string message: ""
    property string thumbnail: ""
    property int mediaType: 0

    property string title: ""
    property string senderDisplayName: isGroupChat && !isAction && !isTyping && senderName.length > 0 ? senderName + ": " : ""
    property string senderName: ""
    property int unreadCount: 0
    property string messageDate: ""

    property string subtitle: {
	    if (isMedia) {
		    return delegateUtils.getMediaTypeString(mediaType);
	    } else {
		    return message;
	    }
    }

    id: listitem
    width: parent.width
    height: units.gu(8)

    showDivider: true

    color: TelegramColors.page_background

    Avatar {
        id: imageShape
        anchors {
            top: parent.top
            topMargin: units.dp(1)
            left: parent.left
            leftMargin: units.gu(1)
            bottom: parent.bottom
            bottomMargin: units.dp(1)
            rightMargin: units.gu(1)
        }
        width: height

        chatId: dialogId
        chatTitle: title
        chatPhoto: thumbnail
    }

    Image {
        id: chatTypeIndicator
        anchors {
            top: parent.top
            left: imageShape.right

            topMargin: units.dp(6)
            leftMargin: units.gu(1)
        }
        height: visible ? units.gu(1.8) : 0
        fillMode: Image.PreserveAspectFit

        source: {
            if (isGroupChat) {
                return Qt.resolvedUrl("../images/grouplist.png");
            } else {
                return "";
            }
        }
        visible: isGroupChat
    }

    Text {
        id: timeText
        anchors {
            top: parent.top
            right: parent.right

            margins: units.dp(4)
            rightMargin: units.gu(1)
        }
        height: parent.height/2
        horizontalAlignment: Text.AlignRight
        verticalAlignment: Text.AlignVCenter

        color: TelegramColors.grey
        text: messageDate
    }

    Rectangle {
        id: unreadBox
        anchors {
            right: timeText.right
            verticalCenter: subtitleText.verticalCenter
        }
        width: Math.max(units.gu(2.8), unread.width + units.gu(1))
        height: units.gu(2.5)
        radius: 2

        color: TelegramColors.unread_green
        visible: unread.text != "" && unread.text != "0"

        Text {
            id: unread
            anchors {
                centerIn: parent
                topMargin: units.dp(4)
            }
            horizontalAlignment: TextInput.AlignHCenter
            verticalAlignment: TextInput.AlignVCenter

            font.weight: Font.DemiBold
            font.pixelSize: FontUtils.sizeToPixels("small")
            color: TelegramColors.white
            text: unreadCount
        }
    }

    Image {
        id: sentIndicator
        anchors {
            verticalCenter: timeText.verticalCenter
            right: timeText.left
            rightMargin: units.dp(4)
        }
        width: units.gu(2)
        height: width
        visible: isOutgoing
        z: 1
        fillMode: Image.PreserveAspectFit
        source: {
            if (!isSent) {
                return Qt.resolvedUrl("../images/msg_clock.png");
            } else if (!isRead) {
                return Qt.resolvedUrl("../images/Checks1_2x.png");
            } else {
                return Qt.resolvedUrl("../images/Checks2_2x.png");
            }
        }
    }

    Text {
        id: titleText
        anchors {
            top: parent.top
            left: isGroupChat ? chatTypeIndicator.right : imageShape.right
            leftMargin: isGroupChat ? units.dp(4) : units.gu(1)
            bottom: timeText.bottom
            right: sentIndicator.left
        }
        verticalAlignment: TextInput.AlignVCenter

        font.pixelSize: FontUtils.sizeToPixels("large")
        //font.weight: Font.DemiBold
        color: TelegramColors.black
        elide: Text.ElideRight
        text: title
    }

    Text {
        id: sender
        anchors {
            top: titleText.bottom
            left: imageShape.right
            leftMargin: units.gu(1)
            bottom: parent.bottom
            bottomMargin: units.gu(1)
        }
        verticalAlignment: TextInput.AlignVCenter
        visible: !isAction && isGroupChat

        font.pixelSize: FontUtils.sizeToPixels("medium")
        color: TelegramColors.dark_blue
        text: senderDisplayName
    }

    Text {
        id: subtitleText
        anchors {
            top: titleText.bottom
            left: sender.right
            leftMargin: 0
            bottom: parent.bottom
            bottomMargin: units.gu(1)
            right: unreadBox.left
            rightMargin: units.dp(4)
        }
        horizontalAlignment: TextInput.AlignLeft
        verticalAlignment: TextInput.AlignVCenter

        font.pixelSize: FontUtils.sizeToPixels("medium")
        color: isMedia || isAction || isTyping ? TelegramColors.dark_blue : TelegramColors.grey
        elide: Text.ElideRight
        text: subtitle
    }
}
