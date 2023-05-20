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
    onLoadingChanged:{
      // interceptor to make the registration captcha work
      var msg = "[Axolotl Web View] [JS] url changed %1".arg(url)
      console.log(msg)
      if (url == "https://signalcaptchas.org/registration/generate.html"){
        console.log("run interceptor")
        var  interceptor = `
        // override the default onload function
			      document.addEventListener('DOMContentLoaded', function () {
                    console.log("DOMContentLoaded");
                    window.renderCallback = function (scheme, sitekey, action, token) {
                    
                        var targetURL = "tauri://localhost/?token=" + [scheme, sitekey, action, token].join(".");
                        var link = document.createElement("a");
                        link.href = targetURL;
                        link.innerText = "open axolotl";
                    
                        document.body.removeAttribute("class");
                        setTimeout(function () {
                        document.getElementById("container").appendChild(link);
                        }, 2000);
                    
                        window.location.href = targetURL;
                    };
                    window.intercept = function() {
                        console.log("intercept")
                        console.log("resetting captcha")
                        document.getElementById("captcha").innerHTML = "";
                        if(useHcaptcha)onloadHcaptcha();
                        else onload();
                      }
                    if (!window.location.href.includes("localhost")){
                        intercept();
                    } else {
                        console.log("localhost detected, not intercepting");
                    }
            });
				`
        webView.runJavaScript(interceptor);
        console.log("run interceptor")
      }

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
