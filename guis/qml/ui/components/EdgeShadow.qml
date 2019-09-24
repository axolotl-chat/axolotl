/*
 * Copyright 2015 Canonical Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import QtGraphicalEffects 1.0

DropShadow {
    property int depth: 1
    property bool topShadow: false

    property real __hOffset: 2 + 2 * depth
    property real __vOffset: 2 + 2 * depth

    anchors.fill: source
    horizontalOffset: topShadow ? 0 : __hOffset
    verticalOffset: topShadow ? - __vOffset : __vOffset
    color: Qt.rgba(0.6, 0.6, 0.6, 0.6)
    radius: 8
    samples: 16 // radius * 2
    transparentBorder: true
    visible: source.visible
}
