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

import QtQuick 2.4
import Ubuntu.Components 1.3
import Ubuntu.Content 1.1

import "../components"
import "../components/TelegramColors.js" as TelegramColors

TelegramPage {
    property int chatId: 0
    property string senderName: ""
    property string photoPreviewSource: ""
    property string audioPreviewSource: ""
    property string videoPreviewSource: ""

    id: previewPage
    title: i18n.tr("From: ") + senderName

    //head.actions: [
    // 
    //     Action {
    //         iconName: "save"
    //         text: i18n.tr("Save")
    //         onTriggered: save()
    //         visible: saveAndShareVisible()
    //     }
    // ]

    function saveAndShareVisible() {
        return (photoPreviewSource !== "" || videoPreviewSource !== "");
    }

    function save() {
        singleMediaViewer.reset();

        if (photoPreviewSource !== "") {
            pageStack.push(picker, {
                "url": photoPreviewSource,
                "handler": ContentHandler.Destination,
                "contentType": ContentType.Pictures
            });
        } else if (videoPreviewSource !== "") {
            pageStack.push(picker, {
                "url": videoPreviewSource,
                "handler": ContentHandler.Destination,
                "contentType": ContentType.Videos
            });
        }
    }

    body: Item {

        anchors {
            fill: parent
        }

        Rectangle {
            anchors.fill: parent
            color: TelegramColors.page_background
        }

        SingleMediaViewer {
            id: singleMediaViewer
            anchors.fill: parent
            maxDimension: 2*Math.max(previewPage.width, previewPage.height)
        }
    }
}
