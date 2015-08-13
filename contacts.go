package main

import (
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

// getAddgetAddressBookContacts gets the phone contacts via the address-book DBus service
func getAddressBookContacts() ([]textsecure.Contact, error) {
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

	contacts := make([]textsecure.Contact, len(vcardContacts))

	i := 0
	for _, c := range vcardContacts {
		di := vcard.NewDirectoryInfoReader(strings.NewReader(c))
		vc := &vcard.VCard{}
		vc.ReadFrom(di)
		if len(vc.Telephones) == 0 {
			continue
		}
		contacts[i].Name = vc.FormattedName
		contacts[i].Tel = strings.Replace(vc.Telephones[0].Number, " ", "", -1)
		i++
	}
	return contacts[:i], nil
}
