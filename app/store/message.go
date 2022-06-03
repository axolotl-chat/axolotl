package store

import (
	"errors"
	"fmt"

	"github.com/nanu-c/axolotl/app/helpers"
	signalservice "github.com/signal-golang/textsecure/protobuf"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	ID            int64 `db:"id"`
	SID           int64
	ChatID        string
	Source        string `db:"source"`
	SourceUUID    string `db:"srcUUID"`
	Message       string
	Outgoing      bool
	SentAt        uint64
	ReceivedAt    uint64
	HTime         string
	CType         int
	Attachment    string
	IsSent        bool `db:"issent"`
	IsRead        bool `db:"isread"`
	Flags         int
	ExpireTimer   uint32 `db:"expireTimer"`
	SendingError  bool   `db:"sendingError"`
	Receipt       bool   `db:"receipt"`
	StatusMessage bool   `db:"statusMessage"`
	QuoteID       int64  `db:"quoteId"`
	QuotedMessage *Message
}

func (s *Store) SaveMessage(m *Message) (error, *Message) {

	res, err := s.DS.Dbx.NamedExec(messagesInsert, m)
	if err != nil {
		return err, nil
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err, nil
	}

	m.ID = id
	return nil, m
}

func (s *Store) UpdateMessageSent(m *Message) error {
	if m.SendingError {
		log.Errorln("[axolotl] sending message failed ", m.SentAt)
	}
	_, err := s.DS.Dbx.NamedExec("UPDATE messages SET sentat = :sentat, sendingError = :sendingError,  issent = :issent, expireTimer = :expireTimer  WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}

func (s *Store) UpdateMessageRead(m *Message) error {
	_, err := s.DS.Dbx.NamedExec("UPDATE messages SET isread = :isread, issent = :issent, receipt = :receipt WHERE SendingError = 0 AND Outgoing = 1 AND Source = :source", m)
	if err != nil {
		return err
	}
	return err
}
func (s *Store) UpdateMessageReceiptSent(m *Message) error {
	_, err := s.DS.Dbx.NamedExec("UPDATE messages SET issent = :issent WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}
func (s *Store) UpdateMessageReceipt(m *Message) error {
	_, err := s.DS.Dbx.NamedExec("UPDATE messages SET issent = :issent, receipt = :receipt WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}
func (s *Store) LoadGroups() error {
	log.Printf("Loading Groups")
	err := s.DS.Dbx.Select(&AllGroups, groupsSelect)
	if err != nil {
		return err
	}
	for _, g := range AllGroups {
		Groups[g.GroupID] = g
	}
	return nil
}
func (s *Store) LoadMessagesFromDB() error {
	err := s.LoadGroups()
	if err != nil {
		return err
	}
	log.Printf("Loading Messages")
	err = s.DS.Dbx.Select(&s.AllSessions, sessionsSelect)
	if err != nil {
		return err
	}
	for _, sess := range s.AllSessions {
		sess.When = helpers.HumanizeTimestamp(sess.Timestamp)
		sess.Active = !sess.IsGroup || (Groups[sess.Tel] != nil && Groups[sess.Tel].Active)
		s.Sessions.Sess = append(s.Sessions.Sess, sess)
		s.Sessions.Len++
		err = s.DS.Dbx.Select(&sess.Messages, messagesSelectWhere, sess.ID)
		sess.Len = len(sess.Messages)
		if err != nil {
			return err
		}
		for _, m := range sess.Messages {
			m.HTime = helpers.HumanizeTimestamp(m.SentAt)
		}
	}
	return nil
}

func (s *Store) DeleteMessage(id int64) error {
	err := s.deleteAttachmentForMessage(id)
	if err != nil {
		log.Errorln("[axolotl] could not delete attachment", err)
		return err
	}
	_, err = s.DS.Dbx.Exec("DELETE FROM messages WHERE id = ?", id)
	return err
}

func (s *Session) GetMessages(i int) *Message {
	//FIXME when is index -1 ?
	if i == -1 || i >= len(s.Messages) {
		return &Message{}
	}
	return s.Messages[i]
}
func (m *Message) GetName() string {
	return TelToName(m.Source)
}

// FindQuotedMessage searches the equivalent message of DataMessage_Quote in our
// DB and returns the local message id
func (s *Store) FindQuotedMessage(quote *signalservice.DataMessage_Quote) (error, int64) {
	var quotedMessages = []Message{}
	err := s.DS.Dbx.Select(&quotedMessages, "SELECT * FROM messages WHERE sentat = ?", quote.GetId())
	if err != nil {
		return err, -1
	}
	if len(quotedMessages) == 0 {
		return errors.New("quoted message not found " + fmt.Sprint(quote.GetId())), -1
	}
	id := quotedMessages[0].ID
	return nil, id
}

// GetMessageById returns a message by it's ID
func (s *Store) GetMessageById(id int64) (*Message, error) {
	var message = []Message{}
	err := s.DS.Dbx.Select(&message, "SELECT * FROM messages WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(message) == 0 {
		return nil, errors.New("Message not found " + fmt.Sprint(id))
	}
	return &message[0], nil
}

// FindOutgoingMessage returns  a message that is found by it's timestamp
func (s *Store) FindOutgoingMessage(timestamp uint64) (*Message, error) {
	var message = []Message{}
	err := s.DS.Dbx.Select(&message, "SELECT * FROM messages WHERE outgoing = 1 AND sentat = ?", timestamp)
	if err != nil {
		return nil, err
	}
	if len(message) == 0 {
		return nil, errors.New("Message not found " + fmt.Sprint(timestamp))
	}
	return &message[0], nil
}
