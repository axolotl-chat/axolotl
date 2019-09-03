import QtQuick 2.9
import QtQuick.Layouts 1.1
import Ubuntu.Components 1.3
import Ubuntu.Components.Popups 1.3
import QtMultimedia 5.4
import Qt.labs.settings 1.0
import QtWebEngine 1.5
import QtQuick.Window 2.2


/* =============================== MAIN.qml ===============================
This file is the start point of the app. It contains all important config variables,
instances of all controller, the layout (mainLayout) and the start point.
*/

MainView {

    /* =============================== MAIN CONFIGS ===============================
    */
    id: root
    objectName: 'mainView'
    applicationName: 'textsecure.nanuc'
    automaticOrientation: true

    // automatically anchor items to keyboard that are anchored to the bottom
    anchorToKeyboard: true

    width: units.gu(125)
    height: units.gu(75)
    Component.onCompleted: {
      console.log("height", Screen.desktopAvailableHeight )
      console.log("width", Screen.desktopAvailableWidth )
    }
    WebEngineView {
        id: webview
        anchors.fill: parent
        url:            "https://google.com/"
      // zoomFactor: units.gu(1) / 8

      // userScripts: [
      //       WebEngineScript {
      //           sourceUrl: Qt.resolvedUrl("./viewport.js")
      //           runOnSubframes: true
      //       }
      //   ]
        onJavaScriptConsoleMessage: {
            var msg = "[Axolotl Web View] [JS] (%1:%2) %3".arg(sourceID).arg(lineNumber).arg(message)
            console.log(msg)
        }


  }
}
