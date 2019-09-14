package handler

import (
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/lang"
	"github.com/nanu-c/textsecure-qml/app/push"
	"github.com/nanu-c/textsecure-qml/app/settings"
	"github.com/nanu-c/textsecure-qml/app/store"
	"github.com/nanu-c/textsecure-qml/app/webserver"
)

//messageHandler is used on incoming message
func MessageHandler(msg *textsecure.Message) {
	var err error
	f := ""
	mt := ""
	if len(msg.Attachments()) > 0 {
		mt = msg.Attachments()[0].MimeType
		f, err = store.SaveAttachment(msg.Attachments()[0])
		if err != nil {
			log.Printf("Error saving %s\n", err.Error())
		}
	}

	msgFlags := 0

	text := msg.Message()
	if msg.Flags() == textsecure.EndSessionFlag {
		text = lang.SessionReset
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
			av, err = ioutil.ReadAll(gr.Avatar)
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

	s := msg.Source()
	if gr != nil {
		s = gr.Hexid
	}
	session := store.SessionsModel.Get(s)
	m := session.Add(text, msg.Source(), f, mt, false, store.ActiveSessionID)
	m.ReceivedAt = uint64(time.Now().UnixNano() / 1000000)
	m.SentAt = msg.Timestamp()
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	//qml.Changed(m, &m.HTime)
	session.Timestamp = m.SentAt
	session.When = m.HTime
	//qml.Changed(session, &session.When)
	if gr != nil && gr.Flags == textsecure.GroupUpdateFlag {
		session.Name = gr.Name
		//qml.Changed(session, &session.Name)
	}

	if msgFlags != 0 {
		m.Flags = msgFlags
		//qml.Changed(m, &m.Flags)
	}
	//TODO: have only one message per chat
	if session.Notification {
		if settings.SettingsModel.EncryptDatabase {
			text = "Encrypted message"
		}
		//only send a notification, when it's not the current chat
		// if session.ID != store.Sessions.GetActiveChat {
		if true {
			n := push.Nh.NewStandardPushMessage(
				session.Name,
				text, "")
			push.Nh.Send(n)
		}
	}
	err, msgSend := store.SaveMessage(m)
	if err != nil {
		log.Printf("Error saving %s\n", err.Error())
	}
	store.UpdateSession(session)
	// webserver.UpdateChatList()
	webserver.MessageHandler(msgSend)
}
func ReceiptMessageHandler(msg *textsecure.Message) {
	webserver.UpdateChatList()
	webserver.UpdateChatList()
}
func TypingMessageHandler(msg *textsecure.Message) {
	webserver.UpdateChatList()
}
func ReceiptHandler(source string, devID uint32, timestamp uint64) {
	s := store.SessionsModel.Get(source)
	for i := len(s.Messages) - 1; i >= 0; i-- {
		m := s.Messages[i]
		if m.SentAt == timestamp {
			m.IsRead = true
			//qml.Changed(m, &m.IsRead)
			store.UpdateMessageRead(m)
			return
		}
	}
	log.Printf("Message with timestamp %d not found\n", timestamp)
}

func SyncSentHandler(msg *textsecure.Message, timestamp uint64) {
	var err error

	f := ""
	mt := ""
	if len(msg.Attachments()) > 0 {
		mt = msg.Attachments()[0].MimeType
		f, err = store.SaveAttachment(msg.Attachments()[0])
		if err != nil {
			log.Printf("Error saving %s\n", err.Error())
		}
	}

	msgFlags := 0

	text := msg.Message()
	if msg.Flags() == textsecure.EndSessionFlag {
		text = lang.SessionReset
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
			av, err = ioutil.ReadAll(gr.Avatar)
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

	s := msg.Source()
	if gr != nil {
		s = gr.Hexid
	}
	session := store.SessionsModel.Get(s)
	// m := session.Add(text, msg.Source(), f, mt, false, store.ActiveSessionID)
	m := session.Add(text, "", f, mt, true, store.ActiveSessionID)

	m.ReceivedAt = uint64(time.Now().UnixNano() / 1000000)
	m.SentAt = msg.Timestamp()
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	//qml.Changed(m, &m.HTime)
	session.Timestamp = m.SentAt
	session.When = m.HTime
	//qml.Changed(session, &session.When)
	if gr != nil && gr.Flags == textsecure.GroupUpdateFlag {
		session.Name = gr.Name
		//qml.Changed(session, &session.Name)
	}

	if msgFlags != 0 {
		m.Flags = msgFlags
		//qml.Changed(m, &m.Flags)
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
