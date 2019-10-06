package webserver

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/textsecure-qml/app/sender"
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

var requestPassword = false

func RequestInput(request string) string {
	if request == "getEncryptionPw" {
		requestPassword = true
	}
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

func sendAttachment(attachment SendAttachmentMessage) error {
	// log.Infoln("[axolotl] send attachment ", attachment.Path)
	// Do not allow sending attachments larger than 100M for now
	var maxAttachmentSize int64 = 100 * 1024 * 1024
	// log.Printf("SendAttachmentApi")
	file := strings.TrimPrefix(attachment.Path, "file://")
	fi, err := os.Stat(file)
	if err != nil {
		log.Errorln("[axolotl] attachment error:", err)
		return err
	}
	if fi.Size() > maxAttachmentSize {
		log.Errorln("[axolotl] attachment error: Attachment too large, not sending")
		return nil
	}
	err, m := sender.SendMessageHelper(attachment.To, attachment.Message, file)
	if err == nil {
		go MessageHandler(m)
	}
	return nil
}
