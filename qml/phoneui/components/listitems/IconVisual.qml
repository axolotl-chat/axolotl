/*
 * Copyright 2012 Canonical Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import QtQuick 2.4
import Ubuntu.Components 0.1

// internal helper class to create the visuals
// for the icon.
Item {
    id: iconVisual

    /*!
      \qmlproperty url source
     */
    property alias source: icon.source
    /*!
      \qmlproperty url fallbackSource
     */
    property alias fallbackSource: icon.fallbackSource
    visible: source != ""
    property bool hasFrame: true

    ImageWithFallback {
        id: icon
        visible: !iconVisual.hasFrame
        opacity: iconVisual.enabled ? 1.0 : 0.5
        fillMode: Image.PreserveAspectCrop
        anchors.fill: parent
        smooth: true
        asynchronous: true
    }

    UbuntuShape {
        id: shape
        visible: iconVisual.hasFrame
        anchors.fill: parent
        image: icon
    }
}
