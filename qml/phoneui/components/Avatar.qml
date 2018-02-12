import QtQuick 2.2
import Ubuntu.Components 0.1
import "../js/avatar.js" as Avatar

Item {
    id: root

    property int chatId: 0
    property bool hasChatId: chatId !== 0
    property string chatTitle: ""
    property string chatPhoto: ""
    property bool hasChatPhoto: chatPhoto !== ""
    property string logo: Qt.resolvedUrl("../images/logo.png")
    property bool showFrame: hasChatId

    signal clicked();

    width: units.gu(4)
    height: units.gu(4)

    // Placed under shape, so it's hidden
    ShaderEffectSource {
        id: source
        anchors.centerIn: parent
        width: 1
        height: 1
        sourceItem: hasChatPhoto ? photoImage : rectImage
    }

    Image {
        id: photoImage
        anchors.centerIn: parent
        width: hasChatPhoto ? parent.width : units.gu(4)
        height: hasChatPhoto ? parent.height : units.gu(4)
        antialiasing: true
        asynchronous: true
        fillMode: Image.PreserveAspectCrop
        source: !hasChatPhoto && !hasChatId ? logo : chatPhoto
        visible: !showFrame
    }

    Rectangle {
        id: rectImage
        anchors.fill: parent
        color: !hasChatPhoto && hasChatId ? Avatar.getColor(chatId) : "#00000000"
        visible: !showFrame && !hasChatPhoto
    }

    Shape {
        id: shape
        image: source
        anchors.fill: parent
        visible: showFrame
    }

    Label {
        id: initialsLabel
        anchors.centerIn: parent
        fontSize: "large"
        color: "white"
        text: !hasChatPhoto && hasChatId ? getInitialsFromTitle(chatTitle) : ""
    }

    MouseArea {
        id: mouseArea
        anchors.fill: parent
        onClicked: {
            mouse.accepted = true;
            root.clicked();
        }
    }

    function getInitialsFromTitle(title) {
        var text = "";
        if (title.length > 0) {
            text = title[0];
        }
        if (title.indexOf(" ") > -1) {
            var lastchar = "";
            for (var a = title.length-1; a > 0; a--) {
                if (lastchar !== "" && title[a] === " ") {
                    break;
                }
                lastchar = title[a];
            }
            text += lastchar;
        }
        return text.toUpperCase();
    }
}
