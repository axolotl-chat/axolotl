package worker

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/handler"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/settings"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/nanu-c/axolotl/app/ui"
	"github.com/signal-golang/textsecure"
	textsecureConfig "github.com/signal-golang/textsecure/config"

	log "github.com/sirupsen/logrus"
)

type TextsecureAPI struct {
	HasContacts     bool
	PushToken       string
	ActiveSessionID string
	PhoneNumber     string
	UUID            string
	LogLevel        bool
}

var Api = &TextsecureAPI{}
var client = &textsecure.Client{}
var sessionStarted = false
var isEncrypted = true

//unregister  signal id
func (Api *TextsecureAPI) Unregister() {
	config.Unregister()
}

//get identitys
func (Api *TextsecureAPI) IdentityInfo(id string) string {
	myID := textsecure.MyIdentityKey()
	theirID, err := textsecure.ContactIdentityKey(id)
	if err != nil {
		log.Errorln("[axolotl] IdentityInfo ", err)
	}
	return "Their identity (they read):" + "<br>" + fmt.Sprintf("% 0X", theirID) + "<br><br>" +
		"Your identity (you read):" + "<br><br>" + fmt.Sprintf("% 0X", myID)
}

func (Api *TextsecureAPI) ContactsImported(path string) {
	config.VcardPath = path
	err := store.RefreshContacts()
	if err != nil {
		ui.ShowError(err)
	}
}
func (Api *TextsecureAPI) AddContact(name string, phone string) {
	err := contact.AddContact(name, phone)
	if err != nil {
		ui.ShowError(err)
	}
	err = store.RefreshContacts()
	if err != nil {
		ui.ShowError(err)
	}
}
func (Api *TextsecureAPI) SetLogLevel() {
	// Api.LogLevel = !Api.LogLevel
	if !Api.LogLevel {
		config.Config.LogLevel = "debug"
		log.SetLevel(log.DebugLevel)
		settings.SettingsModel.DebugLog = true
		log.Infof("Set LogLevel to debug")
		Api.LogLevel = true
	} else {
		config.Config.LogLevel = "info"
		log.SetLevel(log.InfoLevel)
		settings.SettingsModel.DebugLog = false
		log.Infof("Set LogLevel to info")
		Api.LogLevel = false
	}
	Api.SaveSettings()
	// textsecure.WriteConfig(config.ConfigFile, config.Config)
}
func RegisterWithCrayfish(regisrationInfo *textsecure.RegistrationInfo) (*textsecure.CrayfishRegistration, error) {
	registration, err := CrayfishRegister(regisrationInfo)
	if err != nil {
		return nil, err
	}
	return registration, nil

}
func RunBackend() {
	log.Debugf("[axolotl] Run Backend")
	store.DS.SetupDb("")
	go push.NotificationInit()
	ui.InitModels()
	settings.SaveSettings(settings.SettingsModel)

	isEncrypted = settings.SettingsModel.EncryptDatabase
	if isEncrypted {
		pw := ""
		for {
			pw = ui.GetEncryptionPw()
			if store.DS.SetupDb(pw) {
				log.Debugf("[axolotl] DB Encrypted, ready to start")
				isEncrypted = false
				break
			} else {
				ui.ShowError(errors.New("wrong password"))
			}
		}
	}
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
		GetPin: func() string {
			pin := ui.GetPin()
			return pin
		},
		GetCaptchaToken: func() string {
			captcha := ui.GetCaptchaToken()
			return captcha
		},
		GetStoragePassword: func() string {
			password := ui.GetStoragePassword()
			log.Debugf("[axolotl] Asking for password")

			if settings.SettingsModel.EncryptDatabase {
				log.Debugf("[axolotl] Attempting to open encrypted datastore")
				var err error
				store.DS, err = store.NewStorage(password)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("[axolotl] Failed to open encrypted database")
				} else {
					Api.StartAfterDecryption()

				}
			}
			return password
		},
		MessageHandler:        handler.MessageHandler,
		ReceiptHandler:        handler.ReceiptHandler,
		ReceiptMessageHandler: handler.ReceiptMessageHandler,
		CallMessageHandler:    handler.CallMessageHandler,
		TypingMessageHandler:  handler.TypingMessageHandler,
		SyncSentHandler:       handler.SyncSentHandler,
		RegistrationDone:      ui.RegistrationDone,
		GetUsername:           ui.GetUsername,
		RegisterWithCrayfish:  RegisterWithCrayfish,
	}

	if config.IsPhone {
		client.GetLocalContacts = contact.GetAddressBookContactsFromContentHub
	} else {
		client.GetLocalContacts = contact.GetDesktopContacts
	}

	//Load Messages

	// Make sure to use names not numbers in session titles
	badHandshake := false
	for {
		ui.ClearError()
		if !badHandshake {
			if !isEncrypted {
				if !sessionStarted {
					log.Debugf("[axolotl] Start Session after Decryption")
					startSession()

				}
				if err := textsecure.StartListening(); err != nil {
					log.Debugln("[axolotl-ws] error:", err)
					if err.Error() == "websocket: bad handshake" {
						badHandshake = true
					}
					ui.ShowError(err)
				}
			}
			time.Sleep(3 * time.Second)
		} else {
			ui.ShowError(errors.New("Your registration is faulty"))
			time.Sleep(10 * time.Minute)

		}

	}
}
func (Api *TextsecureAPI) StartAfterDecryption() {

	log.Debugf("[axolotl] DB Encrypted, ready to start")
	isEncrypted = false
}

func startSession() {
	log.Debugf("[axolotl] starting Signal connection")
	err := textsecure.Setup(client)
	if _, ok := err.(*strconv.NumError); ok {
		ui.ShowError(fmt.Errorf("[axolotl] switching to unencrypted session store, removing %s\nThis will reset your sessions and reregister your phone.\n", config.StorageDir))
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
	Api.UUID = config.Config.UUID
	if !config.Config.AccountCapabilities.Gv2 {
		log.Debugln("[axolotl] gv2 not set, start gv2 migration")
		// enable gv2 capabilities
		config.Config.AccountCapabilities = textsecureConfig.AccountCapabilities{
			UUID:         false,
			Gv2:          true,
			Storage:      false,
			Gv1Migration: false,
		}
		// err := textsecure.WriteConfig(config.ConfigFile, config.Config)
		if err != nil {
			log.Debugln("[axolotl] gv2 migration save config: ", err)
		}
		textsecure.SetAccountCapabilities(config.Config.AccountCapabilities)
		if err != nil {
			log.Debugln("[axolotl] gv2 migration: ", err)
		}
	}
	for _, s := range store.SessionsModel.Sess {
		s.Name = store.TelToName(s.Tel)
	}
	sender.SendUnsentMessages()

}

func (Api *TextsecureAPI) SaveSettings() error {
	return settings.SaveSettings(settings.SettingsModel)
}

// GetActiveSessionID returns the active session id
func (Api *TextsecureAPI) GetActiveSessionID() int64 {
	return store.ActiveSessionID
}

// SetActiveSessionID updates the active session id
func (Api *TextsecureAPI) SetActiveSessionID(ID int64) {
	store.ActiveSessionID = ID
}

// LeaveChat reset the active session id
func (Api *TextsecureAPI) LeaveChat() {
	store.ActiveSessionID = -1
}

// TgNotification turns the notification for the currently active chat on/off
func (Api *TextsecureAPI) TgNotification(notification bool) {
	sess, err := store.SessionsModel.Get(store.ActiveSessionID)
	if err != nil {
		ui.ShowError(err)
		return
	}
	sess.ToggleSessionNotifcation()
}
