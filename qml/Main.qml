import QtQuick 2.2
import QtWebEngine 1.7

WebEngineView  {
    id: webView
    url: "http://[::1]:9080/"
    onJavaScriptConsoleMessage: {
        var msg = "[Axolotl Web View] [JS] (%1:%2) %3".arg(sourceID).arg(lineNumber).arg(message)
        console.log(msg)
    }
}
