import QtQuick 2.4
import Ubuntu.Components 0.1
import Ubuntu.Components.Popups 0.1

import "../components"

TelegramPage {
    property alias passwordTextField: passwordTextField
    property alias errorLabel: errorLabel

    id: page
    //head.backAction.visible: false
    objectName: "passwordPage"
    pageTitle: i18n.tr("Enter passphrase")
    onlineIndicationOnly: true

    body: Item {
        anchors {
            fill: parent
            margins: units.gu(2)
        }
        Image {
              id: logoImage
              width: 200; height: 200
              fillMode: Image.PreserveAspectFit
              anchors {
                 horizontalCenter: parent.horizontalCenter;
                  top: parent.top
                  topMargin: units.gu(1)
              }
                source: "../images/logo.png"
          }



        TelegramLabel {
            id: errorLabel
            anchors {
                top: logoImage.bottom
                topMargin: units.gu(1)
            }
            width: parent.width
            visible: false
            color: "red"
        }

        TextField {
            id: passwordTextField
            anchors {
                top: errorLabel.bottom
                topMargin: units.gu(1)
                left: parent.left
                right: parent.right
            }

            placeholderText: i18n.tr("Passphrase")
            echoMode: TextInput.Password
            Keys.onEnterPressed: done()
            Keys.onReturnPressed: done()

            horizontalAlignment: TextInput.AlignHCenter

            Component.onCompleted: {
			    forceActiveFocus();
	    }

        }

        TelegramButton {
            id: doneButton
            anchors {
                top: passwordTextField.bottom
                topMargin: units.gu(1)
                right: parent.right
                left: parent.left
            }
            width: parent.width

            enabled: isConnected && passwordTextField.text !== ""
            text: i18n.tr("Submit passphrase")
            onClicked: done()
        }
        TelegramButton {
            id: resetButton
            anchors {
                top: doneButton.bottom
                topMargin: units.gu(1)
                right: parent.right
                left: parent.left
            }
            width: parent.width

            enabled: true
            text: i18n.tr("Disable passphrase?")
            onClicked: {
              PopupUtils.open(Qt.resolvedUrl("dialogs/ConfirmationDialog.qml"),
                          resetButton, {
                              title: i18n.tr("Disable passphrase?"),
                              text: i18n.tr("This will permanently delete all messages!."),
                              onAccept: function() {
                                  storeModel.resetDb()
                                  pageStack.push(dialogsPage);
                              }
                          })
            }
        }
    }

    signal error(int id, int errorCode, int errorText);

    // signal passwordEntered(string text)

    function done() {
        if (busy) return;

        Qt.inputMethod.commit();
        Qt.inputMethod.hide();

        if (passwordTextField.text.length > 0) {
            busy = true;
            clearError();
            if (storeModel.setupDb(passwordTextField.text))pageStack.push(dialogsPage);
            else  {
              setError(i18n.tr("Invalid passphrase!"))
              busy = false;

            }
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
