package worker

// #cgo darwin pkg-config: zbar
// #cgo LDFLAGS: -lzbar
// #include <zbar.h>
import "C"
import (
	"errors"
	"image"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/clsung/grcode"
	"github.com/morph027/textsecure"
	"github.com/nanu-c/textsecure-qml/app/store"
	"github.com/nanu-c/textsecure-qml/app/ui"
)

var qr = false

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
	results := []string{}
	scanner := grcode.NewScanner()
	defer scanner.Close()
	scanner.SetConfig(0, C.ZBAR_CFG_ENABLE, 1)
	zImg := grcode.NewZbarImage(img)
	defer zImg.Close()
	scanner.Scan(zImg)
	symbol := zImg.GetSymbol()
	for ; symbol != nil; symbol = symbol.Next() {
		results = append(results, symbol.Data())
	}
	if len(results) > 0 {
		if strings.Contains(results[0], "tsdevice") {
			log.Println("found tsdevice")
			uuid, pub_key, err := extractUuidPubKey(results[0])
			if err != nil {
				log.Fatal(err)
			}
			qr = true
			pub_key = pub_key
			textsecure.AddNewLinkedDevice(uuid, pub_key)
		}
	}

	// log.Println(results)
	// return result

}
func (Api *TextsecureAPI) AddDevice() error {
	// log.Println("addDevice")
	img := ui.Win.Snapshot()
	ReadQr(img)
	return nil
}
func (Api *TextsecureAPI) UnlinkDevice(id int) error {
	textsecure.UnlinkDevice(id)
	return nil
}
func (Api *TextsecureAPI) RefreshDevices() error {
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
		log.Println(uuid)
		log.Println(pub_key)
		pub_key = strings.Replace(pub_key, "%2F", "/", -1)
		log.Println(pub_key)
		return uuid, pub_key, nil
	} else {

		log.Println("no uuid/pubkey found")
		return "", "", errors.New("Wrong qr" + qr)
	}
}
