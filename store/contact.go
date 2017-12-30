package store

import (
	"github.com/janimo/textsecure"
	qml "gopkg.in/qml.v1"
)

type Contacts struct {
	Contacts []textsecure.Contact
	Len      int
}

var ContactsModel *Contacts = &Contacts{}

func GetContactForTel(tel string) *textsecure.Contact {
	for _, c := range ContactsModel.Contacts {
		if c.Tel == tel {
			return &c
		}
	}
	return nil
}
func RefreshContacts() error {
	c, err := textsecure.GetRegisteredContacts()
	if err != nil {
		return err
	}

	ContactsModel.Contacts = c
	ContactsModel.Len = len(c)
	qml.Changed(ContactsModel, &ContactsModel.Len)
	return nil
}
