package webserver

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/textsecure-qml/app/store"
)

type MessageRecieved struct {
	MessageRecieved *store.Message
}

func MessageHandler(msg *store.Message) {
	messageRecieved := &MessageRecieved{
		MessageRecieved: msg,
	}
	for client := range clients {
		var err error
		message := &[]byte{}
		*message, err = json.Marshal(messageRecieved)
		if err != nil {
			fmt.Println(err)
			return
		}
		if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
			log.Println(err)
			return
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

type SendEnterChatRequest struct {
	Type string
	Chat string
}

func requestEnterChat(chat string) {
	var err error
	for client := range clients {
		request := &SendEnterChatRequest{
			Type: "requestEnterChat",
			Chat: chat,
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
func sendError(client *websocket.Conn, errorMessage string) {
	var err error

	request := &SendError{
		Type:  "Error",
		Error: errorMessage,
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

type SendError struct {
	Type  string
	Error string
}

func ShowError(errorMessage string) {
	for client := range clients {
		sendError(client, errorMessage)
	}
}
