package store

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/signal-golang/textsecure"
	textsecureContacts "github.com/signal-golang/textsecure/contacts"

	yaml "gopkg.in/yaml.v2"
)

type Contacts struct {
	Contacts []textsecureContacts.Contact
	Len      int
}

// var ContactsModel *Contacts = &Contacts{} // TODO: WIP 831

func (c *Contacts) GetContact(i int) textsecureContacts.Contact {
	if i == -1 {
		return textsecureContacts.Contact{}
	}
	return c.Contacts[i]
}
func (c *Contacts) GetContactForTel(tel string) *textsecureContacts.Contact {
	for _, contact := range c.Contacts {
		if contact.Tel == tel {
			return &contact
		}
	}
	return nil
}
func (c *Contacts) GetContactForUUID(uuid string) *textsecureContacts.Contact {
	for _, contact := range c.Contacts {
		if contact.UUID == uuid {
			return &contact
		}
	}
	return nil
}
func (s *Store) RefreshContacts() error {
	registeredContactsFile := config.GetRegisteredContactsFile()
	contacts, err := textsecure.GetRegisteredContacts()
	if err != nil {
		log.Errorln("[axolotl] RefreshContacts", err)
		// when refresh fails, we load the old contacts
		contacts, _ = readRegisteredContacts(registeredContactsFile)
		return err
	} else {
		writeRegisteredContacts(registeredContactsFile, contacts)
	}
	log.Debugln("[axolotl] Refresh contacts count: ", len(contacts))
	s.Contacts.Contacts = contacts
	s.Contacts.Len = len(contacts)
	s.UpdateSessionNames()
	if err != nil {
		return err
	}
	return nil
}

type yamlContacts struct {
	Contacts []textsecureContacts.Contact
}

func writeRegisteredContacts(filename string, contacts []textsecureContacts.Contact) error {
	c := &yamlContacts{contacts}
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0600)
}
func readRegisteredContacts(fileName string) ([]textsecureContacts.Contact, error) {
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

func (c *Contacts) TelToName(tel string) string {
	if g, ok := Groups[tel]; ok {
		return g.Name
	}
	for _, c := range c.Contacts {
		if c.Tel == tel {
			return c.Name
		}
	}
	config := &config.Config{} // TODO: WIP 831: wire through config
	if tel == config.GetMyNumber() {
		return "Me"
	}
	return tel
}
