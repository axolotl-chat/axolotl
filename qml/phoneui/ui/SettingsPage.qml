import QtQuick 2.0
import Ubuntu.Components 1.1
import Ubuntu.Components.ListItems 0.1 as ListItem
import Ubuntu.Components.Popups 0.1
import Ubuntu.Content 1.0

import "../components"
import "../js/avatar.js" as Avatar
import "../js/time.js" as Time

TelegramPage {
    id: page
    pageTitle: i18n.tr("Settings");

    VisualItemModel {
        id: model

        ListItem.Header {
		text: "TextSecure for Ubuntu Phone, version "+appVersion
        }

        ListItem.Header {
		text: i18n.tr("Privacy")
        }

        ListItem.Header {
            text: i18n.tr("Advanced")
        }

        ListItem.Standard {
            text: i18n.tr("Enter key sends")
            showDivider: false

            Switch {
                id: checkbox
                anchors {
                    right: parent.right
                    rightMargin: units.gu(2)
                    verticalCenter: parent.verticalCenter
                }

                onCheckedChanged: settingsModel.sendByEnter = checked
                Component.onCompleted: checked = settingsModel.sendByEnter
            }
    	}
    }

    body: Item {
        anchors.fill: parent

        ListView {
            anchors {
                topMargin: units.gu(2)
                top: parent.top
                left: parent.left
                bottom: parent.bottom
                right: parent.right
            }
            clip: true
            model: model
        }
    }
}
