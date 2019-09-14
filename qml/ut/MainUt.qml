import QtQuick.Controls 2.2
import QtQuick 2.2
import Ubuntu.Components 1.3 as UITK
import QtWebEngine 1.7
import Ubuntu.Components.Popups 1.3 as UITK_Popups
import QtWebSockets 1.0
import QtMultimedia 5.8
import QtQuick.Controls.Suru 2.2

import 'components'


UITK.Page {
  property QtObject requestTest
  property QtObject wsClient
  id: root
  WebSocketServer {
    id: server
    listen: true
    port: 12345

    onClientConnected: {
      console.log('BlobSaver: websocket server connected');
      wsClient = webSocket
      webSocket.onTextMessageReceived.connect(function(base64data) {
        console.log('base64data');
        var path = BlobSaver.write(base64data);
        fileDownloaded(path);
      });
    }

    onErrorStringChanged: {
      console.log('BlobSaver: websocket server error', errorString);
    }
  }
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
      desktopLinkDialog.request = request; // keep the reference to the request
      desktopLinkDialog.visible = true;
      request.dialogAccept(desktopId.text)
      requestTest = request;
    }
  }
  WebEngineProfile{
    id:webProfile
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
        console.log(requestTest)
        requestTest.dialogReject()
        UITK_Popups.PopupUtils.close(desktopLinkDialog)
      }
    }
    UITK.Button {
      text: "Add"
      color: UITK.UbuntuColors.green
      onClicked: {
        wsClient.sendTextMessage(desktopId.text)
        requestTest.dialogAccept()
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
