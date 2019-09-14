import QtQuick 2.9
import QtWebEngine 1.7
import QtQuick.Window 2.13
import QtWebSockets 1.0
Window {
  width:400
  height:600
  id:root
  property QtObject wsClient
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
  WebEngineView  {
    anchors {
      top: parent.top
      left: parent.left
      right: parent.right
      bottom: parent.bottom
    }
      id: webView
      url: "http://localhost:9080/"
      // url: "https://google.de"
      settings.showScrollBars: false
      onJavaScriptConsoleMessage: {
          var msg = "[Axolotl Web View] [JS] (%1:%2) %3".arg(sourceID).arg(lineNumber).arg(message)
          console.log(msg)
      }
  }
}
