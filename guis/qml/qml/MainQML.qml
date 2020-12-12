import QtQuick.Controls 2.2
import QtQuick 2.2
import QtWebEngine 1.7
import QtWebSockets 1.0
import QtMultimedia 5.8

    WebEngineView {
    property QtObject request
    property QtObject wsClient
    property var handler
    property var contentType
    property var activeTransfer
    property var selectionType
    property var requestContentHub : false
    id: webView
    profile:webProfile
    url: "http://localhost:9080/"
    settings.showScrollBars: false
    anchors {
      left: parent.left
      top: parent.top
      right: parent.right
      bottom:Qt.inputMethod.visible? showKeyboard.top: parent.bottom
    }
    onJavaScriptConsoleMessage: {
      var msg = "[Axolotl Web View] [JS] (%1:%2) %3".arg(sourceID).arg(lineNumber).arg(message)
      console.log(msg)
    }
  WebEngineProfile{
    id:webProfile
  }

  Connections {
    id: keyboard
    target: Qt.inputMethod
  }
  Item {
    id: showKeyboard
    height: keyboard.target.visible ? keyboard.target.keyboardRectangle.height / (units.gridUnit / 8) : 0
    width: parent.width
    visible:Qt.inputMethod.visible
    anchors {
      left: parent.left
      right: parent.right
      bottom: parent.bottom
    }
    Rectangle {
    color: "white"
      anchors.fill: parent
    }
  }
  Component.onCompleted:{
    webProfile.clearHttpCache();
  }
}
