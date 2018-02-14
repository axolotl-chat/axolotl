package store

import (
	"os"
	"time"

	qml "github.com/nanu-c/qml-go"
	"github.com/nanu-c/textsecure-qml/app/helpers"
)

type Session struct {
	ID        int64
	Name      string
	Tel       string
	IsGroup   bool
	Last      string
	Timestamp uint64
	When      string
	CType     int
	Messages  []*Message
	Unread    int
	Active    bool
	Len       int
}
type Sessions struct {
	Sess []*Session
	Len  int
}

//TODO that hasn't to  be in the db controller
var AllSessions []*Session

var SessionsModel = &Sessions{
	Sess: make([]*Session, 0),
}

func SaveSession(s *Session) error {
	res, err := DS.Dbx.NamedExec(sessionsInsert, s)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	s.ID = id
	return err
}
func UpdateSession(s *Session) error {
	_, err := DS.Dbx.NamedExec("UPDATE sessions SET name = :name, timestamp = :timestamp, ctype = :ctype, last = :last, unread = :unread WHERE id = :id", s)
	if err != nil {
		return err
	}
	return err
}
func DeleteSession(tel string) error {
	s := SessionsModel.Get(tel)
	_, err := DS.Dbx.Exec("DELETE FROM messages WHERE sid = ?", s.ID)
	if err != nil {
		return err
	}
	_, err = DS.Dbx.Exec("DELETE FROM sessions WHERE id = ?", s.ID)
	if err != nil {
		return err
	}
	index := SessionsModel.GetIndex(s.Tel)
	SessionsModel.Sess = append(SessionsModel.Sess[:index], SessionsModel.Sess[index+1:]...)
	SessionsModel.Len--
	qml.Changed(SessionsModel, &SessionsModel.Len)
	return nil
}
func (s *Sessions) GetSession(i int) *Session {
	return s.Sess[i]
}

func (s *Session) Add(text string, source string, file string, mimetype string, outgoing bool, sessionID string) *Message {

	ctype := helpers.ContentTypeMessage
	if file != "" {
		f, _ := os.Open(file)
		ctype = helpers.ContentType(f, mimetype)
	}
	message := &Message{Message: text,
		SID:        s.ID,
		Outgoing:   outgoing,
		Source:     source,
		CType:      ctype,
		Attachment: file,
		HTime:      "Now",
		SentAt:     uint64(time.Now().UnixNano() / 1000000),
	}
	s.Messages = append(s.Messages, message)
	s.Last = text
	s.Len++
	s.CType = ctype
	qml.Changed(s, &s.Last)
	qml.Changed(s, &s.Len)
	qml.Changed(s, &s.CType)
	//FIXME
	if !outgoing && sessionID != s.Tel {
		s.Unread++
		qml.Changed(s, &s.Unread)
	}
	UpdateSession(s)

	s.moveToTop()
	return message
}
func (s *Session) MarkRead() {
	s.Unread = 0
	qml.Changed(s, &s.Unread)
	UpdateSession(s)
}

// updateTimestamps keeps the timestamps of the last message of each session
// updated in human readable form.
// FIXME: make this lazier, to only update timestamps the user sees at the moment
func UpdateTimestamps() {
	for {
		time.Sleep(1 * time.Minute)
		for _, s := range SessionsModel.Sess {
			if s.Len == 0 {
				continue
			}
			for _, m := range s.Messages {
				m.HTime = helpers.HumanizeTimestamp(m.SentAt)
				qml.Changed(m, &m.HTime)
			}
			s.When = s.Messages[len(s.Messages)-1].HTime
			qml.Changed(s, &s.When)
		}
	}
}
func (s *Sessions) Get(tel string) *Session {
	for _, ses := range s.Sess {
		if ses.Tel == tel {
			return ses
		}
	}
	ses := &Session{Tel: tel, Name: TelToName(tel), Active: true, IsGroup: tel[0] != '+'}
	s.Sess = append(s.Sess, ses)
	s.Len++
	qml.Changed(s, &s.Len)
	SaveSession(ses)
	return ses
}

func (s *Sessions) GetIndex(tel string) int {
	for i, ses := range s.Sess {
		if ses.Tel == tel {
			return i
		}
	}
	return -1
}

var topSession string

func (s *Session) moveToTop() {
	if topSession == s.Tel {
		return
	}

	index := SessionsModel.GetIndex(s.Tel)
	SessionsModel.Sess = append([]*Session{s}, append(SessionsModel.Sess[:index], SessionsModel.Sess[index+1:]...)...)

	// force a length change update
	SessionsModel.Len--
	qml.Changed(SessionsModel, &SessionsModel.Len)
	SessionsModel.Len++
	qml.Changed(SessionsModel, &SessionsModel.Len)

	topSession = s.Tel
}
