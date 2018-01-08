import QtQuick 2.0
import Ubuntu.Components 0.1
import "../components"

TelegramPage {
    property alias passwordTextField: passwordTextField
    property alias passwordRepeatTextField: passwordRepeatTextField
    property alias oldPasswordTextField: oldPasswordTextField

    property alias errorLabel: errorLabel

    id: page
    head.backAction.visible: true
    objectName: "setPasswordPage"
    pageTitle: i18n.tr("Set Master password")
    onlineIndicationOnly: true

    body: Item {
        anchors {
            fill: parent
            margins: units.gu(2)
        }

        TelegramLabel {
            id: infoLabel
            anchors {
                top: parent.top
                margins: units.gu(1)
            }
            width: parent.width
            text: i18n.tr("Please enter the password.\n")
        }

        TelegramLabel {
            id: errorLabel
            anchors {
                top: infoLabel.bottom
                topMargin: units.gu(1)
            }
            width: parent.width
            visible: false
            color: "red"
        }
        TextField {
            id: oldPasswordTextField
            anchors {
                top: errorLabel.bottom
                topMargin: units.gu(1)
                left: parent.left
                right: parent.right
            }
            visible: settingsModel.encryptDatabase
            placeholderText: i18n.tr("Old password")
            echoMode: TextInput.Password
            // Keys.onEnterPressed: done()
            // Keys.onReturnPressed: done()

            horizontalAlignment: TextInput.AlignHCenter

            Component.onCompleted: {
              forceActiveFocus();
	          }

        }

        TextField {
            id: passwordTextField
            anchors {
                top: oldPasswordTextField.bottom
                topMargin: units.gu(1)
                left: parent.left
                right: parent.right
            }

            placeholderText: i18n.tr("Master password")
            echoMode: TextInput.Password
            // Keys.onEnterPressed: done()
            // Keys.onReturnPressed: done()

            horizontalAlignment: TextInput.AlignHCenter

            Component.onCompleted: {
              forceActiveFocus();
	          }

        }
        TextField {
            id: passwordRepeatTextField
            anchors {
                top: passwordTextField.bottom
                topMargin: units.gu(1)
                left: parent.left
                right: parent.right
            }

            placeholderText: i18n.tr("Repeat password")
            echoMode: TextInput.Password
            Keys.onEnterPressed: done()
            Keys.onReturnPressed: done()

            horizontalAlignment: TextInput.AlignHCenter

            Component.onCompleted: {

            }

        }

        TelegramButton {
            id: doneButton
            anchors {
                top: passwordRepeatTextField.bottom
                topMargin: units.gu(3)
                right: parent.right
                left: parent.left
            }
            width: parent.width

            enabled: true
            text: i18n.tr("Enter")
            onClicked: done()
        }
        TelegramButton {
            id: rmButton
            anchors {
                top: doneButton.bottom
                topMargin: units.gu(3)
                right: parent.right
                left: parent.left
            }
            width: parent.width
            visible: settingsModel.encryptDatabase

            enabled: true
            text: i18n.tr("Remove Password")
            onClicked: rmDone()
        }
    }

    signal error(int id, int errorCode, int errorText);

    function done() {
        if (busy) return;

        Qt.inputMethod.commit();
        Qt.inputMethod.hide();

        if (passwordTextField.text.length > 0) {
          if (passwordTextField.text==passwordRepeatTextField.text){
            busy = true;
            clearError();
            if(settingsModel.encryptDatabase){
              if(!storeModel.decryptDb(oldPasswordTextField.text)){
                setError(i18n.tr("Old Password is wrong"))
                busy = false;
              }
            }
            storeModel.encryptDb(passwordTextField.text)
            pageStack.push(dialogsPage)
          }
          else{
             setError(i18n.tr("Passwords not the same"))
             busy = false;
          }
        }
    }
    function rmDone() {
      if (passwordTextField.text.length>0) {
        clearError();
        if(storeModel.decryptDb(oldPasswordTextField.text)){
          setError(i18n.tr("Old Password is wrong"))
          busy = false;
        }
        else {
          pageStack.push(dialogsPage)
        }
      }
      else {
        setError(i18n.tr("Old Password is wrong"))
        busy = false;
      }

    }

    function onError(errorMessage) {
        passwordTextField.text = "";
        busy = false;
        setError(errorMessage);
    }

    function setError(message) {
        errorLabel.text = message;
        errorLabel.visible = true;
    }

    function clearError() {
        if (errorLabel.visible) {
            errorLabel.visible = false;
            errorLabel.text = "";
        }
    }
}
