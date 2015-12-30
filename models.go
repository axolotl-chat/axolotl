package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gosexy/gettext"
	"github.com/janimo/textsecure"
	"github.com/janimo/textsecure/3rd_party/magic"
	"gopkg.in/qml.v1"
	"gopkg.in/yaml.v2"
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

type Settings struct {
	SendByEnter bool `yaml:"sendByEnter"`
}

var settingsModel *Settings

func loadSettings() (*Settings, error) {
	s := &Settings{}

	b, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		return s, err
	}

	err = yaml.Unmarshal(b, s)
	if err != nil {
		return s, err
	}
	return s, nil
}

func saveSettings(s *Settings) error {
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(settingsFile, b, 0600)
}

func (api *textsecureAPI) SaveSettings() error {
	return saveSettings(settingsModel)
}

var (
	msgFlagGroupNew     = 1
	msgFlagGroupUpdate  = 2
	msgFlagGroupLeave   = 4
	msgFlagResetSession = 8
)

// Model for existing chat sessions

type Message struct {
	ID         int64
	SID        int64
	Source     string
	Message    string
	Outgoing   bool
	SentAt     uint64
	ReceivedAt uint64
	HTime      string
	CType      int
	Attachment string
	IsSent     bool
	IsRead     bool
	Flags      int
}

func (m *Message) Info() string {
	timeFormat := time.RFC1123Z
	s := gettext.Gettext("Sent") + ": " + time.Unix(int64(m.SentAt/1000), 0).Format(timeFormat)
	if m.ReceivedAt != 0 {
		s += "\n" + gettext.Gettext("Received") + ": " + time.Unix(int64(m.ReceivedAt/1000), 0).Format(timeFormat)
	}
	return s
}

func (m *Message) Name() string {
	return telToName(m.Source)
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
	Unread    int
	Active    bool
	Len       int
}

type Sessions struct {
	sessions []*Session
	Len      int
}

func (s *Sessions) Session(i int) *Session {
	return s.sessions[i]
}

func (s *Sessions) GetIndex(tel string) int {
	for i, ses := range s.sessions {
		if ses.Tel == tel {
			return i
		}
	}
	return -1
}

func (s *Sessions) Get(tel string) *Session {
	for _, ses := range s.sessions {
		if ses.Tel == tel {
			return ses
		}
	}
	ses := &Session{Tel: tel, Name: telToName(tel), Active: true, IsGroup: tel[0] != '+'}
	s.sessions = append(s.sessions, ses)
	s.Len++
	qml.Changed(s, &s.Len)
	saveSession(ses)
	return ses
}

func (s *Session) MarkRead() {
	s.Unread = 0
	qml.Changed(s, &s.Unread)
	updateSession(s)
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
	Active  bool
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

func mimeTypeToContentType(mt string) int {
	ct := ContentTypeMessage
	if strings.HasPrefix(mt, "image") {
		ct = ContentTypePictures
	}
	if strings.HasPrefix(mt, "video") {
		ct = ContentTypeVideos
	}
	if strings.HasPrefix(mt, "audio") {
		ct = ContentTypeMusic
	}
	return ct
}

// contentType gets the content type (message, picture, video) of an attachment by sniffing its MIME type
func contentType(att io.Reader, mt string) int {
	if att == nil {
		return ContentTypeMessage
	}
	if mt == "" {
		mt, _ = magic.MIMETypeFromReader(att)
	}
	return mimeTypeToContentType(mt)
}

func (s *Session) Add(text string, source string, file string, mimetype string, outgoing bool) *Message {

	ctype := ContentTypeMessage
	if file != "" {
		f, _ := os.Open(file)
		ctype = contentType(f, mimetype)
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
	s.messages = append(s.messages, message)
	s.Last = text
	s.Len++
	s.CType = ctype
	qml.Changed(s, &s.Last)
	qml.Changed(s, &s.Len)
	qml.Changed(s, &s.CType)
	if !outgoing && api.ActiveSessionID != s.Tel {
		s.Unread++
		qml.Changed(s, &s.Unread)
	}
	updateSession(s)

	s.moveToTop()
	return message
}

var topSession string

// moveToTop makes sure the most recently updated session gets to the top of the session list
// it is hacky due to the way models in Go-QML cannot be mutated in a straightforward way
func (s *Session) moveToTop() {
	if topSession == s.Tel {
		return
	}

	index := sessionsModel.GetIndex(s.Tel)
	sessionsModel.sessions = append([]*Session{s}, append(sessionsModel.sessions[:index], sessionsModel.sessions[index+1:]...)...)

	// force a length change update
	sessionsModel.Len--
	qml.Changed(sessionsModel, &sessionsModel.Len)
	sessionsModel.Len++
	qml.Changed(sessionsModel, &sessionsModel.Len)

	topSession = s.Tel
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
	var err error
	settingsModel, err = loadSettings()
	if err != nil {
		log.Println(err)
	}
	engine.Context().SetVar("contactsModel", contactsModel)
	engine.Context().SetVar("settingsModel", settingsModel)
	engine.Context().SetVar("sessionsModel", sessionsModel)

	go updateTimestamps()
}
