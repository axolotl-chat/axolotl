package store

import (
	"errors"
	"log"
	"strings"

	"github.com/signal-golang/textsecure"
)

type LinkedDevices struct {
	LinkedDevices []textsecure.DeviceInfo
	Len           int
}

var LinkedDevicesModel *LinkedDevices = &LinkedDevices{}

func (c *LinkedDevices) GetDevice(i int) textsecure.DeviceInfo {
	log.Println("[axolotl] get linked devices ", i)
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
	//qml.Changed(LinkedDevicesModel, &LinkedDevicesModel.Len)
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
	return nil
}
func RefreshDevices() (*LinkedDevices, error) {
	d, err := textsecure.LinkedDevices()
	if err != nil {
		return nil, err
	}

	LinkedDevicesModel.LinkedDevices = d[:]
	LinkedDevicesModel.Len = len(d)
	//qml.Changed(LinkedDevicesModel, &LinkedDevicesModel.Len)
	return LinkedDevicesModel, nil
}
func AddDevice(url string) error {
	uuid, pubKey, err := extractUuidPubKey(url)
	if err != nil {
		return err
	}
	textsecure.AddNewLinkedDevice(uuid, pubKey)
	RefreshDevices()
	return nil
}
func extractUuidPubKey(qr string) (string, string, error) {
	sUuid := strings.Index(qr, "=")
	eUuid := strings.Index(qr, "&")
	if sUuid > -1 {
		uuid := qr[sUuid+1 : eUuid]
		rest := qr[eUuid+1:]
		sPub_key := strings.Index(rest, "=")
		pub_key := rest[sPub_key+1:]
		pub_key = strings.Replace(pub_key, "%2F", "/", -1)
		pub_key = strings.Replace(pub_key, "%2B", "+", -1)
		return uuid, pub_key, nil
	} else {

		log.Println("[axolotl] no uuid/pubkey found")
		return "", "", errors.New("Wrong qr" + qr)
	}
}
