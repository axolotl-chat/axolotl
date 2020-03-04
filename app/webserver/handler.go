package webserver

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/vincent-petithory/dataurl"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/store"
)

var mu sync.Mutex

type MessageRecieved struct {
	MessageRecieved *store.Message
}

func MessageHandler(msg *store.Message) {
	messageRecieved := &MessageRecieved{
		MessageRecieved: msg,
	}
	// mu.Lock()
	// defer mu.Unlock()
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
	// mu.Lock()
	// defer mu.Unlock()
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
	// mu.Lock()
	// defer mu.Unlock()
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
	// mu.Lock()
	// defer mu.Unlock()
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

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
func uploadSendAttachment(attachment UploadAttachmentMessage) error {
	// log.Infoln("[axolotl] send attachment ", attachment.Path)
	// Do not allow sending attachments larger than 100M for now
	var maxAttachmentSize int64 = 100 * 1024 * 1024
	file := config.AttachDir + "/" + RandStringBytesMaskImprSrcUnsafe(10)
	dataURL, err := dataurl.DecodeString(attachment.Attachment)
	if err != nil {
		fmt.Println(err)
	}
	ioutil.WriteFile(file, dataURL.Data, 0644)

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
