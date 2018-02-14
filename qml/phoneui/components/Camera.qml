import QtQuick 2.4
import QtQuick.Layouts 1.1
import Ubuntu.Components 1.3
import Ubuntu.Components.ListItems 1.3
import Ubuntu.Components.Popups 1.3
import QtMultimedia 5.0
import QtQuick.Window 2.0
import Ubuntu.Content 1.3
Item {
    id: root
    anchors.fill: parent
    // anchors.verticalCenter: parent.verticalCenter
    // anchors.horizontalCenter: parent.horizontalCenter
    // anchors.leftMargin: 5;
    // anchors.topMargin: 10;
    Camera {
        id: camera

        //        flash.mode: torchButton.active ? Camera.FlashTorch : Camera.FlashOff
        //        flash.mode: Camera.FlashTorch

        focus.focusMode: Camera.FocusContinuous
        focus.focusPointMode: Camera.FocusPointAuto

        /* Use only digital zoom for now as it's what phone cameras mostly use.
               TODO: if optical zoom is available, maximumZoom should be the combined
               range of optical and digital zoom and currentZoom should adjust the two
               transparently based on the value. */
        property alias currentZoom: camera.digitalZoom
        property alias maximumZoom: camera.maximumDigitalZoom

        function startAndConfigure() {
            // start();
            focus.focusMode = Camera.FocusContinuous
            focus.focusPointMode = Camera.FocusPointAuto
        }

        Component.onCompleted: {
            captureTimer.start()
        }
    }

    Timer {
        id: captureTimer
        interval: 2000
        repeat: true
        onTriggered: {
          print("capturing");
          textsecure.addDevice();
          camera.startAndConfigure();
        }
    }
    Image {
      id: qrimage
    }
    VideoOutput {
      anchors.fill: parent
        anchors.leftMargin: 15;
        anchors.topMargin: 10;
      anchors.verticalCenter: parent.verticalCenter
      anchors.horizontalCenter: parent.horizontalCenter

      fillMode: Image.PreserveAspectCrop

      orientation: {
          var angle = Screen.primaryOrientation == Qt.PortraitOrientation ? -90 : 0;
          angle += Screen.orientation == Qt.InvertedLandscapeOrientation ? 180 : 0;
          return angle;
      }
      source: camera
      focus: visible

    }
    // function qImage2jpeg(QImage image){
    //
    // }

}
