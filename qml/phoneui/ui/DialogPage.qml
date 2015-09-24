import QtQuick 2.0
import Ubuntu.Components 1.1
import Ubuntu.Components.Popups 0.1
import Ubuntu.Components.ListItems 1.0 as ListItems
import Ubuntu.Content 1.0

import "../components"
import "../components/TelegramColors.js" as TelegramColors
import "../js/avatar.js" as Avatar
import "../js/time.js" as Time

TelegramPage {
    id: dialogPage

    title: messagesModel.name

    property bool isChat: messagesModel.isGroup

    property var messagesToForward: []

    onHeaderClicked: {
        Qt.inputMethod.hide();

        var userId = uid(messagesModel.tel)
        isChat ? openGroupProfile(userId)
               : openProfile(userId);
    }

    property list<Action> defaultActions: [
        Action {
            iconName: "stock_contact"
            text: isChat ? i18n.tr("Group Info") : i18n.tr("Profile Info")
            onTriggered: {
                Qt.inputMethod.hide();
                headerClicked();
            }
        }
    ]

    property list<Action> selectionActions: [
        Action {
            iconName: "select"
            text: i18n.tr("Select all")
            onTriggered: {
                if (list.selectedItems.count === list.listModel) {
                    list.clearSelection()
                    list.selectionClear()
                } else {
                    list.selectAll()
                    list.selectionAll()
                }
            }
        },
        Action {
            iconName: "info"
            text: i18n.tr("Message Info")
            visible: list.selectedItems.count === 1
            onTriggered: {
                    console.log(messagesModel.message(list.sela[0]).info())
            }
        },
        Action {
            id: copySelectedAction
            iconName: "edit-copy"
            text: i18n.tr("Copy")
            onTriggered: list.copySelected()
        },
        Action {
            id: forwardSelectedAction
            iconName: "next"
            text: i18n.tr("Forward")
            visible: true
            onTriggered: list.forwardSelected()
        },
        Action {
            id: multiDeleteAction
            iconName: "delete"
            text: i18n.tr("Delete")
            onTriggered: list.deleteSelected()
        }
    ]

    head.actions: list.isInSelectionMode ? selectionActions : defaultActions

    isInSelectionMode: list.isInSelectionMode
    onSelectionCanceled: {
            list.cancelSelection()
            list.selectionClear()
    }

    body: Item {

        anchors {
            fill: parent
        }

        Image {
            anchors.fill: parent

            Component.onCompleted: {
                // Set size here, so we don't rescale on input method.
                sourceSize.height = parent.height;
                sourceSize.width = height * sourceSize.width / sourceSize.height;
            }
        }

        Component {
            id: attachPopovercomponent

            ActionSelectionPopover {
                id: attachPopover
                contentWidth: units.gu(22)
                focus: false
                z: 3
                delegate: ListItems.Standard {
                    iconFrame: false
                    iconSource: Qt.resolvedUrl(action.iconSource)
                    focus: false
                    text: action.text
                }
                actions: ActionList {
                    Action {
                        iconSource: "../images/ic_attach_gallery.png"
                        text: i18n.tr("Photo")
                        onTriggered: {
                            attachPopover.hide();
                            requestMedia(ContentType.Pictures);
                        }
                    }
                    Action {
                        iconSource: "../images/ic_attach_video.png"
                        text: i18n.tr("Video")
                        onTriggered: {
                            message.forceActiveFocus();
                            attachPopover.hide();
                            requestMedia(ContentType.Videos);
                        }
                    }
                    Action {
                        iconSource: "../images/ic_ab_doc.png"
                        text: i18n.tr("File")
                        onTriggered: {
                            message.forceActiveFocus();
                            attachPopover.hide();
                            requestMedia(ContentType.All);
                        }
                    }
                }
            }
        }

        Rectangle {
            id: bottomRectangle
            anchors {
                left: parent.left
                right: parent.right
                bottom: parent.bottom
            }
            height: message.height + units.gu(2)
            z: 2

            color: "white"

            TextArea {
                id: message

                property int oldLength: 0

                anchors {
                    left: parent.left
                    right: sendButtonBox.left
                    bottom: parent.bottom
                    margins: units.gu(1)
                }

                // this value is to avoid letter and underline being cut off
                height: units.gu(4.3)
                enabled: isConnected
                visible: true
                placeholderText: isConnected ? i18n.tr("Type message") : i18n.tr("Not connected.")
                inputMethodHints: Qt.ImhNone

                autoSize: true
                maximumLineCount: 4
                Keys.onReturnPressed: {
                    if (settingsModel.sendByEnter && isConnected) {
                        Qt.inputMethod.commit();
                        if (message.text.length === 0) return;
                        sendMessage(message.text);
                    } else {
                        event.accepted = false;
                    }
                }
		Component.onCompleted: {
				forceActiveFocus();
		}

            }

            Item {
                id: sendButtonBox
                anchors {
                    top: parent.top
                    bottom: parent.bottom
                    right: parent.right
                    rightMargin: units.gu(2)
                }
                width: buttonsRow.width
                visible: true

                Row {
                    id: buttonsRow
                    spacing: units.gu(2)
                    anchors.verticalCenter: sendButtonBox.verticalCenter

                    Image {
                        id: attachButton
                        height: units.gu(3.5)
                        fillMode: Image.PreserveAspectFit
                        focus: false
                        enabled: isConnected && message.text.length === 0
                        source: "../images/ic_ab_attach.png"

                        states: [
                            State {
                                name: "text"
                                when: message.text.length > 0
                                PropertyChanges {
                                    target: attachButton
                                    width: 0
                                    opacity: 0.0
                                }
                            },
                            State {
                                name: "notext"
                                when: message.text.length  === 0
                                PropertyChanges {
                                    target: attachButton
                                    width: units.gu(3.5)
                                    opacity: 1.0
                                }
                            }
                        ]

                        transitions: [
                            Transition {
                                PropertyAnimation {
                                    target: attachButton
                                    properties: "width, opacity"
                                    duration: UbuntuAnimation.FastDuration
                                }
                            }
                        ]

                        MouseArea {
                            anchors.fill: attachButton
                            onClicked: {
                                if (isConnected) {
                                    message.focus = false;
                                    Qt.inputMethod.hide();
                                    PopupUtils.open(attachPopovercomponent, attachButton)
                                }
                            }
                        }
                    }

                    Image {
                        id: sendButton
                        height: units.gu(3.5)
                        width: units.gu(3.5)
                        fillMode: Image.PreserveAspectFit
                        focus: false
                        enabled: isConnected
                        source: enabled ? "../images/ic_send.png" : "../images/ic_send_disabled.png"

                        MouseArea {
                            anchors.fill: sendButton
                            onClicked: {
                                Qt.inputMethod.commit();
                                if (message.text.length === 0) return;

                                sendMessage(message.text);
                                list.positionViewAtBeginning();
                            }
                        }
                    }
                }
            }
        }

        MultipleSelectionListView {
            id: list
            property string sels
            property var sela: []
            verticalLayoutDirection: ListView.BottomToTop
            anchors {
                top: parent.top
                left: parent.left
                bottom: bottomRectangle.top
                right: parent.right
            }
            header: Item {
                height: units.gu(2)
            }
            footer: Item {
                height: units.gu(1)
            }
            cacheBuffer: units.gu(10)*20
            highlightFollowsCurrentItem: false
            clip: true

            listModel: messagesModel.len
            listDelegate: TelegramDelegate {
                id: delegate
		property int ii: messagesModel.len - 1 - index
                outgoing: messagesModel.message(ii).outgoing
		isAction: false
                isSent: messagesModel.message(ii).isSent
                isRead: messagesModel.message(ii).isRead
		/*
                actionType: model.actionType
                actionTitle: model.actionTitle
                actionUser: model.actionUser
                actionThumbnail: isAction ? model.thumbnail : ""

                //FIXME: When section headers upstream bug is resolved revet this with section headers.
                isNewDay: (index === list.count) || !Time.areSameDay((messagesModel.get(index+1).date * 1000),model.date*1000)

		messageId: model.id
		*/
                message: messagesModel.message(ii).text
                time: messagesModel.message(ii).hTime
                senderId: uid(messagesModel.tel)
                senderName: outgoing? "You" : messagesModel.message(ii).from
                senderDisplayName: outgoing ? "" : senderName
                mediaType: messagesModel.message(ii).cType
                thumbnail: mediaType === ContentType.Pictures? "image://ts/"+messagesModel.tel+":"+ii:""
		/*
                senderColor: Avatar.getColor(model.fromId)
                senderImage: {
                    if (isChat && !outgoing && !list.isInSelectionMode) {
                        return model.fromThumbnail;
                    }
                    return "";
                }
                photo: model.photo
                video: model.video
                document: model.document
                documentFileName: model.documentFileName
                documentSize: model.documentSize
                isDownloading: model.downloading || false
                forwardedFromName: getForwardedFromName();
		*/
                progress: model.downloadedPercentage || 0

                isForwarded: model.fwdFromId > 0
                onProfileImageClicked: {
                    openProfile(model.fromId);
                }

                onForwardedFromClicked: {
                    openProfile(model.fwdFromId);
                }

                leftSideActions: [
                    Action {
                        iconName: "delete"
                        text: i18n.tr("Delete")
                        visible: isConnected
                        onTriggered: {
                        }
                    }
                ]
                rightSideActions: [
                     Action {
                        iconName: "reload"
                        text: i18n.tr("Resend")
                        visible: !isSent && !isAction
                        onTriggered: {
                            if (isPhoto) {
                                sendAttachment(photo);
                            } else if (isVideo) {
                                sendAttachment(video);
                            } else {
                                sendMessage(message);
                            }
                        }
                    },
                    Action {
                        iconName: "edit-copy"
                        text: i18n.tr("Copy")
                        visible: !isPhoto && !isVideo && !isDocument
                        onTriggered: Clipboard.push(message)
                    },
                    Action {
                        iconName: "next"
                        text: i18n.tr("Forward")
                        visible: true
                        onTriggered: forwardMessages([messageId])
                    }
                ]

                selected: list.isSelected(delegate)
                selectionMode: isAction ? false : list.isInSelectionMode

                onItemPressAndHold: {
                    list.startSelection()
                    if (list.isInSelectionMode) {
                        list.selectionToggled(ii)
                        list.selectItem(delegate)
                    }
                }

                onItemClicked: {
                    if (list.isInSelectionMode && !isAction) {
                        list.selectionToggled(ii)
                        if (selected) {
                            list.deselectItem(delegate)
                        } else {
                            list.selectItem(delegate)
                        }
                    }
                }

                locked: !isConnected || isAction
            }

            function selectionToggled(index) {
                var a = list.sela
                var i = a.indexOf(index)
                if (i == -1) {
                    a.push(index)
                } else {
                    a.splice(i, 1)
                }

                a.sort(function(a,b) {return a-b})
                list.sels = a.join(":")
            }

            function selectionClear() {
                list.sela = []
            }

            function selectionAll() {
                list.sela = []
                for (var i = 0; i < messagesModel.len; i++) {
                    list.sela.push(i)
                }
            }

            function copySelected() {
                var message = "", item;
                for (var i = list.selectedItems.count - 1; i >= 0 ; i--) {
                    item = list.selectedItems.get(i);
                    if (delegateUtils.getMediaTypeString(item.model.mediaType) === "Text") {
                        message += item.model.text + "\n";
                    }
                }
                list.endSelection();
                Clipboard.push(message);
            }

            function forwardSelected() {
                var toForward = [];
                for (var i = list.selectedItems.count - 1; i >= 0 ; i--) {
                    toForward.push(list.selectedItems.get(i).model.id);
                }
                list.endSelection();
                forwardMessages(toForward);
            }

            Scrollbar {
                flickableItem: list
            }
        }

        Item {
            anchors.centerIn: parent
            width: pageLabel.width
            height: pageLabel.height

            Rectangle {
                anchors {
                    fill: pageLabel
                    margins: units.gu(-2)
                }
                radius: 5
                color: TelegramColors.chat_section
            }

            Label {
                id: pageLabel
                color: "white"
                fontSize: "medium"
                text: {
                        return i18n.tr("No messages here yet...");
                }
            }

            visible: list.model.count === 0
            z: 1
        }

        DelegateUtils {
            id: delegateUtils
        }
    }

    MediaImport {
        id: mediaImporter

        onMediaReceived: {
            var filePath = String(mediaUrl).replace('file://', '');
            dialogPage.sendAttachment(filePath);
            message.forceActiveFocus();
        }
    }

    function requestMedia(mediaType) {
        Qt.inputMethod.hide();
        mediaImporter.contentType = mediaType;
        mediaImporter.requestMedia();
    }

    function sendMessage(text) {
        if (text.length === 0) return;
        textsecure.sendMessage(messagesModel.tel, text);
        message.text = "";
    }

    Timer {
        id: sendAttachmentTimer

        property int attempts: 1
        property string path: ""

        function send(attachmentPath) {
            busy = true
            stop();
            path = attachmentPath;
            attempts = 0;
            restart();
        }

        repeat: false
        onTriggered: {
                textsecure.sendAttachment(messagesModel.tel, message.text, path)
                message.text = "";
                busy = false
                stop();
        }
    }

    function sendAttachment(path) {
	    console.log("Sending attachment", path)
	    sendAttachmentTimer.send(path)
    }

}
