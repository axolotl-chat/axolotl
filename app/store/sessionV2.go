package store

import (
	"errors"
	"fmt"

	"github.com/nanu-c/axolotl/app/helpers"
	log "github.com/sirupsen/logrus"
)

var (
	sessionsV2Schema             = "CREATE TABLE IF NOT EXISTS sessionsv2 (id INTEGER PRIMARY KEY, directMessageRecipientId INTEGER,unreadCounter integer default 0, expireTimer integer default 0, isArchived boolean NOT NULL DEFAULT 0,isBlocked boolean NOT NULL DEFAULT 0,isPinned boolean NOT NULL DEFAULT 0,isSilenced boolean NOT NULL DEFAULT 0,isMuted boolean NOT NULL DEFAULT 0,draft text DEFAULT '',groupV2Id text,groupV1Id text);"
	sessionsV2Insert             = "INSERT or REPLACE INTO sessionsv2 (id, directMessageRecipientId,expireTimer,groupV2Id,groupV1Id) VALUES (:id, :directMessageRecipientId, :expireTimer, :groupV2Id, :groupV1Id);"
	sessionV2UpdateUnreadCounter = "UPDATE sessionsv2 SET unreadCounter = :unreadCounter WHERE id = :id;"
	GroupRecipientsID            = -1
)
var SessionsV2Model = &SessionsV2{
	Sess: make([]*SessionV2, 0),
}

type SessionV2 struct {
	ID                       int64  `db:"id"`
	DirectMessageRecipientID int64  `db:"directMessageRecipientId"`
	ExpireTimer              int64  `db:"expireTimer"`
	IsArchived               bool   `db:"isArchived"`
	IsBlocked                bool   `db:"isBlocked"`
	IsPinned                 bool   `db:"isPinned"`
	IsSilenced               bool   `db:"isSilenced"`
	IsMuted                  bool   `db:"isMuted"`
	Draft                    string `db:"draft"`
	GroupV2ID                string `db:"groupV2Id"`
	GroupV1ID                string `db:"groupV1Id"`
	UnreadCounter            int64  `db:"unreadCounter"`
}
type SessionsV2 struct {
	Sess []*SessionV2
}

type SessionV2Name struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// CreateSessionForGroupV2 creates a session for a group v2
func (s *SessionsV2) CreateSessionForGroupV2(group string) (*SessionV2, error) {
	ses := &SessionV2{
		GroupV2ID: group,
	}
	ses, err := s.SaveSession(ses)
	if err != nil {
		log.Errorln("[axolotl] CreateSessionForGroupv2 failed:", err)
		return nil, err
	}
	return ses, nil
}

// CreateSessionForGroupV1 creates a session for a group v1
func (s *SessionsV2) CreateSessionForGroup(group string) (*SessionV2, error) {
	ses := &SessionV2{
		GroupV1ID: group,
	}
	ses, err := s.SaveSession(ses)
	if err != nil {
		log.Errorln("[axolotl] CreateSessionForGroupv1 failed:", err)
		return nil, err
	}
	return ses, nil
}

// CreateSessionForDirectMessageRecipient creates a session for a direct message (one to one)
func (s *SessionsV2) CreateSessionForDirectMessageRecipient(recipient int64) (*SessionV2, error) {
	ses := &SessionV2{
		DirectMessageRecipientID: recipient,
	}
	ses, err := s.SaveSession(ses)
	if err != nil {
		log.Errorln("[axolotl] CreateSessionForDirectMesssageRecipient failed:", err)
		return nil, err
	}
	return ses, nil
}

// SaveSession saves a session to the database
func (s *SessionsV2) SaveSession(session *SessionV2) (*SessionV2, error) {
	res, err := DS.Dbx.NamedExec(sessionsV2Insert, session)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	session.ID = id
	return session, err
}

// GetSessionByID returns a session by id
func (s *SessionsV2) GetSessionByID(id int64) (*SessionV2, error) {
	ses := &SessionV2{}
	err := DS.Dbx.Get(ses, "SELECT * FROM sessionsv2 WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return ses, nil
}

// GetSessionByGroupV2ID returns a session by group v2 id
func (s *SessionsV2) GetSessionByGroupV2ID(group string) (*SessionV2, error) {
	ses := &SessionV2{}
	err := DS.Dbx.Get(ses, "SELECT * FROM sessionsv2 WHERE groupV2Id = ?", group)
	if err != nil {
		return nil, err
	}
	return ses, nil
}

// GetSessionByGroupV1ID returns a session by group v1 id
func (s *SessionsV2) GetSessionByGroupV1ID(group string) (*SessionV2, error) {
	ses := &SessionV2{}
	err := DS.Dbx.Get(ses, "SELECT * FROM sessionsv2 WHERE groupV1Id = ?", group)
	if err != nil {
		return nil, err
	}
	return ses, nil
}

// GetSessionByDirectMessageRecipientID returns a session by direct message recipient id
func (s *SessionsV2) GetSessionByDirectMessageRecipientID(recipient int64) (*SessionV2, error) {
	ses := &SessionV2{}
	err := DS.Dbx.Get(ses, "SELECT * FROM sessionsv2 WHERE directMessageRecipientId = ?", recipient)
	if err != nil {
		return nil, err
	}
	return ses, nil
}

// GetAllSessions returns all sessions
func (s *SessionsV2) GetAllSessions() ([]*SessionV2, error) {
	ses := make([]*SessionV2, 0)
	err := DS.Dbx.Select(&ses, "SELECT * FROM sessionsv2")
	if err != nil {
		return nil, err
	}
	return ses, nil
}

// DeleteSession deletes a session
func (s *SessionsV2) DeleteSession(session *SessionV2) error {
	_, err := DS.Dbx.NamedExec("DELETE FROM sessionsv2 WHERE id = :id", session)
	return err
}

// DeleteAllSessions deletes all sessions
func (s *SessionsV2) DeleteAllSessions() error {
	_, err := DS.Dbx.Exec("DELETE FROM sessionsv2")
	return err
}

// UpdateUnreadCounterForSession updates the unread counter for a session
func (s *SessionsV2) UpdateUnreadCounterForSession(session *SessionV2) error {
	unreadCounter, err := GetUnreadMessageCounterForSession(session.ID)
	if err != nil {
		return err
	}
	session.UnreadCounter = unreadCounter

	_, err = DS.Dbx.NamedExec(sessionV2UpdateUnreadCounter, session)
	return err
}

// UpdateAllUnreadCountersForSessions updates all unread counters for all sessions
func (s *SessionsV2) UpdateAllUnreadCountersForSessions() error {
	ses, err := s.GetAllSessions()
	if err != nil {
		return err
	}
	for _, session := range ses {
		err := s.UpdateUnreadCounterForSession(session)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSessionNames returns all session names
func (s *SessionsV2) GetSessionNames() ([]SessionV2Name, error) {
	ses, err := s.GetAllSessions()
	if err != nil {
		return nil, err
	}
	names := make([]SessionV2Name, 0)
	for _, session := range ses {
		name, err := session.GetName()
		if err != nil {
			name = "Unknown"
		}

		names = append(names, SessionV2Name{
			ID:   session.ID,
			Name: name,
		})
	}
	return names, nil
}

// isGroup returns true if the session is a group session
func (s *SessionV2) IsGroup() bool {
	if s.GroupV1ID != "" || s.GroupV2ID != "" {
		return true
	}
	return false
}

// GetMessageList returns a list of messages for a session
func (s *SessionV2) GetMessageList(limit int, offset int) ([]*Message, error) {
	messages := make([]*Message, 0)
	err := DS.Dbx.Select(&messages, "SELECT * FROM messages WHERE sid = ? ORDER BY sentat DESC LIMIT ? OFFSET ?", s.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// MarkRead marks a session as read
func (s *SessionV2) MarkRead() error {
	// set messages as read
	_, err := DS.Dbx.Exec("UPDATE messages SET read = 1 WHERE sId = ?", s.ID)
	if err != nil {
		return err
	}
	_, err = DS.Dbx.Exec("UPDATE sessionsv2 SET unreadCounter = 0 WHERE id = ?", s.ID)

	return err
}

// GetMoreMessageList loads more messages from before the lastID
func (s *SessionsV2) GetMoreMessageList(ID int64, lastID string) (error, *MessageList) {
	if ID != -1 {
		sess, err := s.GetSessionByID(ID)
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
					qm, err := GetMessageById(m.QuoteID)
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

// NotificationsToggletoggles the notifications for a session
func (s *SessionV2) NotificationsToggle() error {
	toggle := !s.IsMuted
	_, err := DS.Dbx.Exec("UPDATE sessionsv2 SET notifications = ? WHERE id = ?", toggle, s.ID)
	return err
}

// GetName returns the name of the session
func (s *SessionV2) GetName() (string, error) {
	if s.DirectMessageRecipientID != int64(GroupRecipientsID) {
		recipient := RecipientsModel.GetRecipientById(s.DirectMessageRecipientID)
		return recipient.Username, nil
	} else {
		group, err := GroupV2sModel.GetGroupById(s.GroupV2ID)
		if err != nil {
			return "", fmt.Errorf("GetSessionNames failed group: %s", err)
		}
		if group != nil {
			return group.Name, nil
		}
	}
	return "", fmt.Errorf("GetSessionNames failed")
}
