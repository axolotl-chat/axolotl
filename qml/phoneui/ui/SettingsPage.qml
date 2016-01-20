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
import Ubuntu.Components 1.1
import "../components"

TelegramPage {
    title: i18n.tr("Settings")
    property bool bug1341671workaround: true

    Column {
        id: settingsColumn

        anchors.fill: parent

        StandardListItem {
            title: i18n.tr("Advanced")
            onClicked: pageStack.push(Qt.resolvedUrl("settings/AdvancedPage.qml"))
        }
    }
}
