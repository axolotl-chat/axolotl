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
import "../components"

TelegramPage {
    header:PageHeader{
      title: i18n.tr("Settings")
      id: pageHeader
      leadingActionBar.actions:[
        Action {
          id: backAction
          iconName: "back"
          onTriggered:{
            back();
          }
        }
      ]
    }
    property bool bug1341671workaround: true
    function back() {
        pageStack.pop();
    }
    Column {
        id: settingsColumn
        anchors {
          top: pageHeader.bottom
          left: parent.left
          right: parent.right
          bottom: parent.bottom
        }
        ListItem.Standard {
          text: settingsModel.encryptDatabase ? i18n.tr("Change passphrase") : i18n.tr("Create passphrase")
          onClicked: pageStack.push(Qt.resolvedUrl("SetPasswordPage.qml"))
        }
        ListItem.Standard {
          text: i18n.tr("Advanced")
          onClicked: pageStack.push(Qt.resolvedUrl("settings/AdvancedPage.qml"))
        }
        ListItem.Standard {
          text: i18n.tr("Linked Devices")
          onClicked: pageStack.push(Qt.resolvedUrl("settings/LinkedDevicesPage.qml"))
        }
    }
}
