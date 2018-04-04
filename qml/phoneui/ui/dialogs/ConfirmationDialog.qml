import QtQuick 2.4
import Ubuntu.Components 1.3
import Ubuntu.Components.Popups 1.0

Dialog {
    id: dialogue

    property string acceptText: i18n.tr("OK")
    property string altText: "" 
    property string cancelText: i18n.tr("Cancel")
    property var onAccept: function() {}
    property var onAlt: function() {}
    property var onCancel: function() {
        PopupUtils.close(dialogue);
    }

    Button {
        text: acceptText
        color: UbuntuColors.green
        onClicked: optionSelected(onAccept)
    }

    Button {
        text: altText
        strokeColor: UbuntuColors.lightGrey
        visible: text.length > 0
        onClicked: optionSelected(onAlt)
    }

    Button {
        text: cancelText
        color: UbuntuColors.lightGrey
        onClicked: optionSelected(onCancel)
    }

    function optionSelected(option) {
        option();
        PopupUtils.close(dialogue);
    }
}
