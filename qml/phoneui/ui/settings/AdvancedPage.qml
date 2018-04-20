/*
 * Copyright (C) 2015 Canonical Ltd
 *
 * This file is part of Ubuntu Weather App
 *
 * Ubuntu Weather App is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * Ubuntu Weather App is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import QtQuick 2.4
import Ubuntu.Components 1.3
import Ubuntu.Components.ListItems 1.0 as ListItem
import Ubuntu.Components.Popups 1.0
import "../../components"


TelegramPage {
    id: advancedSettings
    header: PageHeader {
      title: i18n.tr("Advanced")
          id: pageHeader
          leadingActionBar.actions: [
            Action {
              iconName: "back"
              text: "Back"
              onTriggered:{
                back();
              }
            }
          ]
    }

    Column {
      anchors {
        top: pageHeader.bottom
        left: parent.left
        right: parent.right
        bottom: parent.bottom
      }

        ListItem.Subtitled {
            text: i18n.tr("Unregistering")
            subText: textsecure.phoneNumber
            onClicked: {
                PopupUtils.open(Qt.resolvedUrl("../dialogs/ConfirmationDialog.qml"),
                root, {
                    title: i18n.tr("Disable Signal messages and calls?"),
                    text: i18n.tr("Disable Signal messages and calls by unregistering from the server. You will need to re-register your phone number to use them again in the future."),
                    onAccept: function() {
                        textsecure.unregister()
                    }
                })
            }
        }
        ListItem.Standard {
            control: CheckBox {
                checked: settingsModel.sendByEnter

                onCheckedChanged: {
                    settingsModel.sendByEnter = checked
                    textsecure.saveSettings()
                }

            }
            text: i18n.tr("Enter key sends")

            onClicked: control.checked = !control.checked
        }
        ListItem.Subtitled {
            text: i18n.tr("Submit debug log")
            subText: "Signal "+appVersion
            onClicked: {
                var ret = textsecure.submitDebugLog()

                var url = ret[0]

                if (url != "") {
                    PopupUtils.open(Qt.resolvedUrl("../dialogs/InfoDialog.qml"),
                    root, {
                        title: i18n.tr("Success!"),
                        textWithLink: i18n.tr("Copy this URL and add it to your issue report or support email:")+"<br> <a href='"+ url + "'>"+url+"</a>"
                    })
                }
            }
        }
    }
}
