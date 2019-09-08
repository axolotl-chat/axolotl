package webserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/contact"
	"github.com/nanu-c/textsecure-qml/app/sender"
	"github.com/nanu-c/textsecure-qml/app/store"
)

var clients = make(map[*websocket.Conn]bool)
var activeChat = ""

type MessageListEnvelope struct {
	MessageList *store.MessageList
}
type MoreMessageListEnvelope struct {
	MoreMessageList *store.MessageList
}
type ChatListEnvelope struct {
	ChatList []*store.Session
}
type ContactListEnvelope struct {
	ContactList []textsecure.Contact
}

func Run() error {
	log.Printf("Starting Axolotl-gui")
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
	wsReader(ws)
}

type Message struct {
	Type string                 `json:"request"`
	Data map[string]interface{} `json:"-"` // Rest of the fields should go here.
}
type GetMessageListMessage struct {
	Type string `json:"request"`
	ID   string `json:"id"`
}
type GetMoreMessages struct {
	Type   string `json:"request"`
	LastID string `json:"lastId"`
}
type SendMessageMessage struct {
	Type    string `json:"request"`
	To      string `json:"to"`
	Message string `json:"message"`
}
type RequestCodeMessage struct {
	Type string `json:"request"`
	Tel  string `json:"tel"`
}
type SendCodeMessage struct {
	Type string `json:"request"`
	Code string `json:"code"`
}

func sync() {
	for {
		<-time.After(10 * time.Second)
		go UpdateChatList()
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
		if incomingMessage.Type == "getChatList" {
			sendChatList(conn)
		} else if incomingMessage.Type == "getMessageList" {
			getMessageListMessage := GetMessageListMessage{}
			json.Unmarshal([]byte(p), &getMessageListMessage)
			id := getMessageListMessage.ID
			activeChat = getMessageListMessage.ID
			sendMessageList(conn, id)
		} else if incomingMessage.Type == "getMoreMessages" {
			getMoreMessages := GetMoreMessages{}
			json.Unmarshal([]byte(p), &getMoreMessages)
			sendMoreMessageList(conn, activeChat, getMoreMessages.LastID)
		} else if incomingMessage.Type == "leaveChat" {
			activeChat = ""
		} else if incomingMessage.Type == "sendMessage" {
			sendMessageMessage := SendMessageMessage{}
			json.Unmarshal([]byte(p), &sendMessageMessage)
			sender.SendMessageHelper(sendMessageMessage.To, sendMessageMessage.Message, "")
			go UpdateChatList()
		} else if incomingMessage.Type == "getContacts" {
			go sendContactList(conn)
		} else if incomingMessage.Type == "addContact" {
			fmt.Printf("addContact")
			contact.AddContact("Aaron", "+436706070770")
			go sendContactList(conn)
		} else if incomingMessage.Type == "requestCode" {
			if requestChannel != nil {
				requestCodeMessage := RequestCodeMessage{}
				json.Unmarshal([]byte(p), &requestCodeMessage)
				requestChannel <- requestCodeMessage.Tel
			}
		} else if incomingMessage.Type == "sendCode" {
			if requestChannel != nil {
				sendCodeMessage := SendCodeMessage{}
				json.Unmarshal([]byte(p), &sendCodeMessage)
				requestChannel <- sendCodeMessage.Code
			}
			// sender.SendMessageHelper(sendMessageMessage.To, sendMessageMessage.Message, "")
		} else if incomingMessage.Type == "getRegistrationStatus" {
			if registered {
				sendRequest(conn, "registrationDone")
			} else {
				sendRequest(conn, "getPhoneNumber")
			}

		}

	}
}
func attachmentsHandler(w http.ResponseWriter, r *http.Request) {
	Filename := r.URL.Query().Get("file")
	fmt.Println(Filename)
	if Filename == "" {
		//Get not set, send a 400 bad request
		http.Error(w, "Get 'file' not specified in url.", 400)
		return
	}
	fmt.Println("Client requests: " + Filename)

	//Check if file exists and open
	Openfile, err := os.Open(Filename)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(w, "File not found.", 404)
		return
	}
	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(w, Openfile) //'Copy' the file to the client
	return
}
func attachmentServer() {
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
func sendChatList(client *websocket.Conn) {
	var err error
	chatListEnvelope := &ChatListEnvelope{
		ChatList: store.SessionsModel.Sess,
	}
	message := &[]byte{}
	*message, err = json.Marshal(chatListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println(err)
		return
	}
}
func sendContactList(client *websocket.Conn) {
	var err error
	contactListEnvelope := &ContactListEnvelope{
		ContactList: store.ContactsModel.Contacts,
	}
	message := &[]byte{}
	*message, err = json.Marshal(contactListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println(err)
		return
	}
}
func sendMessageList(client *websocket.Conn, id string) {
	message := &[]byte{}
	err, messageList := store.SessionsModel.GetMessageList(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	chatListEnvelope := &MessageListEnvelope{
		MessageList: messageList,
	}
	*message, err = json.Marshal(chatListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println(err)
		return
	}
}
func sendMoreMessageList(client *websocket.Conn, id string, lastId string) {
	message := &[]byte{}
	err, messageList := store.SessionsModel.GetMoreMessageList(id, lastId)
	if err != nil {
		fmt.Println(err)
		return
	}
	moreMessageListEnvelope := &MoreMessageListEnvelope{
		MoreMessageList: messageList,
	}
	*message, err = json.Marshal(moreMessageListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println(err)
		return
	}
}

var test = false

func UpdateChatList() {

	if activeChat == "" {
		// 	for client := range clients {
		// 		fmt.Printf("blub")
		// 		// sendMessageList(client, activeChat)
		// 	}
		// } else {
		for client := range clients {
			sendChatList(client)
		}
	}
}

type SendRequest struct {
	Type string
}

func sendRequest(client *websocket.Conn, requestType string) {
	var err error

	request := &SendRequest{
		Type: requestType,
	}
	message := &[]byte{}
	*message, err = json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println(err)
		return
	}
}

var registered = false

func RegistrationDone() {
	registered = true
	for client := range clients {
		sendRequest(client, "registrationDone")
	}
}
func RequestInput(request string) string {
	fmt.Println(request)
	for client := range clients {
		sendRequest(client, request)
	}
	requestChannel = make(chan string)
	text := <-requestChannel
	requestChannel = nil
	return text
}
