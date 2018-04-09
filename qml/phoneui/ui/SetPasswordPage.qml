import QtQuick 2.4
import Ubuntu.Components 1.3
import "../components"

TelegramPage {
    property alias passwordTextField: passwordTextField
    property alias passwordRepeatTextField: passwordRepeatTextField
    property alias oldPasswordTextField: oldPasswordTextField

    property alias errorLabel: errorLabel

    id: page
    //head.backAction.visible: true
    objectName: "setPasswordPage"
    onlineIndicationOnly: true
    header:PageHeader{
      title: settingsModel.encryptDatabase ? i18n.tr("Change passphrase") : i18n.tr("Create passphrase")
      leadingActionBar.actions:[
        Action {
          id: backAction
          iconName: "back"
          onTriggered:{
              pageStack.pop();
          }
        }
      ]
    }
    body: Item {
        anchors {
            fill: parent
            margins: units.gu(2)
        }

        TelegramLabel {
            id: errorLabel
            anchors {
                top: parent.top
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
            placeholderText: i18n.tr("Old passphrase")
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

            placeholderText: i18n.tr("New passphrase")
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

            placeholderText: i18n.tr("Repeat new passphrase")
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
                topMargin: units.gu(2)
                right: parent.right
                left: parent.left
            }
            width: parent.width

            enabled: true
            text: settingsModel.encryptDatabase ?  i18n.tr("Change passphrase") : i18n.tr("Submit passphrase")
            onClicked: done()
        }
        TelegramButton {
            id: rmButton
            anchors {
                top: doneButton.bottom
                topMargin: units.gu(1)
                right: parent.right
                left: parent.left
            }
            width: parent.width
            visible: settingsModel.encryptDatabase

            enabled: true
            text: i18n.tr("Disable passphrase?")
            onClicked: rmDone()
        }
    }

    signal error(int id, int errorCode, int errorText);

    function done() {
        if (busy) return;

        Qt.inputMethod.commit();
        Qt.inputMethod.hide();

        if (passwordTextField.text.length >= 6) {

          if (passwordTextField.text==passwordRepeatTextField.text){
            busy = true;
            clearError();
            if(settingsModel.encryptDatabase){
              if(!storeModel.decryptDb(oldPasswordTextField.text)){
                setError(i18n.tr("Incorrect old passphrase!"))
                busy = false;
              }
            }
            storeModel.encryptDb(passwordTextField.text)
            pageStack.push(dialogsPage)
          }
          else{
             setError(i18n.tr("Passphrases don\'t match!"))
             busy = false;
          }
        }
        else{
          setError(i18n.tr("New passphrase to short(6)!"))
          busy = false;
        }
    }
    function rmDone() {
      if (passwordTextField.text.length>0) {
        clearError();
        if(storeModel.decryptDb(oldPasswordTextField.text)){
          setError(i18n.tr("Incorrect old passphrase!"))
          busy = false;
        }
        else {
          pageStack.push(dialogsPage)
        }
      }
      else {
        setError(i18n.tr("Incorrect old passphrase!"))
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
