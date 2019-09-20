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
  // WebSocketServer {
  //   id: server
  //   listen: true
  //   port: 12345
  //
  //   onClientConnected: {
  //     console.log('BlobSaver: websocket server connected');
  //     wsClient = webSocket
  //     webSocket.onTextMessageReceived.connect(function(base64data) {
  //       console.log('base64data');
  //       var path = BlobSaver.write(base64data);
  //       fileDownloaded(path);
  //     });
  //   }
  //
  //   onErrorStringChanged: {
  //     console.log('BlobSaver: websocket server error', errorString);
  //   }
  // }
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
    Component.onCompleted: {
      webView.runJavaScript('ut = "ut"');
      webView.runJavaScript('console.log(ut)');
    }
    onJavaScriptDialogRequested: function(request) {
      request.accepted = true;
      console.log(request.message)
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
    showTitle: false
    contentType: root.contentType //ContentType.Pictures
    handler: root.handler //ContentHandler.Source
    // selectionType: root.selectionType

    onPeerSelected: {
        peer.selectionType = root.selectionType
        root.activeTransfer = peer.request()
    }
    onCancelPressed: {
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
    text: "Scan QR-Code with other app and paste it here"

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
      text: "cancel"
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
  Item {
    id: showKeyboard
    property var kHeight: root.height * (root.landscape ? root.tablet ? 0.34 : 0.49 :
        root.tablet ? 0.31 : 0.40)
      height: kHeight
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
    Component.onCompleted: {
      console.log( root.height)
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
