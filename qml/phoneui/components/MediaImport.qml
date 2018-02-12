/*
 * Copyright (C) 2012-2014 Canonical, Ltd.
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

import QtQuick 2.2
import Ubuntu.Components 1.1
import Ubuntu.Components.Popups 1.0 as Popups
import Ubuntu.Content 1.1 as ContentHub

Item {
    id: root

    property var importDialog: null
    property var contentType: ContentHub.ContentType.All

    signal mediaReceived(string mediaUrl)

    function requestMedia() {
        if (!root.importDialog) {
            root.importDialog = PopupUtils.open(contentHubDialog, root)
        } else {
            console.warn("Import dialog already running")
        }
    }

    Component {
        id: contentHubDialog

        Popups.PopupBase {
            id: dialogue

            property alias activeTransfer: signalConnections.target
            focus: true

            Rectangle {
                anchors.fill: parent

                ContentHub.ContentPeerPicker {
                    id: peerPicker

                    anchors.fill: parent
                    contentType: root.contentType
                    handler: ContentHub.ContentHandler.Source

                    onPeerSelected: {
                        peer.selectionType = ContentHub.ContentTransfer.Single
                        dialogue.activeTransfer = peer.request()
                    }

                    onCancelPressed: {
                        PopupUtils.close(root.importDialog)
                    }
                }
            }

            Connections {
                id: signalConnections

                onStateChanged: {
                    var done = (dialogue.activeTransfer.state === ContentHub.ContentTransfer.Charged)
                             //   (dialogue.activeTransfer.state === ContentHub.ContentTransfer.Aborted))

                    if (dialogue.activeTransfer.state === ContentHub.ContentTransfer.Charged) {
                        dialogue.hide()
                        if (dialogue.activeTransfer.items.length > 0) {
                            root.mediaReceived(dialogue.activeTransfer.items[0].url)
                        }
                    }

                    if (done) {
                        acceptTimer.restart()
                    }
                }
            }

            // WORKAROUND: Work around for application becoming insensitive to touch events
            // if the dialog is dismissed while the application is inactive.
            // Just listening for changes to Qt.application.active doesn't appear
            // to be enough to resolve this, so it seems that something else needs
            // to be happening first. As such there's a potential for a race
            // condition here, although as yet no problem has been encountered.
            Timer {
                id: acceptTimer

                interval: 100
                repeat: true
                running: false
                onTriggered: {
                   if(Qt.application.active) {
                       PopupUtils.close(root.importDialog)
                   }
                }
            }

            Component.onDestruction: root.importDialog = null
        }
    }
}
