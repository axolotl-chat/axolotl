package webserver

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
	"unsafe"

	log "github.com/sirupsen/logrus"
	"github.com/vincent-petithory/dataurl"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/store"
)

var ( // TODO: WIP 831
	registered                 = false
	requestPassword            = false
	requestSmsVerificationCode = false
	requestUsername            = false
)

type MessageRecieved struct {
	MessageRecieved *store.Message
}

func (w *WsApp) MessageHandler(msg *store.Message) {
	messageRecieved := &MessageRecieved{
		MessageRecieved: msg,
	}
	// fetch attached message
	if msg.Flags == helpers.MsgFlagQuote {
		if msg.QuoteID != -1 {
			qm, err := store.GetMessageById(msg.QuoteID)
			if err != nil {
				log.Errorln("[axolotl] Quoted Message not found ", err)
			} else {
				msg.QuotedMessage = qm
			}
		}
	}
	var err error
	message := &[]byte{}
	*message, err = json.Marshal(messageRecieved)
	if err != nil {
		log.Errorln("[axolotl-ws] messageHandler", err)
		return
	}
	w.Broadcast <- *message
	w.UpdateChatList()
}

type UpdateMessage struct {
	UpdateMessage *store.Message
}

// UpdateMessageHandler sents message receipts to all connected clients for the activeChat
func (w *WsApp) UpdateMessageHandler(msg *store.Message) {
	if msg.SID == w.ActiveChat {
		log.Debugln("[axolotl-ws] UpdateMessageHandler ", msg.SentAt)
		updateMessage := &UpdateMessage{
			UpdateMessage: msg,
		}
		var err error
		message := &[]byte{}
		*message, err = json.Marshal(updateMessage)
		if err != nil {
			log.Errorln("[axolotl-ws] UpdateMessageHandler", err)
			return
		}
		w.Broadcast <- *message
		w.UpdateChatList()
	}
}

// UpdateMessageHandlerWithSource checks if the message belongs to the current chat and if yes
// triggers an update on axolotl web
func (w *WsApp) UpdateMessageHandlerWithSource(msg *store.Message) {
	if msg.SID == w.ActiveChat {
		log.Debugln("[axolotl-ws] UpdateMessageHandlerWithSource ", msg.SID, msg.SentAt)
		updateMessage := &UpdateMessage{
			UpdateMessage: msg,
		}
		var err error
		message := &[]byte{}
		*message, err = json.Marshal(updateMessage)
		if err != nil {
			log.Errorln("[axolotl-ws] UpdateMessageHandlerWithSource", err)
			return
		}
		w.Broadcast <- *message
		w.UpdateChatList()
	}

}

type SendRequest struct {
	Type string
}

func (w *WsApp) sendRequest(requestType string) {
	var err error
	// mu.Lock()
	// defer mu.Unlock()
	request := &SendRequest{
		Type: requestType,
	}
	log.Debugln("[axolotl-ws] send request", requestType)
	message := &[]byte{}
	*message, err = json.Marshal(request)
	if err != nil {
		log.Errorln("[axolotl-ws] SendRequest", err)
		return
	}
	w.Broadcast <- *message
}

// RegistrationDone sets restration status to done and sends registration status to axoltol-web
func (w *WsApp) RegistrationDone() {
	registered = true
	w.sendRequest("registrationDone")
}

type SendEnterChatRequest struct {
	Type string
	Chat int64
}

func (w *WsApp) requestEnterChat(chat int64) {
	var err error
	request := &SendEnterChatRequest{
		Type: "requestEnterChat",
		Chat: chat,
	}
	message := &[]byte{}
	*message, err = json.Marshal(request)
	if err != nil {
		log.Errorln("[axolotl] requestEnterChat", err)
		return
	}
	w.ActiveChat = chat
	w.Broadcast <- *message
}

func (w *WsApp) RequestInput(request string) string {
	if request == "getEncryptionPw" {
		requestPassword = true
	} else if request == "getUsername" {
		requestUsername = true
	}
	w.sendRequest(request)
	w.RequestChannel = make(chan string)
	text := <-w.RequestChannel
	w.RequestChannel = nil
	return text
}
func (w *WsApp) sendError(client *websocket.Conn, errorMessage string) {
	var err error

	request := &SendError{
		Type:  "Error",
		Error: errorMessage,
	}
	message := &[]byte{}
	*message, err = json.Marshal(request)
	if err != nil {
		log.Errorln("[axolotl] sendError", err)
		return
	}
	w.Broadcast <- *message
}

type SendError struct {
	Type  string
	Error string
}

func (w *WsApp) ShowError(errorMessage string) {
	for client := range w.Clients {
		w.sendError(client, errorMessage)
	}
}
func (w *WsApp) ClearError() {
	for client := range w.Clients {
		w.sendError(client, "")
	}
}

func (w *WsApp) sendAttachment(attachment SendAttachmentMessage) error {
	log.Infoln("[axolotl] send attachment ")
	file := strings.TrimPrefix(attachment.Path, "file://")
	fi, err := os.Stat(file)
	if err != nil {
		log.Errorln("[axolotl] attachment error:", err)
		return err
	}
	if fi.Size() > config.MaxAttachmentSize {
		log.Errorln("[axolotl] attachment error: Attachment too large, not sending")
		return nil
	}
	m, err := sender.SendMessageHelper(attachment.To, attachment.Message, file, nil, false)
	if err == nil {
		go w.MessageHandler(m)
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
func (w *WsApp) uploadSendAttachment(attachment UploadAttachmentMessage) error {
	log.Debug("[axolotl] uploadSendAttachment to ", attachment.To)
	attachDir := config.GetAttachDir()

	file := attachDir + "/" + RandStringBytesMaskImprSrcUnsafe(10)
	dataURL, err := dataurl.DecodeString(attachment.Attachment)
	if err != nil {
		log.Errorln("[axolotl] uploadSendAttachment", err)

	}
	ioutil.WriteFile(file, dataURL.Data, 0644)

	fi, err := os.Stat(file)
	if err != nil {
		log.Errorln("[axolotl] attachment error:", err)
		return err
	}
	if fi.Size() > config.MaxAttachmentSize {
		log.Errorln("[axolotl] attachment error: Attachment too large, not sending")
		return nil
	}
	m, err := sender.SendMessageHelper(attachment.To, attachment.Message, file, nil, false)
	if err == nil {
		go w.MessageHandler(m)
	}
	return nil
}
func (w *WsApp) uploadSendVoiceNote(voiceNote SendVoiceNoteMessage) error {
	log.Debug("[axolotl] uploadSendVoiceNote to ", voiceNote.To)
	attachDir := config.GetAttachDir()
	file := attachDir + "/" + RandStringBytesMaskImprSrcUnsafe(10) + ".mp3"
	dataURL, err := dataurl.DecodeString(voiceNote.VoiceNote)
	if err != nil {
		log.Errorln("[axolotl] voiceNote error:", err)
	}
	ioutil.WriteFile(file, dataURL.Data, 0644)

	fi, err := os.Stat(file)
	if err != nil {
		log.Errorln("[axolotl] voiceNote error:", err)
		return err
	}
	if fi.Size() > config.MaxAttachmentSize {
		log.Errorln("[axolotl] voiceNote error: Attachment too large, not sending")
		return nil
	}
	m, err := sender.SendMessageHelper(voiceNote.To, "", file, nil, true)
	if err == nil {
		go w.MessageHandler(m)
	}
	return nil
}
