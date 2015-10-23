import QtQuick 2.0
import Ubuntu.Components 0.1
import "../components"
import Ubuntu.Content 1.1

TelegramPage {

    id: picker

    Component {
        id: resultComponent
        ContentItem {}
    }

    visible: false
    property var url
    property var handler
    property var contentType
    property var curTransfer

    function __exportItems(url) {
        if (picker.curTransfer.state === ContentTransfer.InProgress)
        {
            picker.curTransfer.items = [ resultComponent.createObject(root, {"url": url}) ];
            picker.curTransfer.state = ContentTransfer.Charged;
        }
    }

    ContentPeerPicker {
        visible: parent.visible
        contentType: picker.contentType
        handler: picker.handler
        showTitle: false
        onPeerSelected: {
            picker.curTransfer = peer.request();
            pageStack.pop();
            if (picker.curTransfer.state === ContentTransfer.InProgress)
            picker.__exportItems(picker.url);
        }
        onCancelPressed: {
            pageStack.pop();
        }
    }

    Connections {
        target: picker.curTransfer
        onStateChanged: {
            console.log("curTransfer StateChanged: " + picker.curTransfer.state);
            if (picker.curTransfer.state === ContentTransfer.InProgress) {
                picker.__exportItems(picker.url);
            }
        }
    }
}
