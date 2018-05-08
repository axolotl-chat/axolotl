package worker

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gosexy/gettext"
	qml "github.com/nanu-c/qml-go"
	"github.com/nanu-c/textsecure"
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
	LogLevel        bool
}

var Api = &TextsecureAPI{}
var client = &textsecure.Client{}
var sessionStarted = false
var isEncrypted = true

//unregister  signal id
func (Api *TextsecureAPI) Unregister() {
	os.RemoveAll(config.StorageDir)
	os.Remove(config.ConfigFile)
	os.RemoveAll(config.DataDir)
	os.Remove(config.ContactsFile)
	settings.SettingsModel.EncryptDatabase = false
	os.Exit(1)
}

//get identitys
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
func (Api *TextsecureAPI) SetLogLevel() {
	// Api.LogLevel = !Api.LogLevel
	if Api.LogLevel == true {
		config.Config.LogLevel = "debug"
		log.SetLevel(log.DebugLevel)
		settings.SettingsModel.DebugLog = true
	} else {
		config.Config.LogLevel = "info"
		log.SetLevel(log.InfoLevel)
		settings.SettingsModel.DebugLog = false
		log.Infof("Set LogLevel to info")
	}
	Api.SaveSettings()
	textsecure.WriteConfig(config.ConfigFile, config.Config)
}
func RunBackend() {
	log.Debugf("Run Backend")
	notification()
	isEncrypted = settings.SettingsModel.EncryptDatabase
	sessionStarted = false
	Api = &TextsecureAPI{}
	Api.LogLevel = settings.SettingsModel.DebugLog
	client = &textsecure.Client{
		GetConfig: config.GetConfig,
		GetPhoneNumber: func() string {
			if !settings.SettingsModel.Registered {
				phoneNumber := ui.GetPhoneNumber()
				return phoneNumber
			}
			return ""
		},
		GetVerificationCode: func() string {
			if !settings.SettingsModel.Registered {
				log.Debugf("settings.SettingsModel.Registered = false")
				verificationCode := ui.GetVerificationCode()
				settings.SettingsModel.Registered = true
				return verificationCode
			}
			return ""
		},
		GetStoragePassword: func() string {
			password := ui.GetStoragePassword()
			log.Debugf("Asking for password")

			if settings.SettingsModel.EncryptDatabase {
				log.Debugf("Attempting to open encrypted datastore")
				var err error
				store.DS, err = store.NewStorage(password)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to open encrypted database")
				} else {
					Api.StartAfterDecryption()

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
	// start connection to openwhisper
	// if !isEncrypted {
	// 	startSession()
	// }

	//Load Messages

	// Make sure to use names not numbers in session titles

	for {
		if !isEncrypted {
			if !sessionStarted {
				log.Debugf("Start Session after Decryption")
				startSession()

			}
			if err := textsecure.StartListening(); err != nil {
				log.Debugln(err)
			}
		}
		time.Sleep(3 * time.Second)

	}
}
func (Api *TextsecureAPI) StartAfterDecryption() {

	log.Debugf("DB Encrypted, ready to start")
	isEncrypted = false
}

func startSession() {
	log.Debugf("starting Signal connection")
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
	sessionStarted = true
	Api.PhoneNumber = config.Config.Tel
	if helpers.Exists(config.ContactsFile) {
		Api.HasContacts = true
		store.RefreshContacts()
	}
	for _, s := range store.SessionsModel.Sess {
		s.Name = store.TelToName(s.Tel)
	}
	SendUnsentMessages()
	qml.Changed(store.SessionsModel, &store.SessionsModel.Len)

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
