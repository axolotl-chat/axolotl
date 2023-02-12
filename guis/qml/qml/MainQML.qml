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
        var  interceptor = `
        // override the default onload function
			            document.addEventListener('DOMContentLoaded', function () {
                window.renderCallback = function (scheme, sitekey, action, token) {
                    var targetURL = "http://localhost:9080/register?token=" + [scheme, sitekey, action, token].join(".");
                    var link = document.createElement("a");
                    link.href = targetURL;
                    link.innerText = "open axolotl";
                
                    document.body.removeAttribute("class");
                    setTimeout(function () {
                    document.getElementById("container").appendChild(link);
                    }, 2000);
                
                    window.location.href = targetURL;
                };
                function onload() {
                    alert("onload");
                    var action = document.location.href.indexOf("challenge") !== -1 ?
                      "challenge" : "registration";
                    var isDone = false;
                    var sitekey = "6LfBXs0bAAAAAAjkDyyI1Lk5gBAUWfhI_bIyox5W";
                  
                    var widgetId = grecaptcha.enterprise.render("captcha", {
                      sitekey: sitekey,
                      size: "checkbox",
                      theme: getTheme(),
                      callback: function (token) {
                        isDone = true;
                        renderCallback("signal-recaptcha-v2", sitekey, action, token);
                      },
                    });
                  
                    function execute() {
                      if (isDone) {
                        return;
                      }
                  
                      grecaptcha.enterprise.execute(widgetId, { action: action });
                  
                      // Below, we immediately reopen if the user clicks outside the widget. If they
                      //   close it some other way (e.g., by pressing Escape), we force-reopen it
                      //   every second.
                      setTimeout(execute, 1000);
                    }
                  
                    // If the user clicks outside the widget, reCAPTCHA will open it, but we'll
                    //   immediately reopen it. (We use onclick for maximum browser compatibility.)
                    document.body.onclick = function () {
                      if (!isDone) {
                        grecaptcha.enterprise.execute(widgetId, { action: action });
                      }
                    };
                  
                    execute();
                  }
                onload();
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
