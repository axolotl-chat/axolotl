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

Item {
    id: attach_panel

    property bool isShown: false

    signal photoRequested()
    signal videoRequested()
    signal audioRequested()
    signal contactRequested()
    signal close()

    anchors {
        right: parent.right
        left: parent.left
        bottom: parent.bottom
    }
    clip: false

    onClose: attach_panel.isShown = false

    Rectangle {
        id: attach_box
        width: parent.width
        height: units.gu(24)+units.gu(3)
        y: attach_panel.isShown ? -height+units.gu(3) : shadow.height

        Behavior on y {
            SequentialAnimation {
                NumberAnimation { easing.type: Easing.OutBack; duration: 300; }
                ScriptAction {
                    script: {
                        if (!attach_panel.isShown) {
                            privates.attachmentItem.destroy()
                            privates.attachmentItem = undefined
                        }
                    }
                }
            }
        }

        Rectangle {
            anchors.fill: parent
            color: "#aaffffff"
        }

        InverseMouseArea {
            anchors.fill: parent
            acceptedButtons: Qt.LeftButton
            onPressed: attach_panel.isShown = false
        }

        Grid {
            id: attachment_grid
            anchors {
                top: parent.top
                topMargin: units.gu(1)
                horizontalCenter: parent.horizontalCenter
            }
            columns: 4
            rows: 2
            horizontalItemAlignment: Grid.AlignHCenter
            verticalItemAlignment: Grid.AlignVCenter
            spacing: units.gu(2.8)

            AttachPanelItem {
                id: attach_photo_item
                // TRANSLATORS: Used in attach menu, when sending a photo to the conversation.
                text: i18n.tr("Image")
                image: Qt.resolvedUrl("../images/files/android/attach_gallery.png")
                onClicked: {
                    Haptics.play()
                    attach_panel.photoRequested()
                    attach_panel.close()
                }
            }

            AttachPanelItem {
                // TRANSLATORS: Used in attach menu, when sending a video to the conversation.
                text: i18n.tr("Video")
                image: Qt.resolvedUrl("../images/files/android/attach_video.png")
                onClicked: {
                    Haptics.play()
                    attach_panel.videoRequested()
                    attach_panel.close()
                }
            }

            AttachPanelItem {
                // TRANSLATORS: Used in attach menu, when sending a file to the conversation.
                text: i18n.tr("Audio")
                image: Qt.resolvedUrl("../images/files/android/attach_audio.png")
                onClicked: {
                    Haptics.play()
                    attach_panel.audioRequested()
                    attach_panel.close()
                }
            }

            AttachPanelItem {
                // TRANSLATORS: Used in attach menu, when sending a file to the conversation.
                text: i18n.tr("Contact")
                image: Qt.resolvedUrl("../images/files/android/attach_contact.png")
                onClicked: {
                    Haptics.play()
                    attach_panel.contactRequested()
                    attach_panel.close()
                }
            }

            AttachPanelItem {
                height: attach_photo_item.height
                // TRANSLATORS: Used in attach menu, when sending a file to the conversation.
                image: Qt.resolvedUrl("../images/files/android/attach_hide1.png")
                showTick: true
                onClicked: {
                    Haptics.play()
                    attach_panel.close()
                }
            }

        }
    }

    EdgeShadow {
        id: shadow
        source: attach_box
        topShadow: true
    }
}
