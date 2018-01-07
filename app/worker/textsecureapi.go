package worker

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gosexy/gettext"
	"github.com/morph027/textsecure"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/contact"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/settings"
	"github.com/nanu-c/textsecure-qml/app/store"
	"github.com/nanu-c/textsecure-qml/app/ui"
)

type TextsecureAPI struct {
	HasContacts     bool
	PushToken       string
	ActiveSessionID string
	PhoneNumber     string
}

var Api = &TextsecureAPI{}

func (Api *TextsecureAPI) Unregister() {
	os.RemoveAll(config.StorageDir)
	os.Remove(config.ConfigFile)
	os.Exit(1)
}
func (Api *TextsecureAPI) IdentityInfo(id string) string {
	myID := textsecure.MyIdentityKey()
	theirID, err := textsecure.ContactIdentityKey(id)
	if err != nil {
		log.Println(err)
	}
	return gettext.Gettext("Their identity (they read):") + "<br>" + fmt.Sprintf("% 0X", theirID) + "<br><br>" +
		gettext.Gettext("Your identity (you read):") + "<br><br>" + fmt.Sprintf("% 0X", myID)
}

func (Api *TextsecureAPI) ContactsImported(path string) {
	config.VcardPath = path

	err := store.RefreshContacts()
	if err != nil {
		ui.ShowError(err)
	}
}
func RunBackend() {
	Api = &TextsecureAPI{}
	client := &textsecure.Client{
		GetConfig:           config.GetConfig,
		GetPhoneNumber:      ui.GetPhoneNumber,
		GetVerificationCode: ui.GetVerificationCode,
		GetStoragePassword: func() string {
			password := ui.GetStoragePassword()
			log.Infof("Asking for password")

			if settings.SettingsModel.EncryptDatabase {
				log.Infof("Attempting to open encrypted datastore")
				var err error
				store.DS, err = store.NewStorage(password)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to open encrypted database")
				}
			}

			return password
		},
		MessageHandler:   messageHandler,
		ReceiptHandler:   receiptHandler,
		RegistrationDone: ui.RegistrationDone,
	}

	if config.IsPhone {
		client.GetLocalContacts = contact.GetAddressBookContactsFromContentHub
	} else {
		client.GetLocalContacts = contact.GetDesktopContacts
	}

	err := textsecure.Setup(client)
	if _, ok := err.(*strconv.NumError); ok {
		ui.ShowError(fmt.Errorf("Switching to unencrypted session store, removing %s\nThis will reset your sessions and reregister your phone.\n", config.StorageDir))
		os.RemoveAll(config.StorageDir)
		os.Exit(1)
	}
	if err != nil {
		ui.ShowError(err)
		return
	}

	Api.PhoneNumber = config.Config.Tel

	if helpers.Exists(config.ContactsFile) {
		Api.HasContacts = true
		store.RefreshContacts()
	}

	SendUnsentMessages()

	// Make sure to use names not numbers in session titles
	for _, s := range store.SessionsModel.Sess {
		s.Name = store.TelToName(s.Tel)
	}

	for {
		if err := textsecure.StartListening(); err != nil {
			log.Println(err)
			time.Sleep(3 * time.Second)
		}
	}
}
func (Api *TextsecureAPI) FilterContacts(sub string) {
	sub = strings.ToUpper(sub)

	fc := []textsecure.Contact{}
	for _, c := range store.ContactsModel.Contacts {
		if strings.Contains(strings.ToUpper(store.TelToName(c.Tel)), sub) {
			fc = append(fc, c)
		}
	}

	cm := &store.Contacts{fc, len(fc)}
	ui.Engine.Context().SetVar("contactsModel", cm)
}
func (Api *TextsecureAPI) SaveSettings() error {
	return settings.SaveSettings(settings.SettingsModel)
}
func (Api *TextsecureAPI) GetActiveSessionID() string {
	return Api.ActiveSessionID
}
func (Api *TextsecureAPI) SetActiveSessionID(sId string) {
	Api.ActiveSessionID = sId
}
