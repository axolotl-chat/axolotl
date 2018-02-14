package store

import (
	"log"

	"github.com/morph027/textsecure"
	qml "github.com/nanu-c/qml-go"
)

type LinkedDevices struct {
	LinkedDevices []textsecure.DeviceInfo
	Len           int
}

var LinkedDevicesModel *LinkedDevices = &LinkedDevices{}

func (c *LinkedDevices) GetDevice(i int) textsecure.DeviceInfo {
	log.Println(i)
	if i == -1 {
		return textsecure.DeviceInfo{}
	}
	if i >= LinkedDevicesModel.Len {

		return textsecure.DeviceInfo{}
	}

	tmp := LinkedDevicesModel.LinkedDevices[i]
	return tmp
}
func (c *LinkedDevices) RefreshDevices() error {
	d, err := textsecure.LinkedDevices()
	if err != nil {
		return err
	}

	LinkedDevicesModel.LinkedDevices = d[:]
	LinkedDevicesModel.Len = len(d)
	qml.Changed(LinkedDevicesModel, &LinkedDevicesModel.Len)
	return nil
}
func (c *LinkedDevices) UnlinkDevice(id int) error {
	textsecure.UnlinkDevice(id)
	return nil
}
func (c *LinkedDevices) DeleteDevice() error {
	d, err := textsecure.LinkedDevices()
	if err != nil {
		return err
	}

	LinkedDevicesModel.LinkedDevices = d[:]
	LinkedDevicesModel.Len = len(d)
	qml.Changed(LinkedDevicesModel, &LinkedDevicesModel.Len)
	return nil
}
func RefreshDevices() error {
	d, err := textsecure.LinkedDevices()
	if err != nil {
		return err
	}

	LinkedDevicesModel.LinkedDevices = d[:]
	LinkedDevicesModel.Len = len(d)
	qml.Changed(LinkedDevicesModel, &LinkedDevicesModel.Len)
	return nil
}
