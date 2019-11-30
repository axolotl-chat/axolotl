package store

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/axolotl/app/config"
	yaml "gopkg.in/yaml.v2"
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
	log.Debugln("[axolotl] refresh contacts", len(c))

	if err != nil {
		c, _ = readRegisteredContacts(config.RegisteredContactsFile)
	} else {
		writeRegisteredContacts(config.RegisteredContactsFile, c)

	}
	ContactsModel.Contacts = c
	ContactsModel.Len = len(c)
	SessionsModel.UpdateSessionNames()
	if err != nil {
		return err
	}
	return nil
}

type yamlContacts struct {
	Contacts []textsecure.Contact
}

func writeRegisteredContacts(filename string, contacts []textsecure.Contact) error {
	c := &yamlContacts{contacts}
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0600)
}
func readRegisteredContacts(fileName string) ([]textsecure.Contact, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	contacts := &yamlContacts{}
	err = yaml.Unmarshal(b, contacts)
	if err != nil {
		return nil, err
	}
	return contacts.Contacts, nil
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
