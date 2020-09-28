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
		text = "Secure session reset."
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
	if msg.Quote() != nil {
		msgFlags = helpers.MsgFlagQuote
		text = ">" + msg.Quote().GetText() + `

		` + msg.Message()
	}
	session := store.SessionsModel.Get(s)
	m := session.Add(text, msg.Source(), f, mt, false, store.ActiveSessionID)
	m.ReceivedAt = uint64(time.Now().UnixNano() / 1000000)
	m.SentAt = msg.Timestamp()
	session.ExpireTimer = msg.ExpireTimer()
	store.UpdateSession(session)
	m.ExpireTimer = msg.ExpireTimer()
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	session.Timestamp = m.SentAt
	session.When = m.HTime
	if gr != nil && gr.Flags == textsecure.GroupUpdateFlag {
		session.Name = gr.Name
	}
	if msgFlags != 0 {
		m.Flags = msgFlags
	}
	//TODO: have only one message per chat
	if session.Notification {
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
func ReceiptHandler(source string, devID uint32, timestamp uint64) {
	log.Println("[axolotl] receiptHandler for message ", timestamp)
	webserver.UpdateChatList()

	s := store.SessionsModel.Get(source)
	for i := len(s.Messages) - 1; i >= 0; i-- {
		m := s.Messages[i]
		if m.SentAt == timestamp {
			m.IsSent = true
			store.UpdateMessageReceiptSent(m)
			webserver.UpdateActiveChat()
			return
		}
	}
	webserver.UpdateChatList()
	log.Printf("[axolotl] receipt: Message with timestamp %d not found\n", timestamp)
}

func ReceiptMessageHandler(msg *textsecure.Message) {
	log.Println("[axolotl] receiptMessageHandler for message ", msg.Timestamp())

	webserver.UpdateChatList()
	s := store.SessionsModel.Get(msg.Source())
	for i := len(s.Messages) - 1; i >= 0; i-- {
		m := s.Messages[i]
		if m.SentAt == msg.Timestamp() {
			if msg.Message() == "readReceiptMessage" {
				m.IsRead = true
				store.UpdateMessageRead(m)
			} else {
				log.Debugln("[axolotl] unhandeld receipt message type for message ", msg.Timestamp(), msg.Message())
			}
			webserver.UpdateActiveChat()
			return
		}
	}
	webserver.UpdateChatList()
	log.Printf("[axolotl] receipt: Message with timestamp %d not found\n", msg.Timestamp())
	log.Println("[axolotl] receiptMessageHandler: Message ", msg)
}

func SyncSentHandler(msg *textsecure.Message, timestamp uint64) {
	var err error

	var f []store.Attachment
	mt := ""
	if len(msg.Attachments()) > 0 {
		for i, a := range msg.Attachments() {
			mt = msg.Attachments()[i].MimeType
			file, err := store.SaveAttachment(a)
			if err != nil {
				log.Printf("[Axolotl] Error saving %s\n", err.Error())
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
	//Group Message
	gr := msg.Group()

	if gr != nil && gr.Flags != 0 {
		_, ok := store.Groups[gr.Hexid]
		members := ""
		if ok {
			members = store.Groups[gr.Hexid].Members
		}
		av := []byte{}

		if gr.Avatar != nil {
			av, err = ioutil.ReadAll(bytes.NewReader(gr.Avatar))
			if err != nil {
				log.Println(err)
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
	if msg.Sticker() != nil {
		msgFlags = helpers.MsgFlagSticker
		text = "Unsupported Message: sticker"
	}
	if msg.Contact() != nil {
		msgFlags = helpers.MsgFlagContact
		c := msg.Contact()
		text = c[0].String()
	}
	s := msg.Source()
	if gr != nil {
		s = gr.Hexid
	}
	session := store.SessionsModel.Get(s)
	m := session.Add(text, "", f, mt, true, store.ActiveSessionID)
	m.ReceivedAt = uint64(time.Now().UnixNano() / 1000000)
	m.SentAt = msg.Timestamp()
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	session.Timestamp = m.SentAt
	session.When = m.HTime
	if gr != nil && gr.Flags == textsecure.GroupUpdateFlag {
		session.Name = gr.Name
	}
	if msgFlags != 0 {
		m.Flags = msgFlags
		// m.StatusMessage = true
	}
	if len(text) == 0 {
		// m.StatusMessage = true
	}
	m.IsSent = true
	//TODO: have only one message per chat
	// if session.Notification {
	// 	if settings.SettingsModel.EncryptDatabase{
	// 		text = "Encrypted message"
	// 	}
	// 	n := Nh.NewStandardPushMessage(
	// 		session.Name,
	// 		text, "")
	// 	Nh.Send(n)
	// }

	store.SaveMessage(m)
	store.UpdateSession(session)
}
