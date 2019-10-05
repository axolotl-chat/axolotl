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
	"github.com/nanu-c/textsecure-qml/app/sender"
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
	log.Println("Client Connected")
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
			log.Debugln("Enter chat ", id)
			sendMessageList(conn, id)
		case "getMoreMessages":
			getMoreMessages := GetMoreMessages{}
			json.Unmarshal([]byte(p), &getMoreMessages)
			sendMoreMessageList(conn, activeChat, getMoreMessages.LastID)
		case "createChat":
			createChatMessage := CreateChatMessage{}
			json.Unmarshal([]byte(p), &createChatMessage)
			log.Println("Create chat for ", createChatMessage.Tel)
			createChat(createChatMessage.Tel)
			activeChat = createChatMessage.Tel
			store.ActiveSessionID = activeChat
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
			store.RefreshContacts()
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
			// sender.SendMessageHelper(sendMessageMessage.To, sendMessageMessage.Message, "")
		case "getRegistrationStatus":
			if registered {
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
		case "refreshContacts":
			refreshContactsMessage := RefreshContactsMessage{}
			json.Unmarshal([]byte(p), &refreshContactsMessage)
			config.VcardPath = refreshContactsMessage.Url
			contact.GetAddressBookContactsFromContentHub()
			store.RefreshContacts()
			go sendContactList(conn)
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
			store.RefreshContacts()
			go sendContactList(conn)
		case "delContact":
			fmt.Println("delContact")
			delContactMessage := DelContactMessage{}
			json.Unmarshal([]byte(p), &delContactMessage)
			contact.DelContact(delContactMessage.ID)
			go sendContactList(conn)
		case "editContact":
			fmt.Println("editContact")
			editContactMessage := EditContactMessage{}
			json.Unmarshal([]byte(p), &editContactMessage)
			replaceContact := textsecure.Contact{
				Tel:  editContactMessage.Phone,
				Name: editContactMessage.Name,
			}
			contact.EditContact(editContactMessage.ID, replaceContact)
			store.RefreshContacts()
			go sendContactList(conn)
		case "delChat":
			delChatMessage := DelChatMessage{}
			json.Unmarshal([]byte(p), &delChatMessage)
			store.DeleteSession(delChatMessage.ID)
			store.RefreshContacts()
			sendChatList(conn)
		}
	}
}

func webserver() {
	http.Handle("/", http.FileServer(http.Dir("./axolotl-web/dist")))
	http.HandleFunc("/attachments", attachmentsHandler)
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
