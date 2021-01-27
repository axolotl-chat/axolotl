package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/signal-golang/textsecure"
	log "github.com/sirupsen/logrus"
)

// Session defines how a session looks like
type Session struct {
	ID           int64
	UUID         string `db:"uuid"`
	Name         string
	Tel          string
	IsGroup      bool
	Type         int64 //describes the type of the session, wether it's a private conversation or groupv1 or groupv2
	Last         string
	Timestamp    uint64
	When         string
	CType        int
	Messages     []*Message
	Unread       int
	Active       bool
	Len          int
	Notification bool
	ExpireTimer  uint32 `db:"expireTimer"`
}
type MessageList struct {
	ID       int64
	Session  *Session
	Messages []*Message
}
type Sessions struct {
	Sess       []*Session
	ActiveChat string
	Len        int
	Type       string
}

const (
	invalidSession = -1
	invalidQuote   = -1
)

//TODO that hasn't to  be in the db controller
var AllSessions []*Session

var SessionsModel = &Sessions{
	Sess: make([]*Session, 0),
}

// SaveSession saves a newly created session in the database
func SaveSession(s *Session) (*Session, error) {
	res, err := DS.Dbx.NamedExec(sessionsInsert, s)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	s.ID = id
	return s, err
}

// UpdateSession updates a session in the database
func UpdateSession(s *Session) error {
	_, err := DS.Dbx.NamedExec("UPDATE sessions SET name = :name, timestamp = :timestamp, ctype = :ctype, last = :last, unread = :unread, notification = :notification, expireTimer = :expireTimer, uuid = :uuid WHERE id = :id", s)
	if err != nil {
		return err
	}
	return err
}

// DeleteSession deletes a session in the database
func DeleteSession(ID int64) error {
	_, err := DS.Dbx.Exec("DELETE FROM messages WHERE sid = ?", ID)
	if err != nil {
		return err
	}
	_, err = DS.Dbx.Exec("DELETE FROM sessions WHERE id = ?", ID)
	if err != nil {
		return err
	}

	LoadChats()
	return nil
}

// GetSession at a certain index
func (s *Sessions) GetSession(i int64) *Session {
	return s.Sess[i]
}

// GetMessageList returns the message list for the session id
func (s *Sessions) GetMessageList(ID int64) (error, *MessageList) {
	if ID != invalidSession {
		sess, err := s.Get(ID)
		if err != nil {
			log.Errorln("[axolotl] get messagelist", err)
			return err, nil
		}
		messageList := &MessageList{
			ID:      ID,
			Session: sess,
		}
		err = DS.Dbx.Select(&messageList.Messages, messagesSelectWhere, ID)
		if err != nil {
			log.Errorln("[axolotl] get messagelist", err)
			return err, nil
		}
		// attach the quoted messages
		for i, m := range messageList.Messages {
			if m.Flags == helpers.MsgFlagQuote {
				if m.QuoteID != invalidQuote {
					err, qm := GetMessageById(m.QuoteID)
					if err != nil {
						log.Debugln("[axolotl] messagelist quoted message: ", err)
					} else {
						m.QuotedMessage = qm
						messageList.Messages[i] = m
					}
				}
			}
		}
		if err != nil {
			log.Errorln("[axolotl] GetMessageList ", err)
			return err, nil
		}
		return nil, messageList
	}
	return errors.New("wrong index"), nil
}

// GetMoreMessageList loads more messages from before the lastID
func (s *Sessions) GetMoreMessageList(ID int64, lastID string) (error, *MessageList) {
	if ID != -1 {
		sess, err := s.Get(ID)
		if err != nil {
			log.Errorln("[axolotl] GetMoreMessageList", err)
			return err, nil
		}
		messageList := &MessageList{
			ID:      ID,
			Session: sess,
		}
		err = DS.Dbx.Select(&messageList.Messages, messagesSelectWhereMore, messageList.Session.ID, lastID)
		if err != nil {
			log.Errorln("[axolotl] GetMoreMessageList", err)
			return err, nil
		}
		// attach the quoted messages
		for i, m := range messageList.Messages {
			if m.Flags == helpers.MsgFlagQuote {
				if m.QuoteID != -1 {
					err, qm := GetMessageById(m.QuoteID)
					if err != nil {
						log.Debugln("[axolotl] messagelist quoted message: ", err)
					} else {
						m.QuotedMessage = qm
						messageList.Messages[i] = m
					}
				}
			}
		}
		return nil, messageList
	}
	return errors.New("wrong index"), nil
}

// Add message to a session
func (s *Session) Add(text string, source string, file []Attachment, mimetype string, outgoing bool, sessionID int64) *Message {
	var files []Attachment

	ctype := helpers.ContentTypeMessage
	if len(file) > 0 {
		for _, fi := range file {
			f, _ := os.Open(fi.File)
			ctype = helpers.ContentType(f, mimetype)
			files = append(files, Attachment{File: fi.File, CType: ctype, FileName: fi.FileName})
		}
	}
	fJSON, err := json.Marshal(files)
	if err != nil {
		log.Errorln("[axolotl] chatlist add", err)
	}
	message := &Message{Message: text,
		SID:        s.ID,
		ChatID:     s.Tel,
		Outgoing:   outgoing,
		Source:     source,
		CType:      ctype,
		Attachment: string(fJSON),
		HTime:      "Now",
		SentAt:     uint64(time.Now().UnixNano() / 1000000),
	}
	s.Messages = append(s.Messages, message)
	s.Last = text
	s.Len++
	s.CType = ctype
	//FIXME not shure if it breaks unread message counter
	if !outgoing {
		s.Unread++
		//qml.Changed(s, &s.Unread)
	}
	UpdateSession(s)

	s.moveToTop()
	return message
}

// MarkRead marks a session as read
func (s *Session) MarkRead() {
	s.Unread = 0
	UpdateSession(s)
}

// ToggleSessionNotifcation turns on/off notification for a session
func (s *Session) ToggleSessionNotifcation() {
	s.Notification = !s.Notification
	txt := ""
	if s.Notification {
		txt = "notifications on"
	} else {
		txt = "notifications off"

	}
	//qml.Changed(s, &s.Notification)
	log.Debugln("[axolotl] ToggleSessionNotifcation ", txt)
	UpdateSession(s)
}

// UpdateTimestamps keeps the timestamps of the last message of each session
// updated in human readable form.
func UpdateTimestamps() {
	for {
		time.Sleep(1 * time.Minute)
		for _, s := range SessionsModel.Sess {
			if s.Len == 0 {
				continue
			}
			for _, m := range s.Messages {
				m.HTime = helpers.HumanizeTimestamp(m.SentAt)
			}
			s.When = s.Messages[len(s.Messages)-1].HTime
		}
	}
}

// Get returns the session by it's id
func (s *Sessions) Get(id int64) (*Session, error) {
	for _, ses := range s.Sess {
		if ses.ID == id {
			return ses, nil
		}
	}
	return nil, fmt.Errorf("Session with id %d not found", id)
}

// GetByE164 returns the session by the telephone number and creates it if it doesn't exists
func (s *Sessions) GetByE164(tel string) *Session {
	for _, ses := range s.Sess {

		if ses.Tel == tel {
			return ses
		}
	}
	newSession := s.CreateSessionForE164(tel, "0")
	return newSession
}
// GetAllSessionsByE164 returns multiple sessions when they are duplicated
func (s *Sessions) GetAllSessionsByE164(tel string) []*Session {
	var sessions = []*Session{}
	for _, ses := range s.Sess {
		if ses.Tel == tel {
			sessions = append(sessions, ses)
		}
	}
	return sessions
}
// CreateSessionForE164 creates a new Session for the phone number
func (s *Sessions) CreateSessionForE164(tel string, UUID string) *Session {
	ses := &Session{Tel: tel,
		Name:         TelToName(tel),
		Active:       true,
		IsGroup:      false,
		Notification: true,
		UUID:         UUID,
	}
	s.Sess = append(s.Sess, ses)
	s.Len++
	SaveSession(ses)
	return ses
}

func (s *Sessions) CreateSessionForUUID(UUID string) *Session {
	contact := GetContactForUUID(UUID)
	ses := &Session{Tel: contact.Tel,
		Name:         contact.Name,
		Active:       true,
		IsGroup:      false,
		Notification: true,
		UUID:         UUID,
	}
	s.Sess = append(s.Sess, ses)
	s.Len++
	SaveSession(ses)
	return ses
}

// CreateSessionForGroup creates a session for a group
func (s *Sessions) CreateSessionForGroup(group *textsecure.Group) *Session {
	ses := &Session{Tel: group.Hexid, // for legacy reasons add group id also as Tel number
		Name:         group.Name,
		Active:       true,
		IsGroup:      true,
		Notification: true,
		UUID:         group.Hexid,
	}
	s.Sess = append(s.Sess, ses)
	s.Len++
	ses, err := SaveSession(ses)
	if err != nil {
		log.Errorln("[axolotl] CreateSessionForGroup failed:", err)
		return nil
	}
	return ses
}

// GetByUUID returns the session by the ChatUUID
func (s *Sessions) GetByUUID(UUID string) (*Session, error) {
	if len(UUID) == 0 {
		return nil, fmt.Errorf("Empty session id %s", UUID)
	}
	for _, ses := range s.Sess {
		if ses.UUID == UUID {
			log.Debugln("Session with uuid found", UUID)

			return ses, nil
		}
	}
	return nil, fmt.Errorf("Session with uuid %s not found", UUID)
}
func HexToUUID(id string) string {
	if len(id) != 32 {
		return id
	}
	msbHex := id[:16]
	lsbHex := id[16:]
	return msbHex[:8] + "-" + msbHex[8:12] + "-" + msbHex[12:] + "-" + lsbHex[:4] + "-" + lsbHex[4:]
}

// UpdateSessionNames updates the non groups with the name from the phone book
func (s *Sessions) UpdateSessionNames() {
	log.Debugln("[axoltl] update session names + uuids")
	for _, ses := range s.Sess {
		if ses.IsGroup == false {
			ses.Name = TelToName(ses.Tel)
			if ses.UUID == "" || ses.UUID == "0" {
				c := GetContactForTel(ses.Tel)
				if c != nil && c.UUID != "" && c.UUID != "0" && (c.UUID[0] != 0 || c.UUID[len(c.UUID)-1] != 0) {
					uuid := c.UUID
					log.Debugln("[axolotl] update session from tel to uuid", ses.Tel, uuid)
					index := strings.Index(uuid, "-")

					if index == -1 {
						uuid = HexToUUID(uuid)
					}
					ses.UUID = uuid
				}
			}

			UpdateSession(ses)
		}
	}
}

// GetIndex returns the current index of session ID
func (s *Sessions) GetIndex(ID int64) int {
	for i, ses := range s.Sess {
		if ses.ID == ID {
			return i
		}
	}
	return -1
}

var topSession int64

func (s *Session) moveToTop() {
	if topSession == s.ID {
		return
	}

	index := SessionsModel.GetIndex(s.ID)
	SessionsModel.Sess = append([]*Session{s}, append(SessionsModel.Sess[:index], SessionsModel.Sess[index+1:]...)...)

	// force a length change update
	SessionsModel.Len--
	SessionsModel.Len++

	topSession = s.ID
}
func LoadChats() error {
	log.Printf("[axolotl] Loading Chats")
	err := DS.Dbx.Select(&AllGroups, groupsSelect)
	if err != nil {
		return err
	}
	for _, g := range AllGroups {
		Groups[g.GroupID] = g
	}

	err = DS.Dbx.Select(&AllSessions, sessionsSelect)
	if err != nil {
		return err
	}
	for _, s := range AllSessions {
		s.When = helpers.HumanizeTimestamp(s.Timestamp)
		s.Active = !s.IsGroup || (Groups[s.Tel] != nil && Groups[s.Tel].Active)
		SessionsModel.Sess = append(SessionsModel.Sess, s)
		SessionsModel.Len++
		err = DS.Dbx.Select(&s.Messages, messagesSelectWhereLastMessage, s.ID)
		// s.Len = len(s.Messages)
		if err != nil {
			return err
		}
		for _, m := range s.Messages {
			m.HTime = helpers.HumanizeTimestamp(m.SentAt)
		}
	}
	return nil
}
