package webserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/contact"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/sender"
	"github.com/nanu-c/textsecure-qml/app/settings"
	"github.com/nanu-c/textsecure-qml/app/store"
)

var clients = make(map[*websocket.Conn]bool)
var activeChat = ""

func Run() error {
	log.Printf("[axolotl] Starting axolotl ws")
	go sync()
	go attachmentServer()
	webserver()
	return nil
}

var requestChannel chan string

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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
	log.Println("[axolotl] Client Connected")
	// err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		// log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	SetGui()
	UpdateChatList()
	UpdateContactList()
	wsReader(ws)
}

func sync() {
	for {
		<-time.After(10 * time.Second)
		UpdateChatList()
		UpdateContactList()
	}
}
func wsReader(conn *websocket.Conn) {
	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		incomingMessage := Message{}
		json.Unmarshal([]byte(p), &incomingMessage)
		// fmt.Println(string(p), incomingMessage.Type)
		switch incomingMessage.Type {
		case "getChatList":
			sendChatList(conn)
		case "getMessageList":
			getMessageListMessage := GetMessageListMessage{}
			json.Unmarshal([]byte(p), &getMessageListMessage)
			id := getMessageListMessage.ID
			activeChat = getMessageListMessage.ID
			store.ActiveSessionID = activeChat
			log.Debugln("[axolotl] Enter chat ", id)
			sendMessageList(conn, id)
		case "getMoreMessages":
			getMoreMessages := GetMoreMessages{}
			json.Unmarshal([]byte(p), &getMoreMessages)
			sendMoreMessageList(conn, activeChat, getMoreMessages.LastID)
		case "createChat":
			createChatMessage := CreateChatMessage{}
			json.Unmarshal([]byte(p), &createChatMessage)
			log.Println("[axolotl] Create chat for ", createChatMessage.Tel)
			createChat(createChatMessage.Tel)
			activeChat = createChatMessage.Tel
			store.ActiveSessionID = activeChat
			s := store.SessionsModel.Get(createChatMessage.Tel)
			sendCurrentChat(conn, s)
		case "leaveChat":
			activeChat = ""
			store.ActiveSessionID = ""
		case "createGroup":
			createGroupMessage := CreateGroupMessage{}
			json.Unmarshal([]byte(p), &createGroupMessage)
			log.Println("Create group ", createGroupMessage.Name)
			group := createGroup(createGroupMessage)
			activeChat = group.Tel
			store.ActiveSessionID = activeChat
			requestEnterChat(activeChat)
			sendContactList(conn)
		case "sendMessage":
			sendMessageMessage := SendMessageMessage{}
			json.Unmarshal([]byte(p), &sendMessageMessage)
			log.Debugln("[axolotl] send message to ", sendMessageMessage.To)
			err, m := sender.SendMessageHelper(sendMessageMessage.To, sendMessageMessage.Message, "")
			if err == nil {
				go MessageHandler(m)
			}
		case "getContacts":
			go sendContactList(conn)
		case "addContact":
			log.Infoln("Add contact")
			addContactMessage := AddContactMessage{}
			json.Unmarshal([]byte(p), &addContactMessage)
			log.Println(addContactMessage.Name)
			contact.AddContact(addContactMessage.Name, addContactMessage.Phone)
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			go sendContactList(conn)
		case "requestCode":
			if requestChannel != nil {
				requestCodeMessage := RequestCodeMessage{}
				json.Unmarshal([]byte(p), &requestCodeMessage)
				requestChannel <- requestCodeMessage.Tel
			}
		case "sendCode":
			if requestChannel != nil {
				sendCodeMessage := SendCodeMessage{}
				json.Unmarshal([]byte(p), &sendCodeMessage)
				requestChannel <- sendCodeMessage.Code
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
			if settings.SettingsModel.EncryptDatabase {
				if !store.DS.DecryptDb(setPasswordMessage.CurrentPw) {
					// setError(i18n.tr("Incorrect old passphrase!"))
				}
			}
			if setPasswordMessage.Pw != "" {
				store.DS.EncryptDb(setPasswordMessage.Pw)
				settings.SettingsModel.EncryptDatabase = true
			} else {
				settings.SettingsModel.EncryptDatabase = false
			}
			settings.SaveSettings(settings.SettingsModel)
			os.Exit(0)
		case "getRegistrationStatus":
			if requestPassword {
				sendRequest(conn, "getEncryptionPw")
			} else if registered {
				sendRequest(conn, "registrationDone")
			} else {
				sendRequest(conn, "getPhoneNumber")
			}
		case "addDevice":
			addDeviceMessage := AddDeviceMessage{}
			json.Unmarshal([]byte(p), &addDeviceMessage)
			fmt.Println(addDeviceMessage.Url)
			if addDeviceMessage.Url != "" {
				if strings.Contains(addDeviceMessage.Url, "tsdevice") {
					fmt.Printf("found tsdevice")
					store.AddDevice(addDeviceMessage.Url)
				}
			}
		case "delDevice":
			delDeviceMessage := DelDeviceMessage{}
			json.Unmarshal([]byte(p), &delDeviceMessage)
			log.Println(delDeviceMessage.Id)
			textsecure.UnlinkDevice(delDeviceMessage.Id)
			go sendDeviceList(conn)
		case "getDevices":
			go sendDeviceList(conn)
		case "unregister":
			config.Unregister()
		case "getConfig":
			sendConfig(conn)
		case "refreshContacts":
			refreshContactsMessage := RefreshContactsMessage{}
			json.Unmarshal([]byte(p), &refreshContactsMessage)
			go refreshContacts(conn, refreshContactsMessage.Url)
		case "uploadVcf":
			uploadVcf := UploadVcf{}
			json.Unmarshal([]byte(p), &uploadVcf)
			f, err := os.Create("import.vcf")
			if err != nil {
				fmt.Println(err)
				return
			}
			l, err := f.WriteString(uploadVcf.Vcf)
			if err != nil {
				fmt.Println(err)
				f.Close()
				return
			}
			fmt.Println(l, "bytes written successfully")
			err = f.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
			config.VcardPath = "import.vcf"
			contact.GetAddressBookContactsFromContentHub()
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			go sendContactList(conn)
		case "delContact":
			log.Println("[axolotl] delete contact")
			delContactMessage := DelContactMessage{}
			json.Unmarshal([]byte(p), &delContactMessage)
			contact.DelContact(store.ContactsModel.GetContact(delContactMessage.ID))
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			go sendContactList(conn)
		case "editContact":
			editContactMessage := EditContactMessage{}
			json.Unmarshal([]byte(p), &editContactMessage)
			replaceContact := textsecure.Contact{
				Tel:  editContactMessage.Phone,
				Name: editContactMessage.Name,
			}
			log.Debugln("[axolotl ]editContact", editContactMessage.Name)
			contact.EditContact(store.ContactsModel.GetContact(editContactMessage.ID), replaceContact)
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			go sendContactList(conn)
		case "delChat":
			delChatMessage := DelChatMessage{}
			json.Unmarshal([]byte(p), &delChatMessage)
			store.DeleteSession(delChatMessage.ID)
			err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
			sendChatList(conn)
		case "sendAttachment":
			sendAttachmentMessage := SendAttachmentMessage{}
			json.Unmarshal([]byte(p), &sendAttachmentMessage)
			sendAttachment(sendAttachmentMessage)

			// store.DeleteSession(sendAttachmentMessage.ID)
			// err = store.RefreshContacts()
			if err != nil {
				ShowError(err.Error())
			}
		case "toggleNotifcations":
			toggleNotificationsMessage := ToggleNotificationsMessage{}
			json.Unmarshal([]byte(p), &toggleNotificationsMessage)
			log.Debugln("[axolotl] toggle notification for: ", toggleNotificationsMessage.Chat)
			s := store.SessionsModel.Get(toggleNotificationsMessage.Chat)
			s.ToggleSessionNotifcation()
			sendCurrentChat(conn, s)
			sendChatList(conn)
		case "resetEncryption":
			resetEncryptionMessage := ResetEncryptionMessage{}
			json.Unmarshal([]byte(p), &resetEncryptionMessage)
			log.Debugln("[axolotl] reset encryption for: ", resetEncryptionMessage.Chat)
			s := store.SessionsModel.Get(resetEncryptionMessage.Chat)
			m := s.Add("Secure session reset.", "", "", "", true, store.ActiveSessionID)
			m.Flags = helpers.MsgFlagResetSession
			store.SaveMessage(m)
			go sender.SendMessage(s, m)
			sendChatList(conn)
		case "verifyIdentity":
			verifyIdentityMessage := ToggleNotificationsMessage{}
			json.Unmarshal([]byte(p), &verifyIdentityMessage)
			log.Debugln("[axolotl] identity information for: ", verifyIdentityMessage.Chat)
			myID := textsecure.MyIdentityKey()
			theirID, err := textsecure.ContactIdentityKey(verifyIdentityMessage.Chat)
			if err != nil {
				log.Debugln("[axolotl] identity information ", err)
			}
			sendIdentityInfo(conn, myID, theirID)
		}
	}
}

func webserver() {
	http.Handle("/", http.FileServer(http.Dir("./axolotl-web/dist")))
	http.HandleFunc("/attachments", attachmentsHandler)
	http.HandleFunc("/avatars", avatarsHandler)
	http.HandleFunc("/ws", wsEndpoint)
	http.ListenAndServe(":9080", nil)
}
func print(stdout io.ReadCloser) {
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
}
