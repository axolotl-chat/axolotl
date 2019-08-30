import QtQuick 2.9
import Ubuntu.Components 1.3 as UITK

Item {
    property int connectionState: 0
    id: waitingBar
    anchors {
        top: parent.top
        right: parent.right
        left: parent.left
    }
    height: units.dp(3)

    states: [
        State {
            name: "idle"
            PropertyChanges { target: animation; running: false }
            PropertyChanges { target: waitingBar; visible: false }
        },
        State {
            name: "running"
            PropertyChanges { target: animation; running: true }
            PropertyChanges { target: waitingBar; visible: true }
        }
    ]

    state: connectionState ? "idle" : "running"

    Rectangle {
        id: flyer
        width: parent.width / 4
        height: parent.height
        color: UITK.UbuntuColors.blue

        property var xStart: 0
        property var xEnd: parent.width - width

        SequentialAnimation on x {
            id: animation
            // loops: Animation.Infinite
            onStopped: start() // Workaround for animation length to be updated on screen rotation (width change)

            NumberAnimation {
                from: flyer.xStart; to: flyer.xEnd
                easing.type: Easing.InOutCubic; duration: 1000
            }
            NumberAnimation {
                from: flyer.xEnd; to: flyer.xStart
                easing.type: Easing.InOutCubic; duration: 1400
            }
        }
    }
}
