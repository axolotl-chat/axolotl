import QtQuick 2.9
import Ubuntu.Components 1.3
import QtWebEngine 1.7

MainView {
  applicationName: "textsecure.nanuc"

  automaticOrientation: false

  anchorToKeyboard: true

  WebEngineView  {
    anchors.fill: parent
      id: webView
      url: "http://localhost:9080/axolotl"
      // url: "https://google.de"
      settings.showScrollBars: false
      onJavaScriptConsoleMessage: {
          var msg = "[Axolotl Web View] [JS] (%1:%2) %3".arg(sourceID).arg(lineNumber).arg(message)
          console.log(msg)
      }
  }
}
