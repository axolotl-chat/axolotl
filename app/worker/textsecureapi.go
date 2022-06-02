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
	"github.com/nanu-c/axolotl/app/store"
	"github.com/nanu-c/axolotl/app/ui"
	"github.com/nanu-c/axolotl/app/webserver"
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
	Client          *textsecure.Client
	SessionStarted  bool
	IsEncrypted     bool
	Websocket       *webserver.WsApp
}

// var Api = &TextsecureAPI{}        // TODO
// var client = &textsecure.Client{} // TODO
// var sessionStarted = false        // TODO
// var api.IsEncrypted = true            // TODO

func NewTextsecureAPI() *TextsecureAPI {
	a := &TextsecureAPI{}
	a.Client = &textsecure.Client{}
	a.SessionStarted = false
	a.IsEncrypted = true

	return a
}

//unregister  signal id
func (a *TextsecureAPI) Unregister() {
	a.Websocket.App.Config.Unregister()
}

//get identitys
func (a *TextsecureAPI) IdentityInfo(id string) string {
	myID := textsecure.MyIdentityKey()
	theirID, err := textsecure.ContactIdentityKey(id)
	if err != nil {
		log.Errorln("[axolotl] IdentityInfo ", err)
	}
	return "Their identity (they read):" + "<br>" + fmt.Sprintf("% 0X", theirID) + "<br><br>" +
		"Your identity (you read):" + "<br><br>" + fmt.Sprintf("% 0X", myID)
}

func (a *TextsecureAPI) ContactsImported(path string) {
	a.Websocket.App.Config.VcardPath = path
	err := store.RefreshContacts()
	if err != nil {
		ui.ShowError(err, a.Websocket)
	}
}
func (a *TextsecureAPI) AddContact(name, phone, uuid string) {
	err := contact.AddContact(name, phone, uuid)
	if err != nil {
		ui.ShowError(err, a.Websocket)
	}
	err = store.RefreshContacts()
	if err != nil {
		ui.ShowError(err, a.Websocket)
	}
}

func RunBackend(wsApp *webserver.WsApp) {
	log.Debugf("[axolotl] Run Backend")
	api := NewTextsecureAPI()
	api.Websocket = wsApp
	store.DS.SetupDb("")
	go push.NotificationInit()
	ui.InitModels(api.Websocket)
	api.Websocket.App.Settings.Save()

	api.IsEncrypted = api.Websocket.App.Settings.EncryptDatabase
	if api.IsEncrypted {
		pw := ""
		for {
			pw = ui.GetEncryptionPw(api.Websocket)
			if store.DS.SetupDb(pw) {
				log.Debugf("[axolotl] DB Encrypted, ready to start")
				api.IsEncrypted = false
				break
			}
			ui.ShowError(errors.New("wrong password"), api.Websocket)
		}
	}
	api.SessionStarted = false
	api.Client = &textsecure.Client{
		GetConfig: api.Websocket.App.Config.GetConfig,
		GetPhoneNumber: func() string {
			if !api.Websocket.App.Settings.Registered {
				phoneNumber := ui.GetPhoneNumber(api.Websocket)
				return phoneNumber
			}
			return ""
		},
		GetVerificationCode: func() string {
			if !api.Websocket.App.Settings.Registered {
				log.Debugf("settings.SettingsModel.Registered = false")
				verificationCode := ui.GetVerificationCode(api.Websocket)
				api.Websocket.App.Settings.Registered = true
				return verificationCode
			}
			return ""
		},
		GetPin: func() string {
			pin := ui.GetPin(api.Websocket)
			return pin
		},
		GetCaptchaToken: func() string {
			captcha := ui.GetCaptchaToken(api.Websocket)
			return captcha
		},
		GetStoragePassword: func() string {
			password := ui.GetStoragePassword(api.Websocket)
			log.Debugf("[axolotl] Asking for password")

			if api.Websocket.App.Settings.EncryptDatabase {
				log.Debugf("[axolotl] Attempting to open encrypted datastore")
				var err error
				store.DS, err = store.NewStorage(password)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("[axolotl] Failed to open encrypted database")
				} else {
					api.StartAfterDecryption()

				}
			}
			return password
		},
		MessageHandler: func(m *textsecure.Message) {
			handler.MessageHandler(m, api.Websocket)
		},
		ReceiptHandler: func(s string, id uint32, ts uint64) {
			handler.ReceiptHandler(s, id, ts, api.Websocket)
		},
		ReceiptMessageHandler: func(m *textsecure.Message) {
			handler.ReceiptMessageHandler(m, api.Websocket)
		},
		CallMessageHandler: func(m *textsecure.Message) {
			handler.CallMessageHandler(m, api.Websocket)
		},
		TypingMessageHandler: func(m *textsecure.Message) {
			handler.TypingMessageHandler(m, api.Websocket)
		},
		SyncSentHandler: func(m *textsecure.Message, i uint64) {
			handler.SyncSentHandler(m, i, api.Websocket)
		},
		RegistrationDone: func() {
			ui.RegistrationDone(api.Websocket)
		},
		GetUsername: func() string {
			return ui.GetUsername(api.Websocket)
		},
	}

	if api.Websocket.App.Config.IsPhone {
		log.Debugf("[axolotl] IsPhone")
		api.Client.GetLocalContacts = contact.GetAddressBookContactsFromContentHub
	} else {
		log.Debugf("[axolotl] IsDesktop")
		api.Client.GetLocalContacts = contact.GetDesktopContacts
	}

	//Load Messages

	// Make sure to use names not numbers in session titles
	badHandshake := false
	for {
		ui.ClearError(api.Websocket)
		if !badHandshake {
			if !api.IsEncrypted {
				if !api.SessionStarted {
					log.Debugf("[axolotl] Start Session after Decryption")
					startSession(api)

				}
				if err := textsecure.StartListening(); err != nil {
					log.Debugln("[axolotl-ws] error:", err)
					if err.Error() == "websocket: bad handshake" {
						badHandshake = true
					}
					ui.ShowError(err, api.Websocket)
				}
			}
			time.Sleep(3 * time.Second)
		} else {
			ui.ShowError(errors.New("Your registration is faulty"))
			time.Sleep(10 * time.Minute)

		}

	}
}
func (a *TextsecureAPI) StartAfterDecryption() {

	log.Debugf("[axolotl] DB Encrypted, ready to start")
	a.IsEncrypted = false
}

func startSession(api *TextsecureAPI) {
	log.Debugf("[axolotl] starting Signal connection")
	tel := api.Websocket.App.Config.TsConfig.Tel
	err := textsecure.Setup(api.Client)
	if _, ok := err.(*strconv.NumError); ok {
		log.Errorf("[axolotl] startSession: %s", err)
		ui.ShowError(fmt.Errorf("[axolotl] switching to unencrypted session store, removing %s\nThis will reset your sessions and reregister your phone.\n", config.StorageDir))
		os.RemoveAll(api.Websocket.App.Config.StorageDir)
		os.Exit(1)
	}
	if err != nil {
		ui.ShowError(err, api.Websocket)
		return
	}
	api.SessionStarted = true
	api.PhoneNumber = tel
	if helpers.Exists(api.Websocket.App.Config.ContactsFile) {
		api.HasContacts = true
		store.RefreshContacts()
	}
	api.UUID = api.Websocket.App.Config.TsConfig.UUID
	if !api.Websocket.App.Config.TsConfig.AccountCapabilities.Gv2 {
		log.Debugln("[axolotl] gv2 not set, start gv2 migration")
		// enable gv2 capabilities
		api.Websocket.App.Config.TsConfig.AccountCapabilities = textsecureConfig.AccountCapabilities{
			Gv2:               true,
			SenderKey:         false,
			AnnouncementGroup: false,
			ChangeNumber:      false,
			Gv1Migration:      false,
		}
		// err := textsecure.WriteConfig(config.ConfigFile, config.Config)
		if err != nil {
			log.Debugln("[axolotl] gv2 migration save config: ", err)
		}
		textsecure.SetAccountCapabilities(api.Websocket.App.Config.TsConfig.AccountCapabilities)
		if err != nil {
			log.Debugln("[axolotl] gv2 migration: ", err)
		}
	}
	for _, s := range store.SessionsModel.Sess {
		s.Name = store.TelToName(s.Tel)
	}
	sender.SendUnsentMessages()

}

func (a *TextsecureAPI) SaveSettings() error {
	return a.Websocket.App.Settings.Save()
}

// GetActiveSessionID returns the active session id
func (a *TextsecureAPI) GetActiveSessionID() int64 {
	return store.ActiveSessionID
}

// SetActiveSessionID updates the active session id
func (a *TextsecureAPI) SetActiveSessionID(ID int64) {
	store.ActiveSessionID = ID
}

// LeaveChat reset the active session id
func (a *TextsecureAPI) LeaveChat() {
	store.ActiveSessionID = -1
}

// TgNotification turns the notification for the currently active chat on/off
func (a *TextsecureAPI) TgNotification(notification bool) {
	sess, err := store.SessionsModel.Get(store.ActiveSessionID)
	if err != nil {
		ui.ShowError(err, a.Websocket)
		return
	}
	sess.ToggleSessionNotification()
}
