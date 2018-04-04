import QtQuick 2.4

Loader {
    property bool valid: item !== null

    active: false
    asynchronous: false
    anchors {
        left: parent.left
        bottom: parent.bottom
        right: parent.right
    }
}
