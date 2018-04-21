import QtQuick 2.4
import Ubuntu.Components 1.3
import "../components"

TelegramPage {
    property alias codeTextField: codeTextField
    property alias errorLabel: errorLabel
    property alias countdownText: countdownLabel.text
    property alias countdownTimer: countdownTimer

    id: page
    //head.backAction.visible: false
    objectName: "codeVerificationPage"
    onlineIndicationOnly: true
    header:PageHeader{
          id:pageHeader
          title: i18n.tr("Verifying number")
    }

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
            text: i18n.tr("Signal will now automatically verify your number with a confirmation SMS message.")
        }

        TelegramLabel {
            id: countdownLabel
            anchors {
                top: infoLabel.bottom
                topMargin: units.gu(1)
            }
            width: parent.width
            // TRANSLATORS: the argument refers to a countdown time
            text: i18n.tr("Waiting for SMS verification...") + " " + countdownTimer.getTimeAsText()
        }

        TelegramLabel {
            id: errorLabel
            anchors {
                top: countdownLabel.bottom
                topMargin: units.gu(1)
            }
            width: parent.width
            visible: false
            color: "red"
        }

        TextField {
            id: codeTextField
            anchors {
                top: errorLabel.bottom
                topMargin: units.gu(1)
                left: parent.left
                right: parent.right
            }
            inputMethodHints: Qt.ImhDigitsOnly

            validator: RegExpValidator {
                regExp: /[\w]+/
            }
            placeholderText: i18n.tr("Code")

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
                top: codeTextField.bottom
                topMargin: units.gu(1)
                right: parent.right
                left: parent.left
            }
            width: parent.width

            enabled: isConnected && codeTextField.text !== ""
            text: i18n.tr("OK")
            onClicked: done()
        }
    }

    Timer {
        id: countdownTimer
        readonly property int timeToCall: 120
        property int seconds: timeToCall
        property int timeStarted: 0
        interval: 1000
        repeat: true
        running: false

        onTriggered: {
            seconds = Math.max(0, timeToCall - (Date.now() / 1000 - timeStarted));
            countdownLabel.text = i18n.tr("Waiting for SMS verification...")+ " "+getTimeAsText();
            if (seconds <= 0) {
                stop();
            }
        }

        Component.onCompleted: startTimer()

        function getTimeAsText() {
            var min = Math.floor(seconds / 60);
            var sec = (seconds % 60).toString()

            var pad = "00";
            sec = pad.substring(0, pad.length - sec.length) + sec;
            return min + ':' + sec;
        }
    }

    signal error(int id, int errorCode, int errorText);
    signal authSignInError();
    signal authLoggedIn();
    signal calling();

    onCalling: {
        countdownLabel.text = i18n.tr("Calling you (this may take a while)...");
    }

    onError: {
        if (errorCode === 420) {
            setError(i18n.tr("Please wait a moment and try again"));
        } else if (errorCode === 400) {
            // handled in onAuthSignInError
        } else {
            console.log("VerificationCode error: " + errorCode + " " + errorText);
        }
        busy = false;
    }

    onAuthSignInError: {
        setError(i18n.tr("Incorrect code. Please try again."));
        busy = false;
    }

    Component.onCompleted: {
    }

    Component.onDestruction: {
    }

    function startTimer() {
        countdownTimer.seconds = countdownTimer.timeToCall;
        countdownTimer.timeStarted = Date.now() / 1000;
        countdownTimer.start();
    }

    function stopTimer() {
        countdownTimer.stop();
    }

    signal codeEntered(string text)

    function done() {
        if (busy) return;

        Qt.inputMethod.commit();
        Qt.inputMethod.hide();

        if (codeTextField.text.length > 0) {
            busy = true;
            clearError();
	    countdownTimer.running = false;
	    codeEntered(codeTextField.text);
        }
    }

    function onError(errorMessage) {
        countdownTimer.running = true;
        codeTextField.text = "";
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
