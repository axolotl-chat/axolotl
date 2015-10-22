import QtQuick 2.3
import Ubuntu.Components 1.1
import Ubuntu.Components.Popups 1.0

Dialog {
    id: dialogue
    title: "Verify identity"

    Button {
        text: i18n.tr("OK")
        color: UbuntuColors.green
        onClicked: PopupUtils.close(dialogue);
    }
}
