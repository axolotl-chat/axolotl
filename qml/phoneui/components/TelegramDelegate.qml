import QtQuick 2.0
import Ubuntu.Components 1.1
import Ubuntu.Components.ListItems 0.1 as ListItem
import Ubuntu.Content 0.1
import "listitems"
import "TelegramColors.js" as Color
import "../js/ba-linkify.js" as BaLinkify
import "../js/avatar.js" as Avatar
import "../js/time.js" as Time

ListItemWithActions {
    id: messageDelegate

    property bool outgoing: false
    property bool isSent: false
    property bool isRead: false

    property bool isNewDay: false
    property string newDayText: Time.formatSection(i18n, messageDate * 1000)

    property bool isAction: false
    property string actionTitle: ""
    property string actionUser: ""
    property string actionThumbnail: ""

    property int messageId: 0
    property string message: ""
    property int messageDate: 0
    property int mediaType: 0
    property string senderColor: ""
    // Used in PreviewPage title: From <firstName>
    property int senderId: 0
    property string senderDisplayName: ""
    property string senderName: ""
    property string senderImage: ""
    property string time: ""
    property string thumbnail: ""
    property string attachment: ""
    property string document: ""
    property string documentFileName: ""
    property int documentSize: 0

    property bool isForwarded: false
    property string forwardedFromName: ""

    property bool isVideo: mediaType === ContentType.Videos
    property bool isPhoto: mediaType === ContentType.Pictures
    property bool isAudio: mediaType === ContentType.Music
    property bool isDocument: mediaType == ContentType.All
    property bool isMedia: isVideo || isPhoto || isAudio || isDocument
    property bool isPhotoOrVideo: isVideo || isPhoto || isAudio
    property bool needsDownload: false //isVideo && model.video === "" || isPhoto && model.photo === "" || isDocument && model.document === ""
    property bool isDownloading: false
    property variant progress: undefined

    property string textColor: outgoing ? "white" : "#333333"

    property bool sectionVisible: isNewDay || isAction
    property string sectionText: {
        if (isAction) {
            return delegateUtils.getActionMessageText(
                        outgoing, senderName, actionType,
                        actionUser, actionTitle, actionThumbnail);
        } else if (isNewDay) {
            return newDayText;
        } else return "";
    }

    anchors {
        left: parent ? parent.left : undefined
        right: parent ? parent.right : undefined
    }
    height: isAction || isNewDay ? section.height + internalDelegate.height + units.gu(2)
                                 : internalDelegate.height + units.gu(1)

    signal clicked()
    signal profileImageClicked()
    signal forwardedFromClicked()

    color: Qt.rgba(0, 0, 0, 0)
    selectedColor: Qt.rgba(0, 0, 0, 0.1)

    Column {
        id: column
        anchors {
            left: parent.left
            right: parent.right
            verticalCenter: parent.verticalCenter
        }
        spacing: units.gu(1)

        TelegramSection {
            id: section
            text: sectionText
            image: actionThumbnail
            visible: sectionVisible
        }

        Item {
            id: internalDelegate
            anchors {
                left: parent ? parent.left : undefined
                right: parent ? parent.right : undefined
            }
            height: isAction ? 0 : bubble.height
            visible: !isAction

            Avatar {
                id: imageShape
                anchors {
                    left: parent.left
                    leftMargin: visible ? units.gu(1) : 0
                    bottom: sectionVisible ? bubble.bottom : parent.bottom
                }
                width: visible ? units.gu(3) : 0
                height: width
                visible: !messageDelegate.outgoing

                chatId: senderId
                chatTitle: senderDisplayName
                chatPhoto: senderImage

                onClicked: messageDelegate.profileImageClicked()
            }

            TelegramBubble {
                id: bubble
                outgoing: messageDelegate.outgoing
                anchors {
                    top: parent.top
                    left: outgoing ? undefined : imageShape.right
                    leftMargin: units.gu(2)
                    right: outgoing ? parent.right : undefined
                    rightMargin: units.gu(2)
                }
                height: messageContents.height + units.gu(1)
                width: Math.max(messageLabel.width, senderLabel.width, forwardLabel.width,
                                messageStatusRow.width, loader.width) + (isMedia ? units.gu(1) : units.gu(2))

                Item {
                    id: messageContents
                    anchors {
                        top: parent.top
                        topMargin: units.gu(0.5)
                        left: parent.left
                        leftMargin: units.gu(1)
                        right: parent.right
                        rightMargin: units.gu(1)
                    }
                    height: childrenRect.height

                    Label {
                        id: senderLabel
                        anchors.top: parent.top
                        height: text === "" ? 0 : implicitHeight
                        fontSize: "medium"
                        font.weight: Font.Normal
                        color: senderColor
                        visible: !isPhotoOrVideo
                        elide: Text.ElideRight
                        text: isPhotoOrVideo ? "" : ""

                        Component.onCompleted: {
                            if (senderLabel.paintedWidth > units.gu(28)) {
                                senderLabel.width = units.gu(28);
                            }
                        }
                    }

                    Label {
                        id: forwardLabel
                        anchors.top: senderLabel.bottom
                        visible: isForwarded && !isMedia
                        height: isForwarded ? implicitHeight : 0
                        fontSize: "medium"
                        color: outgoing ? Color.dark_green : Color.blue
                        // FIXME gcollura: find a way to avoid nasty HTML tags
                        text: visible ? i18n.tr("Forwarded from ") + "<b>" + forwardedFromName + "</b>" : ""

                        Component.onCompleted: {
                            if (forwardLabel.paintedWidth > units.gu(28)) {
                                forwardLabel.width = units.gu(28);
                            }
                        }

                        MouseArea {
                            anchors.fill: parent
                            onClicked: {
                                mouse.accepted = true;
                                messageDelegate.forwardedFromClicked();
                            }
                        }
                    }

                    Label {
                        id: messageLabel
                        anchors {
                            top: forwardLabel.bottom
                        }
                        height: paintedHeight
                        width: Math.min(implicitWidth, 0.7 * internalDelegate.width)
                        wrapMode: Text.WrapAtWordBoundaryOrAnywhere
                        fontSize: "small"
                        font.weight: Font.Normal
                        color: textColor
                        text: parseText(message)
                        textFormat: Text.RichText

                        // Taken from messaging-app
                        function parseText(text) {
                            var phoneExp = /(\+?([0-9]+[ ]?)?\(?([0-9]+)\)?[-. ]?([0-9]+)[-. ]?([0-9]+)[-. ]?([0-9]+))/img;
                            // remove html tags
                            text = text.replace(/</g,'&lt;').replace(/>/g,'<tt>&gt;</tt>');
                            // replace line breaks
                            text = text.replace(/(\n)+/g, '<br />');
                            // check for links
                            var htmlText = BaLinkify.linkify(text);
                            if (htmlText !== text) {
                                return htmlText;
                            }
                            // linkify phone numbers if no web links were found
                            return text.replace(phoneExp, '<a href="tel:///$1">$1</a>');
                        }

                        MouseArea {
                            anchors.fill: parent
                            onClicked: {
                                // accept mouse event to avoid unpredictable behaviors
                                // This is a small workaround to avoid bug #1399691,
                                // essentially we're intercepting the mouse click event
                                // let through by ListItemWithActions, see if there's a
                                // link and open it.
                                mouse.accepted = true;
                                var link = messageLabel.linkAt(mouse.x, mouse.y);
                                if (link.length > 0) {
                                    Qt.openUrlExternally(link);
                                }
                            }
                        }
                    }

                    Loader {
                        id: loader
                        anchors {
                            horizontalCenter: parent.horizontalCenter
                            top: senderLabel.bottom
                        }

                        source: {
                            if (isDocument) {
                                return "DocumentDelegate.qml"
                            } else if (isPhotoOrVideo) {
                                return "PhotoVideoDelegate.qml"
                            }
                            return ""
                        }

                        Connections {
                            target: loader.item
                            onClicked: {
                                mouse.accepted = true;
                                if (isDownloading) {
                                    cancelDownload();
                                } else if (needsDownload && isDocument) {
                                    downloadDocument();
                                } else if (needsDownload && isVideo) {
                                    downloadVideo();
                                } else if (needsDownload && isPhoto) {
                                    downloadPhoto();
                                } else {
                                    openPreview();
                                }
                            }
                        }
                    }

                    MessageStatus {
                        id: messageStatusRow
                        anchors {
                            top: isMedia ? undefined : messageLabel.bottom
                            bottom: isMedia ? loader.bottom : undefined
                            bottomMargin: units.dp(2)
                            right: parent.right
                            rightMargin: isPhotoOrVideo ? units.dp(4) : 0
		        }
		        time: messageDelegate.time
                    }
                }

                ProgressBar {
                    anchors {
                        left: parent.left
                        bottom: parent.bottom
                        right: parent.right
                        margins: units.dp(4)
                    }
                    height: units.dp(3)
                    showProgressPercentage: false
                    minimumValue: 0
                    maximumValue: 100
                    visible: isDownloading && isMedia && progress !== 100
                    value: progress
                }
            }
        }
    }

    function openPreview() {
        var properties;
        if (mediaType === ContentType.Pictures) {
            properties = {
                "senderName": senderName,
                "photoPreviewSource": attachment
            };
        } else if (mediaType === ContentType.Videos) {
            properties = {
                "senderName": senderName,
                "videoPreviewSource": attachment
            };
        } else if (mediaType === ContentType.Music) {
            properties = {
                "senderName": senderName,
                "audioPreviewSource": attachment
            };
        } else {
            return
        }
        pageStack.push(previewPage, properties);
    }
}
