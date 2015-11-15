import QtQuick 2.0
import Ubuntu.Components 1.1
import Ubuntu.Components.ListItems 1.0
import Ubuntu.Components.Popups 1.0
import "../components"
import "../components/listitems"
import "../js/time.js" as Time

TelegramPage {
    id: dialogsPage

    property var messagesToForward: []

    head.actions: [
            Action {
                iconName: "search"
                text: i18n.tr("Search")
                visible: !isSearching
                onTriggered: searchPressed()
            },
            Action {
                iconName: "compose"
                text: i18n.tr("Compose")
                enabled: isConnected
                onTriggered: newChat()
            },
            Action {
                iconName: "contact-group"
                text: i18n.tr("New group")
                enabled: isConnected
                onTriggered: newGroupChat()
            },
            Action {
                iconName: "settings"
                text: i18n.tr("Settings")
                onTriggered: openSettings()
            }
        ]

    head.backAction.visible: isSearching || messagesToForward.length > 0

    body: Item {
        anchors.fill: parent

        Label {
            id: listEmptyLabel
            anchors.centerIn: parent
            fontSize: "medium"
            visible: dialogsListView.count === 0
            text: !isOnline ? i18n.tr("Disconnected")
                            : isConnected ? i18n.tr("No chats") : i18n.tr("Waiting for connection...")
            z: 0
    }

        ListView {
            id: dialogsListView
            anchors {
                top: parent.top
                left: parent.left
                right: parent.right
                bottom: parent.bottom
            }
            clip: true
            z: 1

	    cacheBuffer: units.gu(8)*20
	    model: sessionsModel.len
            delegate: TelegramDialogsListItem {
                id: dialogsListItem
                property var ses: sessionsModel.session(index)
                thumbnail: avatarImage(ses.tel)
                dialogId: uid(ses.tel)
                message: ses.last
                mediaType: ses.cType
                height: visible? (messagesToForward.length > 0 ? 0 : units.gu(8)) : 0
                visible: ses.len > 0

                title: ses.name
                messageDate: ses.when
                isGroupChat: ses.isGroup

                onItemClicked: {
                    mouse.accepted = true;
                    searchFinished();
                    var properties = {};
                    if (messagesToForward.length > 0) {
                        PopupUtils.open(Qt.resolvedUrl("dialogs/ConfirmationDialog.qml"), 
                            dialogsListItem, {
                                text: i18n.tr("Forward message to %1?".arg(title)),
                                onAccept: function() {
                                    properties['messagesToForward'] = messagesToForward;
                                    openChatById(dialogId, ses.tel, properties);
                                    messagesToForward = [];
                                }
                            }
                        );
                    } else {
                        openChatById(dialogsListItem.title, ses.tel, properties);
                        messagesToForward = [];
                    }
                }
            }

            Component.onCompleted: {
                // FIXME: workaround for qtubuntu not returning values depending on the grid unit definition
                // for Flickable.maximumFlickVelocity and Flickable.flickDeceleration
                var scaleFactor = units.gridUnit / 8;
                maximumFlickVelocity = maximumFlickVelocity * scaleFactor;
                flickDeceleration = flickDeceleration * scaleFactor;
            }
        }

        Scrollbar {
            flickableItem: dialogsListView
        }

        DelegateUtils {
            id: delegateUtils
        }
    }

    function onSearchTermChanged(t) {
        textsecure.filterSessions(t)
    }

}
