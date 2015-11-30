import QtQuick 2.0
import Ubuntu.Components 0.1
import Ubuntu.Telephony.PhoneNumber 0.1
import "../js/country_data.js" as CountryData
import "../components"

TelegramPage {
    property alias errorLabel: errorLabel

    objectName: "signInPage"
    head.backAction.visible: false

    onlineIndicationOnly: true

    body: Item {
        anchors {
            fill: parent
            margins: units.gu(2)
        }

        Text {
            id: country
            text: i18n.tr("YOUR COUNTRY")
            anchors {
                top: parent.top
                topMargin: units.gu(2)
            }
        }

        OptionSelector {
            id: countrySelector
            anchors {
                top: country.bottom
                topMargin: units.gu(2)
            }
            containerHeight: itemHeight * 4

            onDelegateClicked: {
                var country = model[index]
                var tel = CountryData.name_to_tel[country]
                countryTextField.text = tel
                userTextField.defaultRegion = CountryData.tel_to_iso[tel]
                userTextField.focus = true
            }

            Component.onCompleted: {
                var countries = []
                for (var c in CountryData.name_to_tel) {
                    countries.push(c)
                }
                countrySelector.model = countries
                // lolz
                //countrySelector.selectedIndex = countries.indexOf('United States')

            }
        }

        Text {
            id: countryCode
            text: i18n.tr("YOUR COUNTRY CODE AND PHONE NUMBER")
            anchors {
                top: countrySelector.bottom
                topMargin: units.gu(2)
            }
        }

        Row {
            id: userEntryRow
            anchors {
                top: countryCode.bottom
                topMargin: units.gu(2)
            }
            height: countrySelector.itemHeight
            width: parent.width
            spacing: units.gu(1)

            Label {
                id: label
                text: "+"
                width: units.gu(2)
                height: parent.height
                verticalAlignment: TextInput.AlignVCenter
                horizontalAlignment: TextInput.AlignHCenter
            }
            TextField {
                id: countryTextField
                horizontalAlignment: TextInput.AlignHCenter
                width: units.gu(8)
                height: parent.height

                inputMethodHints: Qt.ImhDialableCharactersOnly
                validator: RegExpValidator {
                    regExp: /[0-9]*/
                }
                placeholderText: {
                    CountryData.iso_to_tel[userTextField.defaultRegion]
                }

                KeyNavigation.tab: userTextField
                onDisplayTextChanged: {
                    var tel = countryTextField.text
                    var country = CountryData.tel_to_name[tel]
                    if (country !== "") {
                        countrySelector.selectedIndex = countrySelector.model.indexOf(country);
                    }
                    if (tel !== "") {
                        var iso = CountryData.tel_to_iso[tel];
                        if (typeof iso != "undefined") {
                            userTextField.defaultRegion = iso;
                        }
                    }
                }
            }
            // XXX: Requires private API - http://pad.lv/1346450
            PhoneNumberField {
                id: userTextField
                horizontalAlignment: TextInput.AlignHCenter
                width: countrySelector.width - countryTextField.width - label.width - units.gu(2)
                height: parent.height

                updateOnlyWhenFocused: false
                defaultRegion: "US"
                autoFormat: userTextField.text.length > 0 && userTextField.text.charAt(0) !== "*" && userTextField.text.charAt(0) !== "#"
                inputMethodHints: Qt.ImhDialableCharactersOnly

                onDisplayTextChanged: clearError()
                Keys.onEnterPressed: done()
                Keys.onReturnPressed: done()
            }
        }

        Text {
            id: infoLabel
            anchors {
                top: userEntryRow.bottom
                margins: units.gu(1)
                topMargin: units.gu(4)
            }
            wrapMode: Text.WordWrap
            horizontalAlignment: Text.AlignLeft
            width: parent.width
            text: i18n.tr("Verify your phone number to connect with Signal.")+"\n\n"+
                  i18n.tr("Registration transmits some contact information to the server. It is not stored.")
        }

        TelegramButton {
            id: doneButton
            anchors {
                top: infoLabel.bottom
                topMargin: units.gu(3)
                left: parent.left
                right: parent.right
            }
            enabled: isConnected
                     && userTextField.text !== ""
                     && countryTextField.text !== ""
            text: i18n.tr("Register")
            onClicked: done()
        }

        TelegramLabel {
            id: errorLabel
            anchors {
                top: infoLabel.bottom
                margins: units.gu(1)
            }
            width: parent.width
            visible: false
            color: "red"
        }
    }

    signal numberEntered(string text)

    function done() {
        if (busy) return;

        Qt.inputMethod.commit();
        Qt.inputMethod.hide();

        busy = true
        numberEntered(getPhoneNumber());
    }

    function getPhoneNumber() {
	    var n = "+" + countryTextField.text + userTextField.text;
	    return n.replace(/[\s\-\(\)]/g, '')
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
