import QtQuick 2.2
import Ubuntu.Components 1.1
import Ubuntu.Connectivity 1.0
import "TelegramColors.js" as TelegramColors

Page {
    id: page

    property bool isInSelectionMode: false
    property bool isSearching: false
    property bool isOnline: NetworkingStatus.Online
    property bool isConnected: {
        if (onlineIndicationOnly) {
            return isOnline;
        } else if (root.activeFocus) {
            return isOnline;
        } return true; // Don't indicate connectivity.
    }
    property bool onlineIndicationOnly: false

    property alias title: header.title
    property string pageImage: ""
    property string pageTitle: i18n.tr("TextSecure")
    property string pageSubtitle: ""
    property string pageSubtitleAlt: ""
    property int chatId: 0
    property string firstName: ""
    property string lastName: ""

    property alias body: body.children
    property alias searchTerm: searchField.text
    property alias busy: activityIndicator.running

    signal headerClicked();

    TelegramHeader {
        id: header
        chatId: page.chatId
        chatPhoto: pageImage
        title: pageTitle
        subtitle: pageSubtitleAlt.length > 0 ? pageSubtitleAlt : pageSubtitle
        isConnecting: !page.isConnected
        visible: !isSearching
        width: parent ? parent.width - units.gu(2) : undefined

        onClicked: headerClicked()
    }

    TextField {
        id: searchField
        visible: isSearching
        anchors {
            left: parent.left
            right: parent.right
            rightMargin: units.gu(2)
        }
        inputMethodHints: Qt.ImhNoPredictiveText

        onTextChanged: {
            if (typeof onSearchTermChanged === 'function') {
                onSearchTermChanged(text);
            }
        }
    }

    head {
        id: head
        contents: isSearching ? searchField : header

        backAction: Action {
            id: backAction
            iconName: isInSelectionMode ? "close" : "back"
            onTriggered: {
                if (isSearching) {
                    searchFinished();
                    return;
                }
                if (isInSelectionMode) {
                    selectionCanceled();
                    return;
                }
                back();
                if (typeof onBackPressed === 'function') {
                    onBackPressed();
                }
            }
        }
    }

    Rectangle {
        id: body
        anchors.fill: parent
        // Due to some fancy Page behavior, in fact,
        // this doesn't end up as white anyway..
        color: TelegramColors.page_background
    }

    ActivityIndicator {
        id: activityIndicator
        anchors.centerIn: parent
        z: 101
    }

    Rectangle {
        id: activityBackground
        anchors.fill: parent
        z: 100
        color: "#44000000"
        visible: activityIndicator.running
        MouseArea {
            anchors.fill: parent
            preventStealing: true
            enabled: activityIndicator.running
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

    function back() {
        pageStack.pop();
    }

    signal selectionCanceled()
}
