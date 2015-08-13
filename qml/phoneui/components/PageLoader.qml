import QtQuick 2.0

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
