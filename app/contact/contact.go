package contact

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"bitbucket.org/llg/vcard"
	"github.com/godbus/dbus"
	vcard_go "github.com/mapaiva/vcard-go"
	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/ttacon/libphonenumber"
)

func PhoneFromVCardFile(file string) (string, error) {
	r, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer r.Close()
	cards, err := vcard_go.GetVCards(file)

	if len(cards) > 0 {
		return cards[0].FormattedName + " " + cards[0].Phone, nil
	}

	return "", errors.New("No phone number for contact.")
}

var pre = regexp.MustCompile("[^0-9+]")

func FormatE164(tel string, country string) string {
	if tel[0] == '+' {
		return pre.ReplaceAllString(tel, "")
	}
	num, err := libphonenumber.Parse(tel, country)
	if err != nil {
		log.Println(err)
		return tel
	}
	return libphonenumber.Format(num, libphonenumber.E164)
}
func GetDesktopContacts() ([]textsecure.Contact, error) {
	return textsecure.ReadContacts(filepath.Join(config.ConfigDir, "contacts.yml"))
}

// getAddgetAddressBookContactsFromDBus gets the phone contacts via the address-book DBus service
func GetAddressBookContactsFromDBus() ([]textsecure.Contact, error) {
	var o dbus.ObjectPath
	var vcardContacts []string

	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	obj := conn.Object("com.canonical.pim", "/com/canonical/pim/AddressBook")
	err = obj.Call("com.canonical.pim.AddressBook.query", 0, "", "", []string{}).Store(&o)
	if err != nil {
		return nil, err
	}
	obj2 := conn.Object("com.canonical.pim", o)
	err = obj2.Call("com.canonical.pim.AddressBookView.contactsDetails", 0, []string{}, int32(0), int32(-1)).Store(&vcardContacts)
	if err != nil {
		return nil, err
	}
	obj.Call("com.canonical.pim.AddressBook.close", 0)
	if err != nil {
		return nil, err
	}

	return parseVCards(vcardContacts)
}

// getAddgetAddressBookContactsFromContentHub gets the phone contacts via the content hub
func GetAddressBookContactsFromContentHub() ([]textsecure.Contact, error) {
	if helpers.Exists(config.ContactsFile) && config.VcardPath == "" {
		return textsecure.ReadContacts(config.ContactsFile)
	}
	config.VcardPath = strings.TrimPrefix(config.VcardPath, "file://")
	contacts, err := getContactsFromVCardFile(config.VcardPath)
	if err != nil {
		return nil, err
	}

	err = textsecure.WriteContacts(config.ContactsFile, contacts)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

// getContactsFromVCardFile reads contacts from a VCF file
func getContactsFromVCardFile(path string) ([]textsecure.Contact, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	vcardContacts := strings.SplitAfter(string(b), "END:VCARD")
	return parseVCards(vcardContacts)
}
func parseVCards(vcardContacts []string) ([]textsecure.Contact, error) {

	country := defaultCountry()

	// for now allocate space for 3 phones for each contact.
	// FIXME: make it cleaner by using up only as much space as needed.
	contacts := make([]textsecure.Contact, len(vcardContacts)*3)

	i := 0
	for _, c := range vcardContacts {
		di := vcard.NewDirectoryInfoReader(strings.NewReader(c))
		vc := &vcard.VCard{}
		vc.ReadFrom(di)
		for t := 0; t < len(vc.Telephones); t++ {
			contacts[i].Name = vc.FormattedName
			contacts[i].Tel = FormatE164(vc.Telephones[t].Number, country)
			if vc.Photo.Data != "" {
				b, err := base64.StdEncoding.DecodeString(vc.Photo.Data)
				if err == nil {
					contacts[i].Photo = string(b)
				} else {
					log.Printf("Parsing VCard %d %s\n", i, err.Error())
				}
			}
			i++
		}
	}
	return contacts[:i], nil
}

func defaultCountry() string {
	num, _ := libphonenumber.Parse(config.Config.Tel, "")
	return libphonenumber.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
}
