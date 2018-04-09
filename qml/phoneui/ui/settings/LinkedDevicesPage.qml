import QtQuick 2.4
import Ubuntu.Components 1.3
import "../../components"
import Ubuntu.Content 1.1
import Ubuntu.Components.Popups 1.0


import "../../components/TelegramColors.js" as TelegramColors

TelegramPage {

    id: linkdevicePage
    title: i18n.tr("Link a Device")

    visible: true
    header: PageHeader {
      title: i18n.tr("Link a Device")
        id: pageHeader
        clip:true
        width: parent.width
        height: units.gu(6)
        trailingActionBar.actions:[
            Action {
                iconName: "add"
                text: i18n.tr("Add device")
                onTriggered: addDevice()
            },
            Action {
                iconName: "view-refresh"
                text: i18n.tr("Refresh devices")
                onTriggered: refresh()
            }
        ]
        leadingActionBar.actions:[
          Action {
            id: backAction
            iconName: "back"
            onTriggered:{
              back();
            }
          }
        ]
    }
    //head.actions: [
    //     Action {
    //         iconName: "add"
    //         text: i18n.tr("Add device")
    //         onTriggered: addDevice()
    //     },
    //     Action {
    //         iconName: "view-refresh"
    //         text: i18n.tr("Refresh devices")
    //         onTriggered: refresh()
    //     }
    // ]
    function addDevice(){
      pageStack.push(Qt.resolvedUrl("AddLinkedDevicesPage.qml"))
    }
    function refresh(){
      deviceModel.clear()
      textsecure.refreshDevices()
      for (var i =1;i<linkedDevicesModel.len;i++ ){
        deviceModel.append(linkedDevicesModel.getDevice(i))
      }
    }
    Column {
      anchors {
        top: pageHeader.bottom
        left: parent.left
        right: parent.right
        bottom: parent.bottom
      }

      // fruitModel.append({"cost": 5.95, "name":"Pizza"})
      ListModel {
        id: deviceModel
      }

    //  ListItem.ThinDivider {}
     Component{
          id: devicesDelegate
          // property var device : linkedDevicesModel.getDevice(index)
            ListItem{
                // text: name;
                ListItemLayout {
                    id: devicesListLayout
                    title.text: name
                }
                leadingActions: ListItemActions {
                        actions: [
                            Action {
                                iconName: "delete"
                                onTriggered: {
                                    PopupUtils.open(Qt.resolvedUrl("../dialogs/ConfirmationDialog.qml"),
                                    linkdevicePage, {
                                        title: i18n.tr("Delete selected Device?"),
                                        text: i18n.tr("This will permanently delete the selected Device."),
                                        onAccept: function() {
                                          // console.log(id);
                                            linkedDevicesModel.unlinkDevice(id)
                                            deviceModel.clear()
                                            textsecure.refreshDevices()
                                            for (var i =1;i<linkedDevicesModel.len;i++ ){
                                              deviceModel.append(linkedDevicesModel.getDevice(i))
                                            }
                                        }
                                    })
                                }
                            }
                        ]
                    }
            }

      }
      ListView {
        id: listView
        anchors.fill: parent;
        anchors.margins: 20
        clip: true
        model:deviceModel
        delegate: devicesDelegate
        // swipeEnabled : true







        Component.onCompleted: {
          refresh()

          // console.log(JSON.stringify(deviceModel));
        }
        function refresh(){
          deviceModel.clear()
          textsecure.refreshDevices()
          // console.log(linkedDevicesModel.getDevice(1).name);
          // console.log(linkedDevicesModel.len);
          for (var i =1;i<linkedDevicesModel.len;i++ ){
            deviceModel.append(linkedDevicesModel.getDevice(i))
          }
        }

     }
   }
 }
