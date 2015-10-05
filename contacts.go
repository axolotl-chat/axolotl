package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/llg/vcard"
	"github.com/godbus/dbus"
	"github.com/janimo/textsecure"
)

// getDesktopContacts reads the contacts for the desktop app from a file
func getDesktopContacts() ([]textsecure.Contact, error) {
	return textsecure.ReadContacts(filepath.Join(configDir, "contacts.yml"))
}

// getAddgetAddressBookContactsFromDBus gets the phone contacts via the address-book DBus service
func getAddressBookContactsFromDBus() ([]textsecure.Contact, error) {
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

func phoneFromVCardFile(file string) (string, error) {
	r, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer r.Close()

	di := vcard.NewDirectoryInfoReader(r)
	vc := &vcard.VCard{}
	vc.ReadFrom(di)
	if len(vc.Telephones) > 0 {
		return vc.Telephones[0].Number, nil
	}

	return "", errors.New("No phone number for contact.")
}

func parseVCards(vcardContacts []string) ([]textsecure.Contact, error) {
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
			contacts[i].Tel = strings.Replace(vc.Telephones[t].Number, " ", "", -1)
			i++
		}
	}
	return contacts[:i], nil
}

// getAddgetAddressBookContactsFromContentHub gets the phone contacts via the content hub
func getAddressBookContactsFromContentHub() ([]textsecure.Contact, error) {
	if exists(contactsFile) && vcardPath == "" {
		return textsecure.ReadContacts(contactsFile)
	}
	fileName := strings.Replace(vcardPath, "file://", "", 1)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	vcardContacts := strings.SplitAfter(string(b), "END:VCARD")
	contacts, err := parseVCards(vcardContacts)
	if err != nil {
		return nil, err
	}

	err = textsecure.WriteContacts(contactsFile, contacts)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}
