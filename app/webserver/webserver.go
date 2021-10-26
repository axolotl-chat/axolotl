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
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/settings"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
	textsecureContacts "github.com/signal-golang/textsecure/contacts"
)

var (
	clients                = make(map[*websocket.Conn]bool)
	activeChat       int64 = -1
	codeVerification       = false
	profile          textsecureContacts.Contact
	broadcast        = make(chan []byte, 100)
	requestChannel   chan string
	upgrader         = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Run runs the webserver and the websocket
func Run() error {
	log.Printf("[axolotl] Starting axolotl ws")
	go syncClients()
	go attachmentServer()
	go websocketSender()
	webserver()
	return nil
}

func removeClientFromList(client *websocket.Conn) {
	log.Debugln("[axolotl-ws] remove client")
	delete(clients, client)
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection

	ws, err := upgrader.Upgrade(w, r, nil)
	clients[ws] = true
	if err != nil {
		log.Println(err)
	}
	log.Println("[axolotl] Client Connected", registered)

	// listen indefinitely for new messages coming
	// through on our WebSocket connection

	// send configs after establishing websocket connection
	SetGui()
	SetUiDarkMode()
	sendRegistrationStatus()

	if registered {
		UpdateChatList()
		UpdateContactList()
	}
	wsReader(ws)
}

func syncClients() {
	for {
		<-time.After(10 * time.Second)
		UpdateChatList()
		UpdateContactList()
	}
}
func wsReader(conn *websocket.Conn) {

	for {
		defer func() {
			if err := recover(); err != nil {
				log.Errorln("[axolotl-ws] wsReader panic occurred:", err)
				recoverFromWsPanic(conn)
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
		// fmt.Println(string(p), incomingMessage.Type)
		switch incomingMessage.Type {
		case "getChatList":
			sendChatList()
		case "getMessageList":
			getMessageListMessage := GetMessageListMessage{}
			json.Unmarshal([]byte(p), &getMessageListMessage)
			id := getMessageListMessage.ID
			activeChat = getMessageListMessage.ID
			store.ActiveSessionID = activeChat
			log.Debugln("[axolotl] Enter chat ", id)
			sendMessageList(id)
		case "setDarkMode":
			setDarkMode := SetDarkMode{}
			json.Unmarshal([]byte(p), &setDarkMode)
			log.Debugln("[axolotl] SetDarkMode ", setDarkMode.DarkMode)
			settings.SettingsModel.DarkMode = setDarkMode.DarkMode
			settings.SaveSettings(settings.SettingsModel)
			SetUiDarkMode()

		case "getMoreMessages":
			getMoreMessages := GetMoreMessages{}
			json.Unmarshal([]byte(p), &getMoreMessages)
			sendMoreMessageList(activeChat, getMoreMessages.LastID)
		case "createChat":
			createChatMessage := CreateChatMessage{}
			json.Unmarshal([]byte(p), &createChatMessage)
			log.Println("[axolotl] Create chat for ", createChatMessage.UUID)
			newChat := createChat(createChatMessage.UUID)
			activeChat = newChat.ID
			store.ActiveSessionID = activeChat
			requestEnterChat(activeChat)
			sendChatList()
		case "openChat":
			openChatMessage := OpenChatMessage{}
			json.Unmarshal([]byte(p), &openChatMessage)

			log.Println("[axolotl] Open chat with id: ", openChatMessage.Id)
			s, err := store.SessionsModel.Get(openChatMessage.Id)
			if err != nil {
				log.Errorln("[axolotl] Open chat with id: ", openChatMessage.Id, "failed", err)
			} else {
				activeChat = openChatMessage.Id
				// TODO: Avatar and profile handling for private chats, decryption is not yet done on the textsecure part
				// if !s.IsGroup {
				// p, _ := textsecure.GetProfile(s.Tel)
				// profile = p
				// }
				store.ActiveSessionID = activeChat
				sendCurrentChat(s)
			}
		case "leaveChat":
			activeChat = -1
			store.ActiveSessionID = -1
		case "createGroup":
			createGroupMessage := CreateGroupMessage{}
			json.Unmarshal([]byte(p), &createGroupMessage)
			log.Println("[axolotl] Create group ", createGroupMessage.Name)
			group, err := createGroup(createGroupMessage)
			if err != nil {
				log.Errorln("[axolotl] Create chat failed: ", err)

			} else {
				activeChat = group.ID
				store.ActiveSessionID = activeChat
				requestEnterChat(activeChat)
				sendContactList()
			}

		case "updateGroup":
			updateGroupMessage := UpdateGroupMessage{}
			json.Unmarshal([]byte(p), &updateGroupMessage)
			log.Println("[axolotl] Update group ", updateGroupMessage.ID)
			updateGroup(updateGroupMessage)
			requestEnterChat(store.ActiveSessionID)
			sendContactList()
		case "sendMessage":
			sendMessageMessage := SendMessageMessage{}
			json.Unmarshal([]byte(p), &sendMessageMessage)
			log.Debugln("[axolotl] send message to ", sendMessageMessage.To)
			updateMessageChannel := make(chan *store.Message)
			err, m := sender.SendMessageHelper(sendMessageMessage.To,
				sendMessageMessage.Message, "", updateMessageChannel)
			if err != nil || m == nil {
				log.Errorln("[axolotl] send message: ", err)
			} else {
				if err == nil {
					go MessageHandler(m)
				}
				// catch status
				go func() {
					m := <-updateMessageChannel
					go UpdateMessageHandlerWithSource(m)

				}()
			}
		case "delMessage":
			deleteMessageMessage := DelMessageMessage{}
			json.Unmarshal([]byte(p), &deleteMessageMessage)
			log.Println("[axolotl] delete message ", deleteMessageMessage.ID)
			store.DeleteMessage(deleteMessageMessage.ID)
		case "getContacts":
			go sendContactList()
		case "addContact":
			addContactMessage := AddContactMessage{}
			json.Unmarshal([]byte(p), &addContactMessage)
			contact.AddContact(addContactMessage.Name, addContactMessage.Phone)
			log.Infoln("[Axolotl] Add contact", addContactMessage.Name)
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			go sendContactList()
		case "requestCode":
			if requestChannel != nil {
				requestCodeMessage := RequestCodeMessage{}
				json.Unmarshal([]byte(p), &requestCodeMessage)
				requestChannel <- requestCodeMessage.Tel
			}
		case "sendCode":
			if requestChannel != nil {
				codeVerification = true
				sendCodeMessage := SendCodeMessage{}
				json.Unmarshal([]byte(p), &sendCodeMessage)
				requestChannel <- sendCodeMessage.Code
			}
		case "sendPin":
			if requestChannel != nil {
				codeVerification = true
				sendPinMessage := SendPinMessage{}
				json.Unmarshal([]byte(p), &sendPinMessage)
				requestChannel <- sendPinMessage.Pin
			}
		case "sendCaptchaToken":
			log.Debugln("[axolotl] got captcha")
			requestSmsVerificationCode = true
			if requestChannel != nil {
				sendCaptchaTokenMessage := SendCaptchaTokenMessage{}
				json.Unmarshal([]byte(p), &sendCaptchaTokenMessage)
				log.Debugln("[axolotl] got captcha2", sendCaptchaTokenMessage.Token)

				requestChannel <- sendCaptchaTokenMessage.Token
			}
		case "sendUsername":
			if requestChannel != nil {
				sendUsernameMessage := SendUsernameMessage{}
				json.Unmarshal([]byte(p), &sendUsernameMessage)
				requestChannel <- sendUsernameMessage.Username
				requestUsername = false
			}
		case "sendPassword":
			if requestChannel != nil {
				sendPasswordMessage := SendPasswordMessage{}
				json.Unmarshal([]byte(p), &sendPasswordMessage)
				requestChannel <- sendPasswordMessage.Pw
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
				settings.SettingsModel.EncryptDatabase = true
			} else {
				settings.SettingsModel.EncryptDatabase = false
			}
			settings.SaveSettings(settings.SettingsModel)
			os.Exit(0)
		case "getRegistrationStatus":
			sendRegistrationStatus()
		case "addDevice":
			addDeviceMessage := AddDeviceMessage{}
			json.Unmarshal([]byte(p), &addDeviceMessage)
			log.Println("[axolotl] add device " + addDeviceMessage.Url)
			if addDeviceMessage.Url != "" {
				if strings.Contains(addDeviceMessage.Url, "sgn") {
					fmt.Printf("found sgn")
					store.AddDevice(addDeviceMessage.Url)
				}
			}
		case "delDevice":
			delDeviceMessage := DelDeviceMessage{}
			json.Unmarshal([]byte(p), &delDeviceMessage)
			log.Println("[axolotl] delDevice", delDeviceMessage.Id)
			textsecure.UnlinkDevice(delDeviceMessage.Id)
			go sendDeviceList()
		case "getDevices":
			go sendDeviceList()
		case "unregister":
			config.Unregister()
		case "getConfig":
			sendConfig()
		case "refreshContacts":
			refreshContactsMessage := RefreshContactsMessage{}
			json.Unmarshal([]byte(p), &refreshContactsMessage)
			go refreshContacts(refreshContactsMessage.Url)
		case "uploadVcf":
			uploadVcf := UploadVcf{}
			json.Unmarshal([]byte(p), &uploadVcf)
			f, err := os.Create("import.vcf")
			if err != nil {
				log.Debugln("[axolotl] import vcf ", err)
				return
			}
			l, err := f.WriteString(uploadVcf.Vcf)
			if err != nil {
				log.Debugln("[axolotl] import vcf ", err)
				f.Close()
				return
			}
			log.Debugln("[axolotl] import vcf bytes written successfully", l)
			err = f.Close()
			if err != nil {
				log.Debugln("[axolotl] import vcf ", err)
				return
			}
			config.VcardPath = "import.vcf"
			contact.GetAddressBookContactsFromContentHub()
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			go sendContactList()
		case "delContact":
			delContactMessage := DelContactMessage{}
			json.Unmarshal([]byte(p), &delContactMessage)
			log.Println("[axolotl] delete contact", delContactMessage.ID)
			tmpContact := store.GetContactForTel(delContactMessage.ID)
			contact.DelContact(*tmpContact)
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			go sendContactList()
		case "editContact":
			editContactMessage := EditContactMessage{}
			json.Unmarshal([]byte(p), &editContactMessage)
			replaceContact := textsecureContacts.Contact{
				Tel:  editContactMessage.Phone,
				Name: editContactMessage.Name,
			}
			log.Debugln("[axolotl] editContact", editContactMessage.Name)
			contact.EditContact(store.ContactsModel.GetContact(editContactMessage.ID), replaceContact)
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			go sendContactList()
		case "delChat":
			delChatMessage := DelChatMessage{}
			json.Unmarshal([]byte(p), &delChatMessage)
			log.Debugln("[axolotl] deleteSession", delChatMessage.ID)
			store.DeleteSession(delChatMessage.ID)
			if err != nil {
				ShowError(err.Error())
			}
			sendChatList()
		case "sendAttachment":
			sendAttachmentMessage := SendAttachmentMessage{}
			json.Unmarshal([]byte(p), &sendAttachmentMessage)
			sendAttachment(sendAttachmentMessage)

			// store.DeleteSession(sendAttachmentMessage.ID)
			// err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
		case "uploadAttachment":
			uploadAttachmentMessage := UploadAttachmentMessage{}
			json.Unmarshal([]byte(p), &uploadAttachmentMessage)
			uploadSendAttachment(uploadAttachmentMessage)

			// store.DeleteSession(sendAttachmentMessage.ID)
			// err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
		case "toggleNotifcations":
			toggleNotificationsMessage := toggleNotificationsMessage{}
			json.Unmarshal([]byte(p), &toggleNotificationsMessage)
			log.Debugln("[axolotl] toggle notification for: ", toggleNotificationsMessage.Chat)
			s, err := store.SessionsModel.Get(toggleNotificationsMessage.Chat)
			if err != nil {
				ShowError(err.Error())
			}
			s.ToggleSessionNotifcation()
			sendCurrentChat(s)
			sendChatList()
		case "resetEncryption":
			resetEncryptionMessage := ResetEncryptionMessage{}
			json.Unmarshal([]byte(p), &resetEncryptionMessage)
			log.Debugln("[axolotl] reset encryption for: ", resetEncryptionMessage.Chat)
			s, err := store.SessionsModel.Get(resetEncryptionMessage.Chat)
			if err != nil {
				ShowError(err.Error())
			}
			m := s.Add("Secure session reset.", "", []store.Attachment{}, "", true, store.ActiveSessionID)
			m.Flags = helpers.MsgFlagResetSession
			store.SaveMessage(m)
			go sender.SendMessage(s, m)
			sendChatList()
		case "verifyIdentity":
			verifyIdentityMessage := verifyIdentityMessage{}
			json.Unmarshal([]byte(p), &verifyIdentityMessage)
			log.Debugln("[axolotl] identity information for: ", verifyIdentityMessage.Chat)
			s, err := store.SessionsModel.Get(verifyIdentityMessage.Chat)
			if err != nil {
				ShowError(err.Error())
			}
			fingerprintNumbers, fingerprintQRCode, err := textsecure.GetFingerprint(s.Tel)
			if err != nil {
				log.Debugln("[axolotl] identity information ", err)
			}
			sendIdentityInfo(fingerprintNumbers, fingerprintQRCode)
		}
	}
}

func webserver() {
	for {
		defer log.Errorln("[axolotl] webserver error")

		path := config.AxolotlWebDir

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
		http.HandleFunc("/ws", wsEndpoint)

		log.Error("[axolotl] webserver error", http.ListenAndServe(config.ServerHost+":"+config.ServerPort, nil))
	}

}
