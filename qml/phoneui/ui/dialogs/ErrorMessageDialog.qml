import QtQuick 2.3
import Ubuntu.Components 1.1
import Ubuntu.Components.Popups 1.0

Dialog {
    id: dialogue
    title: "Signal error"

    text: "Unknown error"

    Button {
        text: i18n.tr("OK")
        color: UbuntuColors.red
        onClicked: PopupUtils.close(dialogue);
    }
}
