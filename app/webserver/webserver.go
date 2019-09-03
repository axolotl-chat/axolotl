package webserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/textsecure-qml/app/sender"
	"github.com/nanu-c/textsecure-qml/app/store"
)

var clients = make(map[*websocket.Conn]bool)
var activeChat = ""

type MessageListEnvelope struct {
	MessageList *store.MessageList
}
type ChatListEnvelope struct {
	ChatList []*store.Session
}

func Run() error {
	// cmd := exec.Command("qmlscene", "qml/Main.qml")
	log.Printf("Starting Axolotl-gui")
	go webserver()
	go sync()
	// stdout, _ := cmd.StdoutPipe()
	// err := cmd.Run()
	// go print(stdout)
	// log.Printf("Axolotl-gui finished with error: %v", err)

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
	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
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
		}
		if incomingMessage.Type == "getMessageList" {
			getMessageListMessage := GetMessageListMessage{}
			json.Unmarshal([]byte(p), &getMessageListMessage)
			id := getMessageListMessage.ID
			activeChat = getMessageListMessage.ID
			sendMessageList(conn, id)
		}
		if incomingMessage.Type == "sendMessage" {
			sendMessageMessage := SendMessageMessage{}
			json.Unmarshal([]byte(p), &sendMessageMessage)
			sender.SendMessageHelper(sendMessageMessage.To, sendMessageMessage.Message, "")
			go UpdateChatList()
		}
		if incomingMessage.Type == "requestCode" {
			if requestChannel != nil {
				requestCodeMessage := RequestCodeMessage{}
				json.Unmarshal([]byte(p), &requestCodeMessage)
				requestChannel <- requestCodeMessage.Tel
			}
		}
		if incomingMessage.Type == "sendCode" {
			if requestChannel != nil {
				sendCodeMessage := SendCodeMessage{}
				json.Unmarshal([]byte(p), &sendCodeMessage)
				requestChannel <- sendCodeMessage.Code
			}
			// sender.SendMessageHelper(sendMessageMessage.To, sendMessageMessage.Message, "")
		}
		if incomingMessage.Type == "getRegistrationStatus" {
			if registered {
				sendRequest(conn, "registrationDone")
			} else {
				sendRequest(conn, "getPhoneNumber")
			}

		}

	}
}
func webserver() {
	http.Handle("/", http.FileServer(http.Dir("./axolotl-web")))
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

var test = false

func UpdateChatList() {
	// fmt.Println("updateChatList")
	// rq := make(chan string)
	for client := range clients {
		sendChatList(client)
		// if !test {
		// 	RequestInput("getPhoneNumber", rq)
		// } else {
		// 	RequestInput("getVerificationCode", rq)
		// }
	}
	if activeChat != "" {
		for client := range clients {
			sendMessageList(client, activeChat)
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
