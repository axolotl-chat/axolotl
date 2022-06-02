package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/axolotl/app"
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
	textsecureContacts "github.com/signal-golang/textsecure/contacts"
)

// var ( // TODO
// 	clients                = make(map[*websocket.Conn]bool)
// 	activeChat       int64 = -1
// 	codeVerification = false
// 	profile        textsecureContacts.Contact
// 	broadcast      = make(chan []byte, 100)
// 	requestChannel chan string
// 	upgrader = websocket.Upgrader{
// 		ReadBufferSize:  1024,
// 		WriteBufferSize: 1024,
// 	}
// )

type WsApp struct {
	App              *app.App
	Clients          map[*websocket.Conn]bool
	ActiveChat       int64
	CodeVerification bool
	Profile          textsecureContacts.Contact
	Broadcast        chan []byte
	RequestChannel   chan string
	Upgrader         websocket.Upgrader
}

func NewWsApp(a *app.App) *WsApp {
	w := &WsApp{
		App:              a,
		Clients:          make(map[*websocket.Conn]bool),
		ActiveChat:       -1,
		CodeVerification: false,
		Broadcast:        make(chan []byte, 100),
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
	return w
}

// Run runs the webserver and the websocket
func Run(a *app.App) error {
	log.Printf("[axolotl] Starting axolotl ws")
	w := NewWsApp(a)
	go w.syncClients()
	go w.websocketSender()
	w.webserver()
	return nil
}

func (w *WsApp) removeClientFromList(client *websocket.Conn) {
	log.Debugln("[axolotl-ws] remove client")
	delete(w.Clients, client)
}

func wsEndpoint(w *WsApp) func(w http.ResponseWriter, r *http.Request) {
	handler := func(wr http.ResponseWriter, r *http.Request) {
		w.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		// upgrade this connection to a WebSocket
		// connection

		ws, err := w.Upgrader.Upgrade(wr, r, nil)
		w.Clients[ws] = true
		if err != nil {
			log.Println(err)
		}
		log.Println("[axolotl] Client Connected", registered)

		// listen indefinitely for new messages coming
		// through on our WebSocket connection

		// send configs after establishing websocket connection
		w.SetGui()
		w.SetUiDarkMode()
		w.sendRegistrationStatus()

		if registered {
			w.UpdateChatList()
			w.UpdateContactList()
		}
		w.wsReader(ws)
	}
	return handler
}

func (w *WsApp) syncClients() {
	for {
		<-time.After(10 * time.Second)
		w.UpdateChatList()
		w.UpdateContactList()
	}
}
func (w *WsApp) wsReader(conn *websocket.Conn) {

	for {
		defer func() {
			if err := recover(); err != nil {
				log.Errorln("[axolotl-ws] wsReader panic occurred:", err)
				w.recoverFromWsPanic(conn)
			}
		}()
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Errorln("[axolotl-ws] wsReader ", err)
			return
		}
		incomingMessage := Message{}
		json.Unmarshal([]byte(p), &incomingMessage)
		switch incomingMessage.Type {
		case "getChatList":
			w.sendChatList()
		case "getMessageList":
			getMessageListMessage := GetMessageListMessage{}
			json.Unmarshal([]byte(p), &getMessageListMessage)
			id := getMessageListMessage.ID
			w.ActiveChat = getMessageListMessage.ID
			store.ActiveSessionID = w.ActiveChat
			log.Debugln("[axolotl] Enter chat ", id)
			w.sendMessageList(id)
		case "setDarkMode":
			setDarkMode := SetDarkMode{}
			json.Unmarshal([]byte(p), &setDarkMode)
			log.Debugln("[axolotl] SetDarkMode ", setDarkMode.DarkMode)
			w.App.Settings.DarkMode = setDarkMode.DarkMode
			w.App.Settings.Save()
			w.SetUiDarkMode()

		case "getMoreMessages":
			getMoreMessages := GetMoreMessages{}
			json.Unmarshal([]byte(p), &getMoreMessages)
			w.sendMoreMessageList(getMoreMessages.LastID)
		case "createChat":
			createChatMessage := CreateChatMessage{}
			json.Unmarshal([]byte(p), &createChatMessage)
			log.Println("[axolotl] Create chat for ", createChatMessage.UUID)
			newChat := createChat(createChatMessage.UUID)
			w.ActiveChat = newChat.ID
			store.ActiveSessionID = w.ActiveChat
			w.requestEnterChat(w.ActiveChat)
			w.sendChatList()
		case "openChat":
			openChatMessage := OpenChatMessage{}
			json.Unmarshal([]byte(p), &openChatMessage)

			log.Println("[axolotl] Open chat with id: ", openChatMessage.Id)
			s, err := store.SessionsModel.Get(openChatMessage.Id)
			if err != nil {
				log.Errorln("[axolotl] Open chat with id: ", openChatMessage.Id, "failed", err)
			} else {
				w.ActiveChat = openChatMessage.Id
				// TODO: Avatar and profile handling for private chats, decryption is not yet done on the textsecure part
				// if !s.IsGroup {
				// p, _ := textsecure.GetProfile(s.Tel)
				// profile = p
				// }
				store.ActiveSessionID = w.ActiveChat
				w.sendCurrentChat(s)
			}
		case "leaveChat":
			w.ActiveChat = -1
			store.ActiveSessionID = -1
		case "createGroup":
			createGroupMessage := CreateGroupMessage{}
			json.Unmarshal([]byte(p), &createGroupMessage)
			log.Println("[axolotl] Create group ", createGroupMessage.Name)
			group, err := w.createGroup(createGroupMessage)
			if err != nil {
				log.Errorln("[axolotl] Create chat failed: ", err)

			} else {
				w.ActiveChat = group.ID
				store.ActiveSessionID = w.ActiveChat
				w.requestEnterChat(w.ActiveChat)
				w.sendContactList()
			}

		case "updateGroup":
			updateGroupMessage := UpdateGroupMessage{}
			json.Unmarshal([]byte(p), &updateGroupMessage)
			log.Println("[axolotl] Update group ", updateGroupMessage.ID)
			w.updateGroup(updateGroupMessage)
			w.requestEnterChat(store.ActiveSessionID)
			w.sendContactList()
		case "joinGroup":
			joinGroupMessage := JoinGroupMessage{}
			json.Unmarshal([]byte(p), &joinGroupMessage)
			log.Println("[axolotl] Join group ", joinGroupMessage.ID)
			go w.joinGroup(joinGroupMessage)
		case "sendMessage":
			sendMessageMessage := SendMessageMessage{}
			json.Unmarshal([]byte(p), &sendMessageMessage)
			log.Debugln("[axolotl] send message to ", sendMessageMessage.To)
			updateMessageChannel := make(chan *store.Message)
			m, err := sender.SendMessageHelper(sendMessageMessage.To,
				sendMessageMessage.Message, "", updateMessageChannel, false)
			if err != nil || m == nil {
				log.Errorln("[axolotl] send message: ", err)
			} else {
				if err == nil {
					go w.MessageHandler(m)
				}
				// catch status
				go func() {
					m := <-updateMessageChannel
					go w.UpdateMessageHandlerWithSource(m)
				}()
			}
		case "sendVoiceNote":
			sendVoiceNoteMessage := SendVoiceNoteMessage{}
			json.Unmarshal([]byte(p), &sendVoiceNoteMessage)
			w.uploadSendVoiceNote(sendVoiceNoteMessage)
		case "delMessage":
			deleteMessageMessage := DelMessageMessage{}
			json.Unmarshal([]byte(p), &deleteMessageMessage)
			log.Println("[axolotl] delete message ", deleteMessageMessage.ID)
			store.DeleteMessage(deleteMessageMessage.ID)
		case "getContacts":
			go w.sendContactList()
		case "addContact":
			addContactMessage := AddContactMessage{}
			json.Unmarshal([]byte(p), &addContactMessage)
			log.Infoln("[axolotl] Add contact", addContactMessage.Name)
			err = contact.AddContact(addContactMessage.Name, addContactMessage.Phone, addContactMessage.UUID)
			if err != nil {
				log.Errorln("[axolotl] Add contact failed: ", err)
			}
			err = store.RefreshContacts()
			if err != nil {
				log.Errorln("[axolotl] Refresh contacts failed: ", err)
				w.ShowError(err.Error())
			}
			go w.sendContactList()
		case "requestCode":
			if w.RequestChannel != nil {
				requestCodeMessage := RequestCodeMessage{}
				json.Unmarshal([]byte(p), &requestCodeMessage)
				w.RequestChannel <- requestCodeMessage.Tel
			}
		case "sendCode":
			if w.RequestChannel != nil {
				w.CodeVerification = true
				sendCodeMessage := SendCodeMessage{}
				json.Unmarshal([]byte(p), &sendCodeMessage)
				w.RequestChannel <- sendCodeMessage.Code
			}
		case "sendPin":
			if w.RequestChannel != nil {
				w.CodeVerification = true
				sendPinMessage := SendPinMessage{}
				json.Unmarshal([]byte(p), &sendPinMessage)
				w.RequestChannel <- sendPinMessage.Pin
			}
		case "sendCaptchaToken":
			log.Debugln("[axolotl] got captcha")
			requestSmsVerificationCode = true
			if w.RequestChannel != nil {
				sendCaptchaTokenMessage := SendCaptchaTokenMessage{}
				json.Unmarshal([]byte(p), &sendCaptchaTokenMessage)
				log.Debugln("[axolotl] got captcha2", sendCaptchaTokenMessage.Token)

				w.RequestChannel <- sendCaptchaTokenMessage.Token
			}
		case "sendUsername":
			if w.RequestChannel != nil {
				sendUsernameMessage := SendUsernameMessage{}
				json.Unmarshal([]byte(p), &sendUsernameMessage)
				w.RequestChannel <- sendUsernameMessage.Username
				requestUsername = false
			}
		case "sendPassword":
			if w.RequestChannel != nil {
				sendPasswordMessage := SendPasswordMessage{}
				json.Unmarshal([]byte(p), &sendPasswordMessage)
				w.RequestChannel <- sendPasswordMessage.Pw
			}
		case "setPassword":
			setPasswordMessage := SetPasswordMessage{}
			json.Unmarshal([]byte(p), &setPasswordMessage)
			log.Infoln("[axolotl] set password")
			// TODO: proof current password is correct
			// if settings.SettingsModel.EncryptDatabase {
			// 	if !store.DS.DecryptDb(setPasswordMessage.CurrentPw) {
			// 		// setError(i18n.tr("Incorrect old passphrase!"))
			// 	}
			// }
			if setPasswordMessage.Pw != "" {
				store.DS.EncryptDb(setPasswordMessage.Pw)
				w.App.Settings.EncryptDatabase = true
			} else {
				w.App.Settings.EncryptDatabase = false
			}
			w.App.Settings.Save()
			os.Exit(0)
		case "getRegistrationStatus":
			w.sendRegistrationStatus()
		case "addDevice":
			addDeviceMessage := AddDeviceMessage{}
			json.Unmarshal([]byte(p), &addDeviceMessage)
			log.Println("[axolotl] add device " + addDeviceMessage.Url)
			if addDeviceMessage.Url != "" {
				if strings.Contains(addDeviceMessage.Url, "sgnl") {
					fmt.Printf("found sgnl")
					store.AddDevice(addDeviceMessage.Url)
				}
			}
		case "delDevice":
			delDeviceMessage := DelDeviceMessage{}
			json.Unmarshal([]byte(p), &delDeviceMessage)
			log.Println("[axolotl] delDevice", delDeviceMessage.Id)
			textsecure.UnlinkDevice(delDeviceMessage.Id)
			go w.sendDeviceList()
		case "getDevices":
			go w.sendDeviceList()
		case "unregister":
			w.App.Config.Unregister()
		case "getConfig":
			w.sendConfig()
		case "refreshContacts":
			refreshContactsMessage := RefreshContactsMessage{}
			json.Unmarshal([]byte(p), &refreshContactsMessage)
			go w.refreshContacts(refreshContactsMessage.Url)
		case "uploadVcf":
			uploadVcf := UploadVcf{}
			json.Unmarshal([]byte(p), &uploadVcf)
			f, err := os.Create("import.vcf")
			if err != nil {
				log.Debugln("[axolotl] import vcf ", err)
			} else {
				l, err := f.WriteString(uploadVcf.Vcf)
				if err != nil {
					log.Errorln("[axolotl] import vcf ", err)
					f.Close()
				}
				log.Debugln("[axolotl] import vcf bytes written successfully", l)
				err = f.Close()
				if err != nil {
					log.Errorln("[axolotl] import vcf ", err)
				} else {
					//non blocking vcf import
					go w.importVcf()
				}

			}

		case "delContact":
			delContactMessage := DelContactMessage{}
			json.Unmarshal([]byte(p), &delContactMessage)
			log.Println("[axolotl] delete contact", delContactMessage.ID)
			tmpContact := store.GetContactForTel(delContactMessage.ID)
			contact.DelContact(*tmpContact)
			err = store.RefreshContacts()
			if err != nil {
				w.ShowError(err.Error())
			}
			go w.sendContactList()
		case "editContact":
			editContactMessage := EditContactMessage{}
			json.Unmarshal([]byte(p), &editContactMessage)
			replaceContact := textsecureContacts.Contact{
				Tel:  editContactMessage.Phone,
				Name: editContactMessage.Name,
				UUID: editContactMessage.UUID,
			}
			log.Debugln("[axolotl] editContact", editContactMessage.Name)
			contact.EditContact(store.ContactsModel.GetContact(editContactMessage.ID), replaceContact)
			// todo: dont refresh contacts when only the name is changed to avoid hitting the server limit
			err = store.RefreshContacts()
			if err != nil {
				w.ShowError(err.Error())
			}
			go w.sendContactList()
		case "delChat":
			delChatMessage := DelChatMessage{}
			json.Unmarshal([]byte(p), &delChatMessage)
			log.Debugln("[axolotl] deleteSession", delChatMessage.ID)
			store.DeleteSession(delChatMessage.ID)
			if err != nil {
				w.ShowError(err.Error())
			}
			w.sendChatList()
		case "sendAttachment":
			sendAttachmentMessage := SendAttachmentMessage{}
			json.Unmarshal([]byte(p), &sendAttachmentMessage)
			w.sendAttachment(sendAttachmentMessage)

			// store.DeleteSession(sendAttachmentMessage.ID)
			// err = store.RefreshContacts()
			if err != nil {
				w.ShowError(err.Error())
			}
		case "uploadAttachment":
			uploadAttachmentMessage := UploadAttachmentMessage{}
			json.Unmarshal([]byte(p), &uploadAttachmentMessage)
			w.uploadSendAttachment(uploadAttachmentMessage)

			// store.DeleteSession(sendAttachmentMessage.ID)
			// err = store.RefreshContacts()
			if err != nil {
				w.ShowError(err.Error())
			}
		case "setLogLevel":
			setLogLevelMessage := SetLogLevelMessage{}
			json.Unmarshal([]byte(p), &setLogLevelMessage)
			w.App.Config.SetLogLevel(setLogLevelMessage.Level)
		case "toggleNotifications":
			toggleNotificationsMessage := toggleNotificationsMessage{}
			json.Unmarshal([]byte(p), &toggleNotificationsMessage)
			log.Debugln("[axolotl] toggle notification for: ", toggleNotificationsMessage.Chat)
			s, err := store.SessionsModel.Get(toggleNotificationsMessage.Chat)
			if err != nil {
				w.ShowError(err.Error())
			}
			s.ToggleSessionNotification()
			w.sendCurrentChat(s)
			w.sendChatList()
		case "resetEncryption":
			resetEncryptionMessage := ResetEncryptionMessage{}
			json.Unmarshal([]byte(p), &resetEncryptionMessage)
			log.Debugln("[axolotl] reset encryption for: ", resetEncryptionMessage.Chat)
			s, err := store.SessionsModel.Get(resetEncryptionMessage.Chat)
			if err != nil {
				w.ShowError(err.Error())
			}
			m := s.Add("Secure session reset.", "", []store.Attachment{}, "", true, store.ActiveSessionID)
			m.Flags = helpers.MsgFlagResetSession
			store.SaveMessage(m)
			go sender.SendMessage(s, m, false)
			w.sendChatList()
		case "verifyIdentity":
			verifyIdentityMessage := verifyIdentityMessage{}
			json.Unmarshal([]byte(p), &verifyIdentityMessage)
			log.Debugln("[axolotl] identity information for: ", verifyIdentityMessage.Chat)
			s, err := store.SessionsModel.Get(verifyIdentityMessage.Chat)
			if err != nil {
				w.ShowError(err.Error())
			}
			fingerprintNumbers, fingerprintQRCode, err := textsecure.GetFingerprint(s.UUID, s.Tel)
			if err != nil {
				log.Debugln("[axolotl] identity information ", err)
			}
			w.sendIdentityInfo(fingerprintNumbers, fingerprintQRCode)
		default:
			if incomingMessage.Type != "" {
				log.Debugf("[axolotl] unknown message type: %v", incomingMessage)
			}
		}

	}
}

func (w *WsApp) webserver() {
	for {
		defer log.Errorln("[axolotl] webserver error")

		path := w.App.Config.AxolotlWebDir

		axolotlWebDirEnv := os.Getenv("AXOLOTL_WEB_DIR")
		if len(axolotlWebDirEnv) > 0 {
			path = axolotlWebDirEnv
		}

		snapEnv := os.Getenv("SNAP")
		if len(snapEnv) > 0 && !strings.Contains(snapEnv, "/snap/go/") {
			path = os.Getenv("SNAP") + "/bin/axolotl-web/"
		}
		log.Debugln("[axolotl] Using axolotl-web path", path)

		http.Handle("/", http.FileServer(http.Dir(path)))
		http.HandleFunc("/attachments", attachmentsHandler)
		http.HandleFunc("/avatars", avatarsHandler)
		http.HandleFunc("/ws", wsEndpoint(w))

		log.Error("[axolotl] webserver error", http.ListenAndServe(w.App.Config.ServerHost+":"+w.App.Config.ServerPort, nil))
	}

}
