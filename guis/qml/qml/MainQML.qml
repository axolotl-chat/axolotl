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
				
				window.onload=function() {
					var action = 'registration';
					var isDone = false;
					var sitekey = '6LfBXs0bAAAAAAjkDyyI1Lk5gBAUWfhI_bIyox5W';
			
					var widgetId = grecaptcha.enterprise.render('container', {
					sitekey: sitekey,
					size: 'checkbox',
					callback: function (token) {
						isDone = true;
						document.body.removeAttribute('class');
						window.location = ['http://localhost:9080/?token=', sitekey, action, token].join(".");
					},
					});
				}
				// cleanup
				var bodyTag = document.getElementsByTagName('body')[0];	
				bodyTag.innerHTML ='<div id="container"></div>'
				grecaptcha  = undefined

				// reload recaptcha
				var script = document.createElement('script');
				script.type = 'text/javascript';
				script.src = 'https://www.google.com/recaptcha/enterprise.js?onload=onload&render=explicit';
				bodyTag.appendChild(script);
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
