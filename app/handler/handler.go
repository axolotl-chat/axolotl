package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gen2brain/beeep"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/settings"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/nanu-c/axolotl/app/webserver"
	"github.com/signal-golang/textsecure"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

//messageHandler is used on incoming message
func MessageHandler(msg *textsecure.Message) {
	buildAndSaveMessage(msg, false)
}
func buildAndSaveMessage(msg *textsecure.Message, syncMessage bool) {
	var err error
	var f []store.Attachment //should be array
	mt := ""                 //
	if len(msg.Attachments()) > 0 {
		for i, a := range msg.Attachments() {
			mt = msg.Attachments()[i].MimeType
			file, err := store.SaveAttachment(a)
			if err != nil {
				log.Printf("[axolotl] MessageHandler Error saving attachments %s\n", err.Error())
			}
			f = append(f, file)
		}
	}

	msgFlags := 0

	text := msg.Message()
	if msg.Flags() == textsecure.EndSessionFlag {
		text = "Secure session reset."
		msgFlags = helpers.MsgFlagResetSession
	}
	if msg.Flags() == 2 {
		text = "Message timer update."
		msgFlags = helpers.MsgFlagExpirationTimerUpdate
	}
	//Group Message
	gr := msg.Group()

	if gr != nil && gr.Flags != 0 || gr != nil && gr.Name != store.Groups[gr.Hexid].Name {
		_, ok := store.Groups[gr.Hexid]
		members := ""
		if ok {
			members = store.Groups[gr.Hexid].Members
			if store.Groups[gr.Hexid].Name == gr.Hexid {
				textsecure.RemoveGroupKey(gr.Hexid)
				textsecure.RequestGroupInfo(gr)
			}
		}
		av := []byte{}

		if gr.Avatar != nil {
			av, err = ioutil.ReadAll(bytes.NewReader(gr.Avatar))
			if err != nil {
				log.Println("[axolotl]", err)
				return
			}
		}
		store.Groups[gr.Hexid] = &store.GroupRecord{
			GroupID: gr.Hexid,
			Members: strings.Join(gr.Members, ","),
			Name:    gr.Name,
			Avatar:  av,
			Active:  true,
		}
		if ok {
			store.UpdateGroup(store.Groups[gr.Hexid])
		} else {
			store.SaveGroup(store.Groups[gr.Hexid])
		}

		if gr.Flags == textsecure.GroupUpdateFlag {
			dm, _ := helpers.MembersDiffAndUnion(members, strings.Join(gr.Members, ","))
			text = store.GroupUpdateMsg(dm, gr.Name)
			msgFlags = helpers.MsgFlagGroupUpdate
		}
		if gr.Flags == textsecure.GroupLeaveFlag {
			text = store.TelToName(msg.Source()) + " has left the group."
			msgFlags = helpers.MsgFlagGroupLeave
		}
	}

	s := msg.Source()
	if gr != nil {
		s = gr.Hexid
	}
	if msg.Sticker() != nil {
		msgFlags = helpers.MsgFlagSticker
		text = "Unsupported Message: sticker"
	}
	if msg.Contact() != nil {
		msgFlags = helpers.MsgFlagContact
		c := msg.Contact()[0]
		text = c.GetName().GetDisplayName() + " " + c.GetNumber()[0].GetValue()
	}
	if msg.Reaction() != nil {
		msgFlags = helpers.MsgFlagReaction
		text = msg.Reaction().GetEmoji()
	}
	session := store.SessionsModel.Get(s)
	var m *store.Message
	if syncMessage {
		m = session.Add(text, "", f, mt, true, store.ActiveSessionID)
		m.IsSent = true
	} else {
		m = session.Add(text, msg.Source(), f, mt, false, store.ActiveSessionID)
	}
	m.ReceivedAt = uint64(time.Now().UnixNano() / 1000000)
	m.SentAt = msg.Timestamp()
	session.ExpireTimer = msg.ExpireTimer()
	store.UpdateSession(session)
	m.ExpireTimer = msg.ExpireTimer()
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	if msg.Quote() != nil {
		msgFlags = helpers.MsgFlagQuote
		text = msg.Message()
		err, id := store.FindQuotedMessage(msg.Quote())
		if err != nil || id == -1 {
			// create quoted message
			quoteMessage := session.Add(text, msg.Quote().GetAuthorE164(), nil, msg.Quote().GetText(), false, store.ActiveSessionID)
			quoteMessage.Flags = helpers.MsgFlagHiddenQuote
			err, savedQuoteMessage := store.SaveMessage(quoteMessage)
			id = savedQuoteMessage.ID
			if err != nil {
				log.Debugln("[axolotl] Error saving quote message")
			}
		}
		m.QuoteID = id
	}
	session.Timestamp = m.SentAt
	session.When = m.HTime
	if gr != nil && gr.Flags == textsecure.GroupUpdateFlag {
		session.Name = gr.Name
	}
	if msgFlags != 0 {
		m.Flags = msgFlags
	}
	//TODO: have only one message per chat

	if session.Notification && !syncMessage {
		if settings.SettingsModel.EncryptDatabase {
			text = "Encrypted message"
		}
		//only send a notification, when it's not the current chat
		// if session.ID != store.Sessions.GetActiveChat {
		if s != store.ActiveSessionID {
			if config.Gui == "ut" {
				n := push.Nh.NewStandardPushMessage(
					session.Name,
					text, "", s)
				push.Nh.Send(n)
			} else {
				err := beeep.Notify("Axolotl: "+session.Name, text, "axolotl-web/dist/public/axolotl.png")
				if err != nil {
					log.Errorln("[axolotl] notification ", err)
				}
			}
		}
	}
	err, msgSend := store.SaveMessage(m)
	if err != nil {
		log.Printf("[axolotl] MessageHandler: Error saving message: %s\n", err.Error())
	}
	store.UpdateSession(session)
	// webserver.UpdateChatList()
	webserver.MessageHandler(msgSend)
}
func CallMessageHandler(msg *textsecure.Message) {
	log.Debugln("[axolotl] CallMessageHandler", msg)
	session := store.SessionsModel.Get(msg.Source())
	var f []store.Attachment
	m := session.Add(msg.Message(), "", f, "", true, store.ActiveSessionID)
	store.SaveMessage(m)
	webserver.UpdateChatList()
	webserver.UpdateChatList()
}
func TypingMessageHandler(msg *textsecure.Message) {
	webserver.UpdateChatList()
}

// ReceiptHandler handles receipts for outgoing messages
func ReceiptHandler(source string, devID uint32, timestamp uint64) {
	m, err := store.FindOutgoingMessage(timestamp)
	if err != nil {
		log.Printf("[axolotl] ReceiptHandler: Message with timestamp %d not found\n", timestamp)
	} else {
		log.Println("[axolotl] ReceiptHandler for message ", timestamp, m.Source)
		m.IsSent = true
		m.Receipt = true
		store.UpdateMessageReceipt(m)
		webserver.UpdateMessageHandlerWithSource(m, m.Source)
		return
	}
}

// ReceiptMessageHandler handles outgoing message receipts and marks messages as read
func ReceiptMessageHandler(msg *textsecure.Message) {
	log.Println("[axolotl] ReceiptMessageHandler for message ", msg.Timestamp())
	m, err := store.FindOutgoingMessage(msg.Timestamp())
	if err != nil {
		log.Printf("[axolotl] ReceiptMessageHandler: Message with timestamp %d not found\n", msg.Timestamp())
		return
	} else {
		if msg.Message() == "readReceiptMessage" {
			m.IsRead = true
			m.IsSent = true
			store.UpdateMessageRead(m)
		} else if msg.Message() == "deliveryReceiptMessage" {
			log.Debugln("[axolotl] unhandeld receipt message type for message ", msg.Timestamp(), msg.Message())
			m.IsSent = true
			store.UpdateMessageReceiptSent(m)
		}
		webserver.UpdateMessageHandlerWithSource(m, m.Source)
		return
	}
}

// SyncSentHandler handle sync messages from signal desktop
func SyncSentHandler(msg *textsecure.Message, timestamp uint64) {
	log.Debugln("[axolotl] handle sync message", msg.Timestamp())
	// use the same routine to save sync messages as incoming messages
	buildAndSaveMessage(msg, true)
}
