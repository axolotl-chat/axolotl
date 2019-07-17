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

import QtQuick 2.4
import Ubuntu.Components 1.3
import Ubuntu.Components.Popups 1.0 as Popups
import Ubuntu.Content 1.1 as ContentHub
// import Ubuntu.Contacts 0.1
import QtContacts 5.0

Item {
	id: root

	property var importDialog: null

	function contactImportDialog() {
		PopupUtils.open(dialog, root)
	}

	function requestContact()
	{
		if (!root.importDialog) {
			root.importDialog = PopupUtils.open(contentHubDialog, root)
		} else {
			console.warn("Import dialog already running")
		}
	}

	function importContacts(url) {
		textsecure.contactsImported(url)
	}

	Component {
		id: dialog
		Popups.Dialog {
			id: dlg
			title: qsTr("Import contacts")

			Button {
				text: qsTr("From Address Book")
				color: UbuntuColors.orange
				onClicked: {
					PopupUtils.close(dlg)
					root.requestContact();
				}
			}
			Button {
				text: qsTr("Cancel")
				onClicked: PopupUtils.close(dlg)
			}
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
					contentType: ContentHub.ContentType.Contacts
					handler: ContentHub.ContentHandler.Source

					onPeerSelected: {
						peer.selectionType = ContentHub.ContentTransfer.Multiple
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
					if (dialogue.activeTransfer.state === ContentHub.ContentTransfer.Charged) {
						dialogue.hide()
						if (dialogue.activeTransfer.items.length > 0) {
							importContacts(dialogue.activeTransfer.items[0].url);
						}
					}
				}
			}
		}
	}
}
