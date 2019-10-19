import QtQuick.Controls 2.2
import QtQuick 2.2
import Ubuntu.Components 1.3
import QtWebEngine 1.7
import Ubuntu.Components.Popups 1.3
Dialog {
  id: desktopLinkDialog
  property QtObject request
  title: request.message
  text: "Add signal desktop"
  UITK.TextArea {
      height: units.gu(10)
      autoSize: true
      anchors {
          left: parent.left
          right: parent.right
          topMargin: units.gu(0.1)
      }
      id: optionalMessage
      placeholderText: i18n.tr("Enter optional message...")
  }

  Button {
    text: "cancel"
    onClicked:{
      console.log(requestTest)
      requestTest.dialogReject()
      PopupUtils.close(desktopLinkDialog)
    }
  }
  Button {
    text: "Add"
    color: UbuntuColors.green
    onClicked:{
      requestTest.dialogAccept("dialogAccepttext")
      PopupUtils.close(desktopLinkDialog)
    }
  }
}
