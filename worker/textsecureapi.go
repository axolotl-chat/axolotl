package worker

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosexy/gettext"
	"github.com/morph027/textsecure"
	"github.com/nanu-c/textsecure-qml/contact"
	"github.com/nanu-c/textsecure-qml/models"
	"github.com/nanu-c/textsecure-qml/settings"
	"github.com/nanu-c/textsecure-qml/store"
	"github.com/nanu-c/textsecure-qml/ui"
)

type TextsecureAPI struct {
	HasContacts     bool
	PushToken       string
	ActiveSessionID string
	PhoneNumber     string
}

var Api = &TextsecureAPI{}

func (Api *TextsecureAPI) Unregister() {
	os.RemoveAll(store.StorageDir)
	os.Remove(store.ConfigFile)
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
	store.VcardPath = path

	err := store.RefreshContacts()
	if err != nil {
		ui.ShowError(err)
	}
}
func RunBackend() {
	Api = &TextsecureAPI{}
	client := &textsecure.Client{
		GetConfig:           store.GetConfig,
		GetPhoneNumber:      ui.GetPhoneNumber,
		GetVerificationCode: ui.GetVerificationCode,
		GetStoragePassword:  ui.GetStoragePassword,
		MessageHandler:      messageHandler,
		ReceiptHandler:      receiptHandler,
		RegistrationDone:    ui.RegistrationDone,
	}

	if store.IsPhone {
		client.GetLocalContacts = contact.GetAddressBookContactsFromContentHub
	} else {
		client.GetLocalContacts = contact.GetDesktopContacts
	}

	err := textsecure.Setup(client)
	if _, ok := err.(*strconv.NumError); ok {
		ui.ShowError(fmt.Errorf("Switching to unencrypted session store, removing %s\nThis will reset your sessions and reregister your phone.\n", store.StorageDir))
		os.RemoveAll(store.StorageDir)
		os.Exit(1)
	}
	if err != nil {
		ui.ShowError(err)
		return
	}

	Api.PhoneNumber = store.Config.Tel

	if models.Exists(store.ContactsFile) {
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
