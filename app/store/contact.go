package store

import (
	qml "github.com/nanu-c/qml-go"
	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/config"
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
	qml.Changed(SessionsModel, &SessionsModel.Len)

	return nil
}
func TelToName(tel string) string {
	if g, ok := Groups[tel]; ok {
		return g.Name
	}
	for _, c := range ContactsModel.Contacts {
		if c.Tel == tel {
			return c.Name
		}
	}
	if tel == config.Config.Tel {
		return "Me"
	}
	return tel
}
