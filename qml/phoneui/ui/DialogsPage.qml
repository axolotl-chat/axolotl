import QtQuick 2.4
import Ubuntu.Components 1.3
import Ubuntu.Components.ListItems 1.0
import Ubuntu.Components.Popups 1.0
import Ubuntu.Connectivity 1.0

import "../components"
import "../components/listitems"
import "../js/time.js" as Time

TelegramPage {
    id: dialogsPage
    focus: !isSearching
    property bool isSecretChat: false
    property bool isConnecting: false
    property bool isOnline: NetworkingStatus.online
    property bool onlineIndicationOnly: false
    property bool isConnected: {
        if (onlineIndicationOnly) {
            return isOnline;
        } else if (root.activeFocus) {
            return isOnline;
        } return true; // Don't indicate connectivity.
    }
    property var messagesToForward: []
    header:PageHeader{
          id:pageHeader
          title: "Signal"
          contents:Item{
            Avatar{
              id: dialogsAvatar
              anchors.leftMargin: 5
              anchors.topMargin: 10
              anchors.left:parent.left
              anchors.top:parent.top
            }
            TextField {
               onActiveFocusChanged: {console.log("focus2") }
                onFocusChanged:{
                  console.log("focus")
                }
                focus: isSearching
                id: searchField
                visible: isSearching
                enabled: isSearching
                activeFocusOnPress: isSearching
                readOnly: !isSearching
                selectByMouse: isSearching
                anchors {
                    left: dialogsAvatar.right
                    rightMargin: units.gu(2)
                    leftMargin: units.gu(2)
                    topMargin: units.gu(1)
                    top: parent.top
                }
                inputMethodHints: Qt.ImhNoPredictiveText

                onTextChanged: {
                    if (typeof onSearchTermChanged === 'function') {
                        onSearchTermChanged(text);
                    }
                }
            }
          }
          trailingActionBar.actions:[
             Action {
                 iconName: "weather-storm"
                 visible: !isConnected
             },
             Action {
                 iconName: "search"
                 text: i18n.tr("Search")
                 visible: !isSearching
                 onTriggered: searchPressed()
             },
             Action {
                 iconName: "address-book-app-symbolic"
                 enabled: isConnected
                 visible: !isSearching
                 onTriggered: newChat()
             },
             Action {
                 iconName: "contact-group"
                 text: i18n.tr("New group")
                 enabled: isConnected
                 visible: !isSearching
                 onTriggered: newGroupChat()
             },
             Action {
                 iconName: "ok"
                 text: i18n.tr("Mark all read")
                 visible: !isSearching
                 enabled: isConnected
                 onTriggered: markAllRead()
             },
             Action {
                 iconName: "settings"
                 text: i18n.tr("Settings")
                 visible: !isSearching
                 onTriggered: openSettings()
             },
             Action {
                 iconName: "help"
                 text: i18n.tr("Help")
                 visible: !isSearching
                 onTriggered: openHelp()
             },
             Action {
                 iconName: "close"
                 text: i18n.tr("Close")
                 visible: isSearching
                 onTriggered: isSearching = false;
             }
          ]
    }

    Rectangle {
      anchors {
        top: pageHeader.bottom
        left: parent.left
        right: parent.right
        bottom: parent.bottom
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
          property var ses: sessionsModel.getSession(index)
          // FIXME Error
          // thumbnail: avatarImage(ses.tel)
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
              mouse.accepted = true;
              // searchFinished();
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
    }
    DelegateUtils {
      id: delegateUtils
    }
  }

    function searchPressed() {
        isSearching = true;
        searchField.forceActiveFocus();
    }
    function searchFinished() {
        if (!isSearching) return;

        isSearching = false;
        searchField.text = "";
    }
    function onSearchTermChanged(t) {
        textsecure.filterSessions(t)
    }

}
