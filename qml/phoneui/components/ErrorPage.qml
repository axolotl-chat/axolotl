import QtQuick 2.2
import Ubuntu.Components 1.1
import Ubuntu.Components.ListItems 1.0 as ListItem

Page {
	id: errorPage
	title: i18n.tr("Error")
	property string message: "default"
	Rectangle {
		color:"red"
		anchors.fill:parent
		Text {
			anchors.fill:parent
			text: errorPage.message
			wrapMode: Text.WordWrap
		}
	}
}

