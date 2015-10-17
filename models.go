package main

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/janimo/textsecure"
	"github.com/janimo/textsecure/vendor/magic"
	"gopkg.in/qml.v1"
)

// Model for the contacts

type Contacts struct {
	contacts []textsecure.Contact
	Len      int
}

func (c *Contacts) Contact(i int) textsecure.Contact {
	if i == -1 {
		return textsecure.Contact{}
	}
	return c.contacts[i]
}

//HACK
func telToName(tel string) string {
	if g, ok := groups[tel]; ok {
		return g.Name
	}
	for _, c := range contactsModel.contacts {
		if c.Tel == tel {
			return c.Name
		}
	}
	if tel == config.Tel {
		return "Me"
	}
	return tel
}

var contactsModel *Contacts = &Contacts{}

func refreshContacts() {
	c, err := textsecure.GetRegisteredContacts()
	if err != nil {
		showError(err)
	}

	contactsModel.contacts = c
	contactsModel.Len = len(c)
	qml.Changed(contactsModel, &contactsModel.Len)
}

func getContactForTel(tel string) *textsecure.Contact {
	for _, c := range contactsModel.contacts {
		if c.Tel == tel {
			return &c
		}
	}
	return nil
}

func (api *textsecureAPI) FilterContacts(sub string) {
	sub = strings.ToUpper(sub)

	fc := []textsecure.Contact{}
	for _, c := range contactsModel.contacts {
		if strings.Contains(strings.ToUpper(telToName(c.Tel)), sub) {
			fc = append(fc, c)
		}
	}

	cm := &Contacts{fc, len(fc)}
	engine.Context().SetVar("contactsModel", cm)
}

// Model for application settings

type Setting struct {
	SendByEnter bool
}

var settingsModel = &Setting{
	SendByEnter: true,
}

// Model for existing chat sessions

type Message struct {
	ID         int64
	SID        int64
	From       string `db:"source"`
	Text       string `db:"message"`
	Outgoing   bool
	SentAt     uint64
	ReceivedAt uint64
	HTime      string
	CType      int
	Attachment string
	IsSent     bool
	IsRead     bool
}

func (m *Message) Info() string {
	timeFormat := time.RFC1123Z
	s := "Sent: " + time.Unix(int64(m.SentAt/1000), 0).Format(timeFormat)
	if m.ReceivedAt != 0 {
		s += "\nReceived: " + time.Unix(int64(m.ReceivedAt/1000), 0).Format(timeFormat)
	}
	return s
}

func (m *Message) Name() string {
	return telToName(m.From)
}

type Session struct {
	ID        int64
	Name      string
	Tel       string
	IsGroup   bool
	Last      string
	Timestamp uint64
	When      string
	CType     int
	messages  []*Message
	Len       int
}

type Sessions struct {
	sessions []*Session
	Len      int
}

func (s *Sessions) Session(i int) *Session {
	return s.sessions[i]
}

func (s *Sessions) Get(tel string) *Session {
	for _, ses := range s.sessions {
		if ses.Tel == tel {
			return ses
		}
	}
	ses := &Session{Tel: tel, Name: telToName(tel), IsGroup: tel[0] != '+'}
	s.sessions = append(s.sessions, ses)
	s.Len++
	qml.Changed(s, &s.Len)
	saveSession(ses)
	return ses
}

var sessionsModel = &Sessions{
	sessions: make([]*Session, 0),
}

func (s *Session) Messages(i int) *Message {
	//FIXME when is index -1 ?
	if i == -1 || i >= len(s.messages) {
		return &Message{}
	}
	return s.messages[i]
}

type GroupRecord struct {
	ID      int64
	GroupID string
	Name    string
	Members string
	Avatar  []byte
}

func (api *textsecureAPI) FilterSessions(sub string) {
	sub = strings.ToUpper(sub)

	sm := &Sessions{
		sessions: make([]*Session, 0),
	}

	for _, s := range sessionsModel.sessions {
		if strings.Contains(strings.ToUpper(telToName(s.Tel)), sub) {
			sm.sessions = append(sm.sessions, s)
			sm.Len++
		}
	}

	engine.Context().SetVar("sessionsModel", sm)
}

//Mirror the Ubuntu.Content QML library constants
//type ContentType int

const (
	ContentTypeMessage int = iota
	ContentTypeDocuments
	ContentTypePictures
	ContentTypeMusic
	ContentTypeContacts
	ContentTypeVideos
	ContentTypeLinks
)

// contentType gets the content type (message, picture, video) of an attachment by sniffing its MIME type
func contentType(att io.Reader) (int, io.Reader) {
	ct := ContentTypeMessage
	if att == nil {
		return ct, nil
	}
	mt, r := magic.MIMETypeFromReader(att)
	if strings.HasPrefix(mt, "image") {
		ct = ContentTypePictures
	}
	if strings.HasPrefix(mt, "video") {
		ct = ContentTypeVideos
	}
	if strings.HasPrefix(mt, "audio") {
		ct = ContentTypeMusic
	}
	return ct, r
}

func (s *Session) Add(text string, from string, file string, outgoing bool) *Message {
	ctype := ContentTypeMessage
	if file != "" {
		f, _ := os.Open(file)
		ctype, _ = contentType(f)
	}
	message := &Message{Text: text,
		SID:        s.ID,
		Outgoing:   outgoing,
		From:       from,
		CType:      ctype,
		Attachment: file,
		HTime:      "Now",
	}
	s.messages = append(s.messages, message)
	s.Last = text
	s.Len++
	s.CType = ctype
	qml.Changed(s, &s.Last)
	qml.Changed(s, &s.Len)
	qml.Changed(s, &s.CType)
	updateSession(s)
	return message
}

// updateTimestamps keeps the timestamps of the last message of each session
// updated in human readable form.
// FIXME: make this lazier, to only update timestamps the user sees at the moment
func updateTimestamps() {
	for {
		time.Sleep(1 * time.Minute)
		for _, s := range sessionsModel.sessions {
			if s.Len == 0 {
				continue
			}
			for _, m := range s.messages {
				m.HTime = humanizeTimestamp(m.SentAt)
				qml.Changed(m, &m.HTime)
			}
			s.When = s.messages[len(s.messages)-1].HTime
			qml.Changed(s, &s.When)
		}
	}
}

// initModels exports the Go models to QML
func initModels() {
	engine.Context().SetVar("contactsModel", contactsModel)
	engine.Context().SetVar("settingsModel", settingsModel)
	engine.Context().SetVar("sessionsModel", sessionsModel)

	go updateTimestamps()
}
