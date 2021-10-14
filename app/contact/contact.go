package contact

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/emersion/go-vcard"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	textsecureContacts "github.com/signal-golang/textsecure/contacts"
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
		log.Errorln("[axolotl] FormatE164", err)
		return tel
	}
	return libphonenumber.Format(num, libphonenumber.E164)
}
func GetDesktopContacts() ([]textsecureContacts.Contact, error) {
	return textsecureContacts.ReadContacts(filepath.Join(config.ConfigDir, "contacts.yml"))
}

func AddContact(name string, phone string) error {
	if phone[0] == '0' && phone[1] == '0' {
		phone = "+" + phone[2:]
	}
	contacts, err := textsecureContacts.ReadContacts(config.ContactsFile)
	if err != nil {
		os.Create(config.ContactsFile)
	}
	contact := &textsecureContacts.Contact{
		Name: name,
		Tel:  phone,
	}
	contacts = append(contacts, *contact)
	sort.Slice(contacts, func(i, j int) bool { return contacts[i].Name < contacts[j].Name })
	err = textsecureContacts.WriteContacts(config.ContactsFile, contacts)
	if err != nil {
		return err
	}
	return nil
}
func DelContact(c textsecureContacts.Contact) error {
	contacts, err := textsecureContacts.ReadContacts(config.ContactsFile)
	log.Debugln("[axolotl] delete contact ", c)
	if err != nil {
		os.Create(config.ContactsFile)
	}
	newContactList := []textsecureContacts.Contact{}
	for i := range contacts {
		if contacts[i].Tel != c.Tel {
			newContactList = append(newContactList, contacts[i])
		}
	}
	err = textsecureContacts.WriteContacts(config.ContactsFile, newContactList)
	// cs, err := textsecureContacts.ReadContacts(config.ContactsFile)
	if err != nil {
		return err
	}
	return nil
}
func EditContact(cContact textsecureContacts.Contact, editContact textsecureContacts.Contact) error {
	contacts, err := textsecureContacts.ReadContacts(config.ContactsFile)

	if err != nil {
		os.Create(config.ContactsFile)
	}
	newContactList := []textsecureContacts.Contact{}

	for i := range contacts {
		if contacts[i].Tel == cContact.Tel {
			newContactList = append(newContactList, editContact)
		} else {
			newContactList = append(newContactList, contacts[i])
		}
	}
	sort.Slice(newContactList, func(i, j int) bool { return newContactList[i].Name < newContactList[j].Name })
	err = textsecureContacts.WriteContacts(config.ContactsFile, newContactList)
	if err != nil {
		return err
	}
	return nil
}

// getAddgetAddressBookContactsFromContentHub gets the phone contacts via the content hub
func GetAddressBookContactsFromContentHub() ([]textsecureContacts.Contact, error) {
	if helpers.Exists(config.ContactsFile) && config.VcardPath == "" {
		return textsecureContacts.ReadContacts(config.ContactsFile)
	}
	config.VcardPath = strings.TrimPrefix(config.VcardPath, "file://")
	newContacts, err := getContactsFromVCardFile(config.VcardPath)
	if err != nil {
		return nil, err
	}
	contacts, err := textsecureContacts.ReadContacts(config.ContactsFile)
	if err != nil {
		return nil, err
	}
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
	err = textsecureContacts.WriteContacts(config.ContactsFile, contacts)
	if err != nil {
		return nil, err
	}
	// for i := range contacts {
	// 	// log.Infof(string(i))
	// }
	return contacts, nil
}

// getContactsFromVCardFile reads contacts from a VCF file
func getContactsFromVCardFile(path string) ([]textsecureContacts.Contact, error) {
	// vcardContacts, err := vcard.GetVCards(path)
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dec := vcard.NewDecoder(f)
	var contacts []textsecureContacts.Contact
	country := defaultCountry()
	for {
		card, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		name := card.PreferredValue(vcard.FieldFormattedName)
		log.Debugln("[axolotl] Import contact: " + name)
		telNums := card.Values(vcard.FieldTelephone)
		for index, tel := range telNums {
			tmp := textsecureContacts.Contact{}
			tmp.Name = name
			if index > 0 {
				tmp.Name = name + " " + strconv.Itoa(index)
			}
			tmp.Tel = FormatE164(tel, country)
			contacts = append(contacts, tmp)
		}
	}

	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	//
	// i := 0
	// for _, c := range vcardContacts {
	// 	if len(c.Phone) > 0 {
	// 		contacts[i].Name = c.FormattedName
	// 		contacts[i].Tel = FormatE164(c.Phone
	// 		if c.Photo != "" {
	// 			b, err := base64.StdEncoding.DecodeString(c.Photo)
	// 			if err == nil {
	// 				contacts[i].Photo = string(b)
	// 			}
	// 		}
	// 		i++
	// 	}
	// }
	return contacts, nil
}

func defaultCountry() string {
	num, _ := libphonenumber.Parse(config.Config.Tel, "")
	return libphonenumber.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
}
