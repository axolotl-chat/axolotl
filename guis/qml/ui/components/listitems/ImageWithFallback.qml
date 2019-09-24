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

Image {
    id: image

    property url fallbackSource
    property bool fallbackRequired: false

    function isSourceDefined(sourceUrl) {
        return sourceUrl != "" && sourceUrl != undefined
    }

    function tryLoadingFallbackSource() {
        if (isSourceDefined(fallbackSource)) {
            source = fallbackSource
        }
    }

    function checkStatus() {
        if (!isSourceDefined(source) || (status == Image.Error && source != fallbackSource)) {
            fallbackRequired = true
            tryLoadingFallbackSource()
        }
    }

    onSourceChanged: fallbackRequired = false
    onFallbackSourceChanged: if (fallbackRequired) tryLoadingFallbackSource()
    onStatusChanged: checkStatus()
    Component.onCompleted: checkStatus()
}
