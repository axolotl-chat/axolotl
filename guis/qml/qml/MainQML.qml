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
    //property string url
    //title: "Axolotl"
    id: webView
    profile:webProfile
    url: "http://localhost:9080/"
    // url: "https://google.de"
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
    onJavaScriptDialogRequested: function(request) {
      request.accepted = true;
      console.log("[axolotl ut] request: ",request.message)
      if(request.message =="desktopLink"){
        desktopLinkDialog.request = request; // keep the reference to the request
        desktopLinkDialog.visible = true;
        root.request = request;
      } else if(request.message =="refreshContacts"){
        root.request = request
        root.requestContentHub = true
        root.contentType = ContentType.Contacts
        root.handler = ContentHandler.Source
        root.selectionType = ContentTransfer.Multiple
      } else if(request.message =="photo"){
        root.request = request
        root.requestContentHub = true
        root.contentType = ContentType.Pictures
        root.handler = ContentHandler.Source
        root.selectionType = ContentTransfer.Single
      } else if(request.message =="video"){
        root.request = request
        root.requestContentHub = true
        root.contentType = ContentType.Videos
        root.handler = ContentHandler.Source
        root.selectionType = ContentTransfer.Single
      } else if(request.message =="document"){
        root.request = request
        root.requestContentHub = true
        root.contentType = ContentType.Documents
        root.handler = ContentHandler.Source
        root.selectionType = ContentTransfer.Single
      } else if(request.message =="audio"){
        root.request = request
        root.requestContentHub = true
        root.contentType = ContentType.Music
        root.handler = ContentHandler.Source
        root.selectionType = ContentTransfer.Single
      } else if(request.message.includes("[oC]")){
        root.request = request
        root.requestContentHub = true
        root.url = request.message.substring(4)
        root.contentType = ContentType.Contacts
        root.handler = ContentHandler.Destination
        root.selectionType = ContentTransfer.Multiple
      } else if(request.message.includes("[oP]")){
        root.request = request
        root.requestContentHub = true
        root.url = request.message.substring(4)
        root.contentType = ContentType.Pictures
        root.handler = ContentHandler.Destination
        root.selectionType = ContentTransfer.Single
      } else if(request.message.includes("[oV]")){
        root.request = request
        root.requestContentHub = true
        root.url = request.message.substring(4)
        root.contentType = ContentType.Videos
        root.handler = ContentHandler.Destination
        root.selectionType = ContentTransfer.Single
      } else if(request.message.includes("[oD]")){
        root.request = request
        root.requestContentHub = true
        root.url = request.message.substring(4)
        root.contentType = ContentType.Documents
        root.handler = ContentHandler.Destination
        root.selectionType = ContentTransfer.Single
      } else if(request.message.toLowerCase().includes("call")){
        root.request = request
        var callUrl = request.message.substring(4)
        Qt.openUrlExternally("tel:///"+callUrl);
        request.dialogAccept();

      } else if(request.message.toLowerCase().includes("http")){
          Qt.openUrlExternally(request.message);
          request.dialogAccept();
      } else{
      simpleDialog.request = request;
      simpleDialog.visible = true;
      root.request = request;

    }
  }
  WebEngineProfile{
    id:webProfile
  }

  Connections {
      target: root.activeTransfer
      onStateChanged: {
        if(handler === ContentHandler.Source ){
          if (root.activeTransfer.state === ContentTransfer.Charged)
              requestContentHub=false
              if (root.activeTransfer.items.length > 0) {
                request.dialogAccept(root.activeTransfer.items[0].url);
              }
            }
      }
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
      // bottomMargin: (UbuntuApplication.inputMethod.visible) ? -height : 0
      // onBottomMarginChanged: hidden = (anchors.bottomMargin == -height ? true : hidden)
      left: parent.left
      right: parent.right
      bottom: parent.bottom
      // Behavior on bottomMargin {
      //   NumberAnimation {
      //     duration: UbuntuAnimation.FastDuration
      //     easing: UbuntuAnimation.StandardEasing
      //   }
      // }
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
