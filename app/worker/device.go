package worker

import (
	"errors"
	"image"
	"strings"

	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
	log "github.com/sirupsen/logrus"
)

var qr = false // TODO: WIP 831

func ReadQr(img image.Image) {
	if !qr {
		go interpretQR(img)

	}
}
func interpretQR(img image.Image) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Println("Qr Reader:", r)
	// 	}
	// }()
	// results := []string{}
	// scanner := grcode.NewScanner()
	// defer scanner.Close()
	// scanner.SetConfig(0, C.ZBAR_CFG_ENABLE, 1)
	// zImg := grcode.NewZbarImage(img)
	// defer zImg.Close()
	// scanner.Scan(zImg)
	// symbol := zImg.GetSymbol()
	// for ; symbol != nil; symbol = symbol.Next() {
	// 	results = append(results, symbol.Data())
	// }
	// if len(results) > 0 {
	// 	if strings.Contains(results[0], "tsdevice") {
	// 		log.Debugln("found tsdevice")
	// 		uuid, pub_key, err := extractUuidPubKey(results[0])
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		qr = true
	// 		pub_key = pub_key
	// 		textsecure.AddNewLinkedDevice(uuid, pub_key)
	// 		timer := time.NewTimer(10 * time.Second)
	// 		go func() {
	// 			<-timer.C
	// 			qr = false
	// 		}()
	// 	}
	// }

	// log.Println(results)
	// return result

}
func (a *TextsecureAPI) AddDevice() error {
	// log.Println("addDevice")
	// img := ui.Win.Snapshot()
	// ReadQr(img)
	return nil
}
func (a *TextsecureAPI) UnlinkDevice(id int) error {
	textsecure.UnlinkDevice(id)
	return nil
}
func (a *TextsecureAPI) RefreshDevices() error {
	// log.Println("addDevice")
	store.RefreshDevices()
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
