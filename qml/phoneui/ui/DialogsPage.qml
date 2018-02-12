import QtQuick 2.4
import Ubuntu.Components 1.3
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
                iconName: "ok"
                text: i18n.tr("Mark all read")
                enabled: isConnected
                onTriggered: markAllRead()
            },
            Action {
                iconName: "settings"
                text: i18n.tr("Settings")
                onTriggered: openSettings()
            },
            Action {
                iconName: "help"
                text: i18n.tr("Help")
                onTriggered: openHelp()
            }
        ]

    head.backAction.visible: isSearching || messagesToForward.length > 0

    body: Item {
        anchors.fill: parent

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
                property var ses: sessionsModel.getSession(index)
                // FIXME Error
                thumbnail: avatarImage(ses.tel)
                dialogId: uid(ses.tel)
                message: ses.last
                unreadCount: ses.unread
                mediaType: ses.cType
                height: visible? (messagesToForward.length > 0 ? 0 : units.gu(8)) : 0
                visible: ses.len > 0

                title: ses.name
                messageDate: ses.when
                isGroupChat: ses.isGroup

                onItemClicked: {
                    console.log("dialogsPageClick");
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
        textsecure.FilterSessions(t)
    }

}
