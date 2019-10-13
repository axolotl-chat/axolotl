import QtQuick.Controls 2.2
import QtQuick 2.2
import Ubuntu.Components 1.3 as UITK
import QtWebEngine 1.7
import Ubuntu.Components.Popups 1.3 as UITK_Popups
import QtWebSockets 1.0
import QtMultimedia 5.8
import QtQuick.Controls.Suru 2.2
import Ubuntu.Content 1.3

import 'components'


UITK.Page {
  property QtObject request
  property QtObject wsClient
  property var handler
  property var contentType
  property var activeTransfer
  property var selectionType
  property var requestContentHub : false
  id: root
  WebEngineView {
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
      console.log("request: ",request.message)
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
    }

    }
  }
  WebEngineProfile{
    id:webProfile
  }
  ContentPeerPicker {
    id: peerPicker
    anchors { fill: parent;}
    visible: root.requestContentHub
    contentType: root.contentType //ContentType.Pictures
    handler: root.handler //ContentHandler.Source
    // selectionType: root.selectionType

    onPeerSelected: {
        peer.selectionType = root.selectionType
        root.activeTransfer = peer.request()
    }
    onCancelPressed: {
        request.dialogAccept("canceld");
        requestContentHub=false
    }
}

  Connections {
      target: root.activeTransfer
      onStateChanged: {
          if (root.activeTransfer.state === ContentTransfer.Charged)
              requestContentHub=false
              if (root.activeTransfer.items.length > 0) {
                request.dialogAccept(root.activeTransfer.items[0].url);
              }
      }
  }
UITK_Popups.Dialog {
    id: desktopLinkDialog
    property QtObject request
    title: "Add signal desktop"
    text: "Scan QR-Code with Tagger from open store and paste the recieved code here"

    TextField {
      id: desktopId
      height: units.gu(10)
      anchors {
        left: parent.left
        right: parent.right
        topMargin: units.gu(0.1)
      }
      onPressAndHold: {
         desktopId.forceActiveFocus();
         desktopId.text = UITK.Clipboard.data.text ? UITK.Clipboard.data.text : "";

      }
      placeholderText: i18n.tr("tsdevice:/?uuid=...")
    }
    UITK.Button {
      text: "Paste"
      onClicked: {
        desktopId.text = UITK.Clipboard.data.text ? UITK.Clipboard.data.text : "";
      }
    }
    UITK.Button {
      text: "Cancel"
      onClicked: {
        UITK_Popups.PopupUtils.close(desktopLinkDialog)
      }
    }
    UITK.Button {
      text: "Add"
      color: UITK.UbuntuColors.green
      onClicked: {
        request.dialogAccept(desktopId.text)
        UITK_Popups.PopupUtils.close(desktopLinkDialog)
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
