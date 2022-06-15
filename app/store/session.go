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
	"github.com/signal-golang/textsecure/groupsv2"
	log "github.com/sirupsen/logrus"
)

// Session defines how a session looks like
type Session struct {
	ID              int64
	UUID            string `db:"uuid"`
	Name            string
	Tel             string
	IsGroup         bool  `db:"isgroup"`
	Type            int32 //describes the type of the session, wether it's a private conversation or groupv1 or groupv2
	Last            string
	Timestamp       uint64
	When            string
	CType           int
	Messages        []*Message
	Unread          int
	Active          bool
	Len             int
	Notification    bool
	ExpireTimer     uint32 `db:"expireTimer"`
	GroupJoinStatus int    `db:"groupJoinStatus"`
}
type MessageList struct {
	ID       int64
	Session  *Session
	Messages []*Message
}
type Sessions struct {
	Sess         []*Session
	ActiveChat   string
	TopSessionID int64
	Len          int
	Type         string
}

// SessionTypes
const (
	invalidSession                  = -1
	invalidQuote                    = -1
	SessionTypePrivateChat    int32 = 0
	SessionTypeGroupV1        int32 = 1
	SessionTypeGroupV2        int32 = 2
	SessionTypeGroupV2Invited int32 = 3
)

//TODO that hasn't to  be in the db controller
// var AllSessions []*Session // TODO: WIP 831

// var SessionsModel = &Sessions{ // TODO: WIP 831
// 	Sess: make([]*Session, 0),
// }

// SaveSession saves a newly created session in the database
func (s *Store) SaveSession(sess *Session) (*Session, error) {
	res, err := s.DS.Dbx.NamedExec(sessionsInsert, sess)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	sess.ID = id
	return sess, err
}

// UpdateSession updates a session in the database
func (s *Store) UpdateSession(sess *Session) error {
	_, err := s.DS.Dbx.NamedExec("UPDATE sessions SET name = :name, timestamp = :timestamp, ctype = :ctype, last = :last, unread = :unread, notification = :notification, expireTimer = :expireTimer, uuid = :uuid WHERE id = :id", sess)
	if err != nil {
		return err
	}
	return err
}

// DeleteSession deletes a session in the database
func (s *Store) DeleteSession(ID int64) error {
	var messagesWithAttachment = []Message{}

	err := s.DS.Dbx.Select(&messagesWithAttachment, "SELECT * FROM messages WHERE attachment NOT LIKE null AND id = ? ", ID)
	if err != nil {
		return err
	}
	if len(messagesWithAttachment) > 0 {
		for _, message := range messagesWithAttachment {
			err := s.deleteAttachmentForMessage(message.ID)
			if err != nil {
				return err
			}
		}
	}

	_, err = s.DS.Dbx.Exec("DELETE FROM messages WHERE sid = ?", ID)
	if err != nil {
		return err
	}
	_, err = s.DS.Dbx.Exec("DELETE FROM sessions WHERE id = ?", ID)
	if err != nil {
		return err
	}

	s.LoadChats()
	return nil
}

// GetSession at a certain index
func (s *Sessions) GetSession(i int64) *Session {
	return s.Sess[i]
}

// GetMessageList returns the message list for the session id
func (s *Store) GetMessageList(ID int64) (error, *MessageList) {
	if ID != invalidSession {
		sess, err := s.Sessions.Get(ID)
		if err != nil {
			log.Errorln("[axolotl] get messagelist", err)
			return err, nil
		}
		messageList := &MessageList{
			ID:      ID,
			Session: sess,
		}
		err = s.DS.Dbx.Select(&messageList.Messages, messagesSelectWhere, ID)
		if err != nil {
			log.Errorln("[axolotl] get messagelist", err)
			return err, nil
		}
		// attach the quoted messages
		for i, m := range messageList.Messages {
			if m.Flags == helpers.MsgFlagQuote {
				if m.QuoteID != invalidQuote {
					qm, err := s.GetMessageById(m.QuoteID)
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
func (s *Store) GetMoreMessageList(ID int64, lastID string) (error, *MessageList) {
	if ID != -1 {
		sess, err := s.Sessions.Get(ID)
		if err != nil {
			log.Errorln("[axolotl] GetMoreMessageList", err)
			return err, nil
		}
		messageList := &MessageList{
			ID:      ID,
			Session: sess,
		}
		err = s.DS.Dbx.Select(&messageList.Messages, messagesSelectWhereMore, messageList.Session.ID, lastID)
		if err != nil {
			log.Errorln("[axolotl] GetMoreMessageList", err)
			return err, nil
		}
		// attach the quoted messages
		for i, m := range messageList.Messages {
			if m.Flags == helpers.MsgFlagQuote {
				if m.QuoteID != -1 {
					qm, err := s.GetMessageById(m.QuoteID)
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

// Add message to a session // TODO: WIP 831 - remove Store from args
func (s *Session) Add(text string, source string, file []Attachment, mimetype string, outgoing bool, sessionID int64) *Message {
	var files []Attachment

	ctype := helpers.ContentTypeMessage
	if len(file) > 0 {
		for _, fi := range file {
			f, _ := os.Open(fi.File)
			if fi.CType == 0 {
				ctype = helpers.ContentType(f, mimetype)
			} else {
				ctype = fi.CType
			}
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
	// Only increments the counter for incoming messages, and only if the
	// user is not currently on the conversation
	if !outgoing && s.ID != sessionID && text != "" && text != "readReceiptMessage" && text != "deliveryReceiptMessage" {
		s.Unread++
	}

	return message
}

// MarkRead marks a session as read
func (s *Session) MarkRead(store *Store) {
	s.Unread = 0
	store.UpdateSession(s)
}

// ToggleSessionNotification turns on/off notification for a session
func (s *Session) ToggleSessionNotification(store *Store) {
	s.Notification = !s.Notification
	txt := ""
	if s.Notification {
		txt = "notifications on"
	} else {
		txt = "notifications off"

	}
	log.Debugln("[axolotl] ToggleSessionNotification ", txt)
	store.UpdateSession(s)
}

// UpdateTimestamps keeps the timestamps of the last message of each session
// updated in human readable form.
func (s *Store) UpdateTimestamps() {
	for {
		time.Sleep(1 * time.Minute)
		for _, sess := range s.Sessions.Sess {
			if sess.Len == 0 {
				continue
			}
			for _, m := range sess.Messages {
				m.HTime = helpers.HumanizeTimestamp(m.SentAt)
			}
			sess.When = sess.Messages[len(sess.Messages)-1].HTime
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
func (s *Store) GetByE164(tel string) *Session {
	for _, ses := range s.Sessions.Sess {
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
func (s *Store) CreateSessionForE164(tel string, UUID string) *Session {
	ses := &Session{Tel: tel,
		Name:         s.Contacts.TelToName(tel),
		Active:       true,
		IsGroup:      false,
		Notification: true,
		UUID:         UUID,
		Type:         SessionTypePrivateChat,
	}
	s.Sessions.Sess = append(s.Sessions.Sess, ses)
	s.Sessions.Len++
	s.SaveSession(ses)
	return ses
}

func (s *Store) CreateSessionForUUID(UUID string) *Session {
	contact := s.Contacts.GetContactForUUID(UUID)
	newSession := &Session{
		Tel:          contact.Tel,
		Name:         contact.Name,
		Active:       true,
		IsGroup:      false,
		Notification: true,
		UUID:         UUID,
	}
	if s.Sessions.Len == 0 {
		newSession.ID = 1
	}
	newSession, err := s.SaveSession(newSession)

	if err != nil {
		log.Errorln("[axolotl] CreateSessionForUUID failed:", err)
		return nil
	}
	s.Sessions.Sess = append(s.Sessions.Sess, newSession)
	s.Sessions.Len = len(s.Sessions.Sess)

	message := &Message{
		Message:    "Chat created",
		SID:        newSession.ID,
		ChatID:     newSession.Tel,
		Source:     newSession.Tel,
		SourceUUID: newSession.UUID,
		Outgoing:   true,
		Flags:      helpers.MsgFlagChatCreated,
		HTime:      "Now",
		SentAt:     uint64(time.Now().UnixNano() / 1000000),
	}
	s.SaveMessage(message)
	newSession.Messages = append(newSession.Messages, message)
	newSession.Last = message.Message
	s.UpdateSession(newSession)
	return newSession
}

// CreateSessionForGroup creates a session for a group
func (s *Store) CreateSessionForGroup(group *textsecure.Group) *Session {
	ses := &Session{Tel: group.Hexid, // for legacy reasons add group id also as Tel number
		Name:            group.Name,
		Active:          true,
		IsGroup:         true,
		Notification:    true,
		UUID:            group.Hexid,
		Type:            SessionTypeGroupV1,
		GroupJoinStatus: 0,
	}
	s.Sessions.Sess = append(s.Sessions.Sess, ses)
	s.Sessions.Len++
	ses, err := s.SaveSession(ses)
	if err != nil {
		log.Errorln("CreateSessionForGroup failed:", err)
		return nil
	}
	return ses
}

// CreateSessionForGroupV2 creates a session for a group
func (s *Store) CreateSessionForGroupV2(group *groupsv2.GroupV2) *Session {
	ses := &Session{Tel: group.Hexid, // for legacy reasons add group id also as Tel number
		Name:            string(group.DecryptedGroup.Title),
		Active:          true,
		IsGroup:         true,
		Notification:    true,
		UUID:            group.Hexid,
		Type:            SessionTypeGroupV2,
		GroupJoinStatus: group.JoinStatus,
	}
	s.Sessions.Sess = append(s.Sessions.Sess, ses)
	s.Sessions.Len++
	ses, err := s.SaveSession(ses)
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
			return ses, nil
		}
	}
	return nil, fmt.Errorf("Session with uuid %s not found", UUID)
}

// UpdateSessionNames updates the non groups with the name from the phone book
func (s *Store) UpdateSessionNames() {
	log.Debugln("[axolotl] update session names + uuids")
	for _, ses := range s.Sessions.Sess {
		if ses.IsGroup == false {
			ses.Name = s.Contacts.TelToName(ses.Tel)
			if ses.UUID == "" || ses.UUID == "0" {
				c := s.Contacts.GetContactForTel(ses.Tel)
				if c != nil && c.UUID != "" && c.UUID != "0" && (c.UUID[0] != 0 || c.UUID[len(c.UUID)-1] != 0) {
					uuid := c.UUID
					log.Debugln("[axolotl] update session from tel to uuid", ses.Tel, uuid)
					index := strings.Index(uuid, "-")

					if index == -1 {
						uuid = helpers.HexToUUID(uuid)
					}
					ses.UUID = uuid
				}
			}

			s.UpdateSession(ses)
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

func (s *Sessions) MoveToTop(sessionID int64) {
	if s.TopSessionID == sessionID {
		return
	}

	index := s.GetIndex(sessionID)
	session := s.GetSession(sessionID)
	s.Sess = append([]*Session{session}, append(s.Sess[:index], s.Sess[index+1:]...)...)

	// force a length change update
	s.Len--
	s.Len++

	s.TopSessionID = sessionID
}

func (s *Store) LoadChats() error {
	log.Printf("[axolotl] Loading Chats")
	err := s.DS.Dbx.Select(&AllGroups, groupsSelect)
	if err != nil {
		return err
	}
	// Reset groups
	newGroups := map[string]*GroupRecord{}
	Groups = newGroups
	for _, g := range AllGroups {
		Groups[g.GroupID] = g
	}

	// Reset session model
	s.Sessions.Sess = make([]*Session, 0)
	s.Sessions.Len = 0
	s.AllSessions = []*Session{}
	for _, g := range AllGroups {
		Groups[g.GroupID] = g
	}

	err = s.DS.Dbx.Select(&s.AllSessions, sessionsSelect)
	if err != nil {
		return err
	}
	for _, sess := range s.AllSessions {
		sess.When = helpers.HumanizeTimestamp(sess.Timestamp)
		sess.Active = !sess.IsGroup || (Groups[sess.Tel] != nil && Groups[sess.Tel].Active)
		s.Sessions.Sess = append(s.Sessions.Sess, sess)
		s.Sessions.Len++
		err = s.DS.Dbx.Select(&sess.Messages, messagesSelectWhereLastMessage, sess.ID)
		// s.Len = len(s.Messages)
		if err != nil {
			return err
		}
		for _, m := range sess.Messages {
			m.HTime = helpers.HumanizeTimestamp(m.SentAt)
		}
	}
	return nil
}
