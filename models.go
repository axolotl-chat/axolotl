package main

import (
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
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
	for _, c := range contactsModel.contacts {
		if c.Tel == tel {
			return c.Name
		}
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

// Model for application settings

type Setting struct {
	SendByEnter bool
}

var settingsModel = &Setting{
	SendByEnter: true,
}

// Model for existing chat sessions

type Message struct {
	From     string
	Text     string
	Outgoing bool
	Time     time.Time
	HTime    string
	CType    int
	Att      []byte
}

type Session struct {
	Name     string
	Tel      string
	IsGroup  bool
	Last     string
	When     string
	CType    int
	messages []*Message
	Len      int
}

type Sessions struct {
	m   map[string]*Session
	ind []string
	Len int
}

func (s *Sessions) Session(i int) *Session {
	return s.m[s.ind[i]]
}

func (s *Sessions) Get(tel string) *Session {
	ses, ok := s.m[tel]
	if ok {
		return ses
	}
	// FIXME: better session id/name separation, group ids may need to be exposed from the libraray;
	// for now, consider anything not starting with '+' a group.
	s.m[tel] = &Session{Tel: tel, Name: telToName(tel), IsGroup: tel[0] != '+'}
	s.ind = append(s.ind, tel)
	s.Len++
	qml.Changed(s, &s.Len)
	return s.m[tel]
}

var sessionsModel = &Sessions{
	m:   make(map[string]*Session),
	ind: make([]string, 0),
}

func (s *Session) Message(i int) *Message {
	//FIXME when is index -1 ?
	if i == -1 || i >= len(s.messages) {
		return &Message{}
	}
	return s.messages[i]
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
	return ct, r
}

func (s *Session) Add(text string, from string, att io.Reader, outgoing bool) {
	ctype, att := contentType(att)
	b := []byte{}
	if att != nil {
		var err error
		b, err = ioutil.ReadAll(att)
		if err != nil {
			showError(err)
		}
	}
	message := &Message{Text: text,
		Outgoing: outgoing,
		Time:     time.Now(),
		From:     from,
		HTime:    humanize.Time(time.Now()),
		CType:    ctype,
		Att:      b,
	}
	s.messages = append(s.messages, message)
	s.Last = text
	s.Len++
	s.When = s.messages[0].HTime
	s.CType = s.messages[0].CType
	qml.Changed(s, &s.Last)
	qml.Changed(s, &s.When)
	qml.Changed(s, &s.Len)
	qml.Changed(s, &s.CType)
}

// updateTimestamps keeps the timestamps of the last message of each session
// updated in human readable form
func updateTimestamps() {
	for {
		time.Sleep(1 * time.Minute)
		for _, s := range sessionsModel.m {
			if s.Len == 0 {
				continue
			}
			s.When = s.messages[0].HTime
			qml.Changed(s, &s.When)
			for _, m := range s.messages {
				m.HTime = humanize.Time(m.Time)
				qml.Changed(m, &m.HTime)
			}
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
