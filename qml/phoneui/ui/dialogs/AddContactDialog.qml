import QtQuick 2.9
import QtQuick.Controls 2.2
import Ubuntu.Components 1.3
import Ubuntu.Components.Popups 1.3
import Ubuntu.Telephony.PhoneNumber 0.1

Dialog {
    id: addContactDialog

    property string acceptText: i18n.tr("OK")
    property string altText: ""
    property string cancelText: i18n.tr("Cancel")
    property var onAccept: function() {
        addContact()
    }
    property var onAlt: function() {}
    property var onCancel: function() {
        PopupUtils.close(addContactDialog);
    }

    TextField {
        id: inputName
        text: qsTr("")
        placeholderText: "Name"
        horizontalAlignment: Text.AlignHCenter
        font.pixelSize: 12
        anchors.top: parent.top
        anchors.topMargin: 40

    }

    Row {
        id: rowInputPhone
        height: 30
        anchors.top: inputName.bottom
        anchors.topMargin: 10

        Text {
            id: plusLabel
            text: qsTr("+")
            anchors.verticalCenter: parent.verticalCenter
            anchors.left: parent.left
            anchors.leftMargin: 0
            font.pixelSize: 12
        }

        TextField {
            id: inputCountry
            width: 60
            text: qsTr("")
            placeholderText: "44"
            anchors.top: parent.top
            anchors.topMargin: 0
            anchors.left: plusLabel.right
            font.pixelSize: 12
            anchors.leftMargin: 5
            horizontalAlignment: Text.AlignHCenter
        }

        TextField {
            id: inputPhone
            text: qsTr("")
            anchors.right: parent.right
            inputMethodHints: Qt.ImhDigitsOnly
            anchors.rightMargin: 0
            placeholderText: "123"
            anchors.top: parent.top
            anchors.topMargin: 0
            anchors.left: inputCountry.right
            anchors.leftMargin: 10
            horizontalAlignment: Text.AlignHCenter
            font.pixelSize: 12
        }
    }
    Button {
        id: acceptButton
        text: acceptText
        anchors.top: rowInputPhone.bottom
        color: UbuntuColors.green
        anchors.topMargin: 10
        onClicked: optionSelected(onAccept)
    }

    Button {
        text: cancelText
        anchors.top: acceptButton.bottom
        color: UbuntuColors.lightGrey
        anchors.topMargin: 10
        onClicked: optionSelected(onCancel)
    }
    function addContact() {
      textsecure.addContact(inputName.text, getPhoneNumber())
    }
    function getPhoneNumber() {
      var n = "+" + inputCountry.text + inputPhone.text;
      return n.replace(/[\s\-\(\)]/g, '')
    }
    function optionSelected(option) {
        option();
        PopupUtils.close(addContactDialog);
    }
}
