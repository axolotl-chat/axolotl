package contact

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mapaiva/vcard-go"
	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/ttacon/libphonenumber"
)

func PhoneFromVCardFile(file string) (string, error) {
	// r, err := os.Open(file)
	// if err != nil {
	// 	return "", err
	// }
	// defer r.Close()
	// // cards, err := vcard_go.GetVCards(file)
	//
	// if len(cards) > 0 {
	// 	return cards[0].FormattedName + " " + cards[0].Phone, nil
	// }

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

func AddContact(name string, phone string) error {
	contacts, err := textsecure.ReadContacts(config.ContactsFile)
	if err != nil {
		os.Create(config.ContactsFile)
	}
	contact := &textsecure.Contact{
		Name: name,
		Tel:  phone,
	}
	contacts = append(contacts, *contact)
	sort.Slice(contacts, func(i, j int) bool { return contacts[i].Name < contacts[j].Name })
	err = textsecure.WriteContacts(config.ContactsFile, contacts)
	if err != nil {
		return err
	}
	return nil
}
func DelContact(c textsecure.Contact) error {
	contacts, err := textsecure.ReadContacts(config.ContactsFile)
	log.Debugln("[axolotl] delete contact ", c)
	if err != nil {
		os.Create(config.ContactsFile)
	}
	newContactList := []textsecure.Contact{}
	for i := range contacts {
		if contacts[i].Tel != c.Tel {
			newContactList = append(newContactList, contacts[i])
		}
	}
	err = textsecure.WriteContacts(config.ContactsFile, newContactList)
	// cs, err := textsecure.ReadContacts(config.ContactsFile)
	if err != nil {
		return err
	}
	return nil
}
func EditContact(cContact textsecure.Contact, editContact textsecure.Contact) error {
	contacts, err := textsecure.ReadContacts(config.ContactsFile)

	if err != nil {
		os.Create(config.ContactsFile)
	}
	newContactList := []textsecure.Contact{}

	for i := range contacts {
		if contacts[i].Tel == cContact.Tel {
			newContactList = append(newContactList, editContact)
			// log.Debugln(id, contacts[i].Tel)
		} else {
			newContactList = append(newContactList, contacts[i])
			// log.Debugln(id, i)
		}
	}
	sort.Slice(newContactList, func(i, j int) bool { return newContactList[i].Name < newContactList[j].Name })
	err = textsecure.WriteContacts(config.ContactsFile, newContactList)
	if err != nil {
		return err
	}
	return nil
}

// getAddgetAddressBookContactsFromContentHub gets the phone contacts via the content hub
func GetAddressBookContactsFromContentHub() ([]textsecure.Contact, error) {
	if helpers.Exists(config.ContactsFile) && config.VcardPath == "" {
		return textsecure.ReadContacts(config.ContactsFile)
	}
	config.VcardPath = strings.TrimPrefix(config.VcardPath, "file://")
	newContacts, err := getContactsFromVCardFile(config.VcardPath)
	if err != nil {
		return nil, err
	}
	contacts, err := textsecure.ReadContacts(config.ContactsFile)
	//check for duplicates in the old contact list
	for _, c := range newContacts {
		found := false
		for i := range contacts {
			if contacts[i].Name == c.Name {
				contacts[i].Tel = c.Tel
				found = true
			}
		}
		if !found {
			contacts = append(contacts, c)
		}
	}
	//sort by name
	sort.Slice(contacts, func(i, j int) bool { return contacts[i].Name < contacts[j].Name })
	err = textsecure.WriteContacts(config.ContactsFile, contacts)
	if err != nil {
		return nil, err
	}
	// for i := range contacts {
	// 	// log.Infof(string(i))
	// }
	return contacts, nil
}

// getContactsFromVCardFile reads contacts from a VCF file
func getContactsFromVCardFile(path string) ([]textsecure.Contact, error) {
	vcardContacts, err := vcard.GetVCards(path)
	if err != nil {
		log.Fatal(err)
	}
	contacts := make([]textsecure.Contact, len(vcardContacts)*3)
	country := defaultCountry()

	i := 0
	for _, c := range vcardContacts {
		log.Debugln("Import " + c.FormattedName)
		if len(c.Phone) > 0 {
			contacts[i].Name = c.FormattedName
			contacts[i].Tel = FormatE164(c.Phone, country)
			if c.Photo != "" {
				b, err := base64.StdEncoding.DecodeString(c.Photo)
				if err == nil {
					contacts[i].Photo = string(b)
				}
			}
			i++
		}
	}
	return contacts, nil
}

func defaultCountry() string {
	num, _ := libphonenumber.Parse(config.Config.Tel, "")
	return libphonenumber.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
}
