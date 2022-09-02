package store

import (
	"errors"
	"fmt"

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

func SaveMessage(m *Message) (*Message, error) {
	//get last messageid
	var lastMessageID = []Message{}
	err := DS.Dbx.Select(&lastMessageID, "SELECT id FROM messages ORDER BY id DESC LIMIT 1")
	if err == nil {
		m.ID = lastMessageID[0].ID + 1
	} else {
		m.ID = 0
	}
	res, err := DS.Dbx.NamedExec(messagesInsert, m)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	m.ID = id
	return m, nil
}

func UpdateMessageSent(m *Message) error {
	if m.SendingError {
		log.Errorln("[axolotl] sending message failed ", m.SentAt)
	}
	_, err := DS.Dbx.NamedExec("UPDATE messages SET sentat = :sentat, sendingError = :sendingError,  issent = :issent, expireTimer = :expireTimer  WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}

func UpdateMessageRead(m *Message) error {
	_, err := DS.Dbx.NamedExec("UPDATE messages SET isread = :isread, issent = :issent, receipt = :receipt WHERE SendingError = 0 AND Outgoing = 1 AND Source = :source", m)
	if err != nil {
		return err
	}
	return err
}
func UpdateMessageReceiptSent(m *Message) error {
	_, err := DS.Dbx.NamedExec("UPDATE messages SET issent = :issent WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}
func UpdateMessageReceipt(m *Message) error {
	_, err := DS.Dbx.NamedExec("UPDATE messages SET issent = :issent, receipt = :receipt WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}
func LoadGroups() error {
	log.Printf("Loading Groups")
	err := DS.Dbx.Select(&AllGroups, groupsSelect)
	if err != nil {
		return err
	}
	for _, g := range AllGroups {
		Groups[g.GroupID] = g
	}
	return nil
}

func DeleteMessage(id int64) error {
	err := deleteAttachmentForMessage(id)
	if err != nil {
		log.Errorln("[axolotl] could not delete attachment", err)
		return err
	}
	_, err = DS.Dbx.Exec("DELETE FROM messages WHERE id = ?", id)
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
func FindQuotedMessage(quote *signalservice.DataMessage_Quote) (error, int64) {
	var quotedMessages = []Message{}
	err := DS.Dbx.Select(&quotedMessages, "SELECT * FROM messages WHERE sentat = ?", quote.GetId())
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
func GetMessageById(id int64) (*Message, error) {
	var message = []Message{}
	err := DS.Dbx.Select(&message, "SELECT * FROM messages WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(message) == 0 {
		return nil, errors.New("Message not found " + fmt.Sprint(id))
	}
	return &message[0], nil
}

// FindOutgoingMessage returns  a message that is found by it's timestamp
func FindOutgoingMessage(timestamp uint64) (*Message, error) {
	var message = []Message{}
	log.Debugln("[axolotl] searching for outgoing message ", timestamp)
	err := DS.Dbx.Select(&message, "SELECT * FROM messages WHERE outgoing = 1 AND sentat = ?", timestamp)
	if err != nil {
		return nil, err
	}
	if len(message) == 0 {
		return nil, errors.New("Message not found " + fmt.Sprint(timestamp))
	}
	return &message[0], nil
}

// GetUnreadMessagesCounterForSession returns an int for the unread messages for a session
func GetUnreadMessageCounterForSession(id int64) (int64, error) {
	var message = []Message{}
	err := DS.Dbx.Select(&message, "SELECT * FROM messages WHERE isread = 0 AND sessionid = ?", id)
	if err != nil {
		return 0, err
	}
	return int64(len(message)), nil
}

// GetLastMessageForSession returns the last message in a session
func GetLastMessageForSession(id int64) (*Message, error) {
	var message = []Message{}
	err := DS.Dbx.Select(&message, "SELECT * FROM messages WHERE sid = ? ORDER BY sentat DESC LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	if len(message) == 0 {
		return nil, errors.New("Message not found " + fmt.Sprint(id))
	}
	return &message[0], nil
}
func GetLastMessagesForAllSessions() ([]Message, error) {
	var messages = []Message{}
	var sessions = []Session{}
	err := DS.Dbx.Select(&sessions, "SELECT id FROM sessionsv2")
	if err != nil {
		return nil, err
	}
	for _, s := range sessions {
		m, err := GetLastMessageForSession(s.ID)
		if err == nil {
			messages = append(messages, *m)
		}
	}
	return messages, nil
}
