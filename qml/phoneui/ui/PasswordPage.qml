import QtQuick 2.0
import Ubuntu.Components 0.1
import "../components"

TelegramPage {
    property alias passwordTextField: passwordTextField
    property alias errorLabel: errorLabel

    id: page
    head.backAction.visible: false
    objectName: "passwordPage"
    pageTitle: i18n.tr("Master password")
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
            id: passwordTextField
            anchors {
                top: errorLabel.bottom
                topMargin: units.gu(1)
                left: parent.left
                right: parent.right
            }

            placeholderText: i18n.tr("Master password")
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
            text: i18n.tr("Enter")
            onClicked: done()
        }
    }

    signal error(int id, int errorCode, int errorText);

    signal passwordEntered(string text)

    function done() {
        if (busy) return;

        Qt.inputMethod.commit();
        Qt.inputMethod.hide();

        if (passwordTextField.text.length > 0) {
            busy = true;
            clearError();
	    passwordEntered(passwordTextField.text);
	    pageStack.push(dialogsPage)
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
