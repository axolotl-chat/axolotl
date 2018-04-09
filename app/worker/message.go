package worker

import (
	"io"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/aebruno/textsecure"
	qml "github.com/nanu-c/qml-go"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/store"
)

var (
	msgFlagGroupNew     = 1
	msgFlagGroupUpdate  = 2
	msgFlagGroupLeave   = 4
	msgFlagResetSession = 8
)

func (Api *TextsecureAPI) SendMessage(to, message string) error {
	return SendMessageHelper(to, message, "")
}
func SendMessage(s *store.Session, m *store.Message) {
	var att io.Reader
	var err error

	if m.Attachment != "" {
		att, err = os.Open(m.Attachment)
		if err != nil {
			return
		} else {
			log.Printf("SendMessage FileOpend")
		}
	}

	ts := SendMessageLoop(s.Tel, m.Message, s.IsGroup, att, m.Flags)

	m.SentAt = ts
	s.Timestamp = m.SentAt
	m.IsSent = true
	//FIXME avoid rerendering the whole qml
	qml.Changed(m, &m.IsSent)
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	qml.Changed(m, &m.HTime)
	s.When = m.HTime
	qml.Changed(s, &s.When)
	store.UpdateMessageSent(m)
	store.UpdateSession(s)
}
func SendMessageLoop(to, message string, group bool, att io.Reader, flags int) uint64 {
	var err error
	var ts uint64
	for {
		err = nil
		if flags == msgFlagResetSession {
			ts, err = textsecure.EndSession(to, "TERMINATE")
		} else if flags == msgFlagGroupLeave {
			err = textsecure.LeaveGroup(to)
		} else if flags == msgFlagGroupUpdate {
			_, err = textsecure.UpdateGroup(to, store.Groups[to].Name, strings.Split(store.Groups[to].Members, ","))
		} else if att == nil {
			if group {
				ts, err = textsecure.SendGroupMessage(to, message)
			} else {
				ts, err = textsecure.SendMessage(to, message)
			}
		} else {
			if group {
				ts, err = textsecure.SendGroupAttachment(to, message, att)
			} else {
				log.Printf("SendMessageLoop sendAttachment")
				// buf := new(bytes.Buffer)
				// buf.ReadFrom(att)
				// s := buf.String()
				// log.Printf(s)

				ts, err = textsecure.SendAttachment(to, message, att)
			}
		}
		if err == nil {
			break
		}
		log.Println(err)
		//If sending failed, try again after a while
		time.Sleep(3 * time.Second)
	}
	return ts
}
func SendMessageHelper(to, message, file string) error {
	var err error
	if file != "" {
		file, err = store.CopyAttachment(file)
		// log.Printf("got Attachment:" + file)
		if err != nil {
			log.Printf("Error Attachment:" + err.Error())
			return err
		}
	}
	session := store.SessionsModel.Get(to)
	m := session.Add(message, "", file, "", true, Api.ActiveSessionID)
	store.SaveMessage(m)
	go SendMessage(session, m)
	return nil
}
func SendUnsentMessages() {
	for _, s := range store.SessionsModel.Sess {
		for _, m := range s.Messages {
			if m.Outgoing && !m.IsSent {
				go SendMessage(s, m)
			}
		}
	}
}
func (Api *TextsecureAPI) DeleteMessage(msg *store.Message, tel string) {
	store.DeleteMessage(msg.ID)
	s := store.SessionsModel.Get(tel)
	for i, m := range s.Messages {
		if m.ID == msg.ID {
			s.Messages = append(s.Messages[:i], s.Messages[i+1:]...)
			s.Len--
			qml.Changed(s, &s.Len)
			return
		}
	}
}
