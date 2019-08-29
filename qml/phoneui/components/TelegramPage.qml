import QtQuick 2.4
import Ubuntu.Components 1.3
import Ubuntu.Connectivity 1.0
import "TelegramColors.js" as TelegramColors

Page {
    id: page

    property bool isInSelectionMode: false
    property bool isSearching: false
    property bool isOnline: NetworkingStatus.online
    property bool isConnected: {
        if (onlineIndicationOnly) {
            return true;
        } else if (root.activeFocus) {
            return true;
            // return isOnline;
        } return true; // Don't indicate connectivity.
    }
    property bool onlineIndicationOnly: false

    // property alias title: telegramheader.title
    property string pageImage: ""
    property string pageTitle: i18n.tr("Signal")
    property string pageSubtitle: ""
    property string pageSubtitleAlt: ""
    property int chatId: 0
    property string firstName: ""
    property string lastName: ""

    property alias body: body.children
    property alias busy: activity.running

    signal headerClicked();




    Rectangle {
        id: body
        // anchors.fill: parent
        anchors {
            top: header.bottom
            left: parent.left
            right: parent.right
            bottom: parent.bottom
        }
        // Due to some fancy Page behavior, in fact,
        // this doesn't end up as white anyway..
        color: TelegramColors.page_background
    }

    ActivityIndicator {
        id: activity
        anchors.centerIn: parent
    }

    Rectangle {
        id: activityBackground
        anchors.fill: parent
        z: 100
        color: "#44000000"
        visible: activity.running
        MouseArea {
            anchors.fill: parent
            preventStealing: true
            enabled: activity.running
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
