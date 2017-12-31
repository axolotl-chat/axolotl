package store

import (
	qml "github.com/amlwwalker/qml"
	"github.com/janimo/textsecure"
)

type Contacts struct {
	Contacts []textsecure.Contact
	Len      int
}

var ContactsModel *Contacts = &Contacts{}

func (c *Contacts) GetContact(i int) textsecure.Contact {
	if i == -1 {
		return textsecure.Contact{}
	}
	return c.Contacts[i]
}
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
