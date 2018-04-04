import QtQuick 2.4
import Ubuntu.Components 1.3
import Ubuntu.Components.Popups 1.0

Dialog {
    id: dialogue

    property alias textWithLink:textBox.text

    Text {
        id: textBox
        wrapMode: Text.WordWrap
        onLinkActivated:Qt.openUrlExternally(link)
    }

    Button {
        text: i18n.tr("OK")
        color: UbuntuColors.green
        onClicked: PopupUtils.close(dialogue);
    }
}
