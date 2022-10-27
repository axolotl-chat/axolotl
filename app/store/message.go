package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	signalservice "github.com/signal-golang/textsecure/protobuf"
	log "github.com/sirupsen/logrus"
)

const (
	lastMessageIdQuery                = "SELECT id FROM messages ORDER BY id DESC LIMIT 1"
	updateMessageSentQuery            = "UPDATE messages SET sentat = :sentat, sendingError = :sendingError,  issent = :issent, expireTimer = :expireTimer  WHERE id = :id"
	updateMessageReadQuery            = "UPDATE messages SET isread = :isread, issent = :issent, receipt = :receipt WHERE SendingError = 0 AND Outgoing = 1 AND Source = :source"
	updateMessageReceiptSentQuery     = "UPDATE messages SET issent = :issent WHERE id = :id"
	updateMessageReceiptQuery         = "UPDATE messages SET issent = :issent, receipt = :receipt WHERE id = :id"
	deleteMessageQuery                = "DELETE FROM messages WHERE id = ?"
	findMessageBySentAtQuery          = "SELECT * FROM messages WHERE sentat = ?"
	findMessageByIdQuery              = "SELECT * FROM messages WHERE id = ?"
	findOutgoingMessageBySendAtQuery  = "SELECT * FROM messages WHERE outgoing = 1 AND sentat = ?"
	findUnreadMessagesForSessionQuery = "SELECT id FROM messages WHERE isread = 0 AND sessionid = ?"
	findMessagesForSession            = "SELECT * FROM messages WHERE sid = ? ORDER BY sentat DESC LIMIT ? OFFSET ?"
	getLastMessagesQuery              = "SELECT *, max(sentat) FROM messages GROUP BY sid ORDER BY sentat DESC"
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
	// get last messageid
	var lastId int64
	err := DS.Dbx.Get(&lastId, lastMessageIdQuery)
	if err != nil {
		lastId = 0
	}
	m.ID = lastId + 1
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
	_, err := DS.Dbx.NamedExec(updateMessageSentQuery, m)
	if err != nil {
		return err
	}
	return err
}

func UpdateMessageRead(m *Message) error {
	_, err := DS.Dbx.NamedExec(updateMessageReadQuery, m)
	if err != nil {
		return err
	}
	return err
}
func UpdateMessageReceiptSent(m *Message) error {
	_, err := DS.Dbx.NamedExec(updateMessageReceiptSentQuery, m)
	if err != nil {
		return err
	}
	return err
}
func UpdateMessageReceipt(m *Message) error {
	_, err := DS.Dbx.NamedExec(updateMessageReceiptQuery, m)
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
	_, err = DS.Dbx.Exec(deleteMessageQuery, id)
	return err
}

func (m *Message) GetName() string {
	return TelToName(m.Source)
}

// FindQuotedMessage searches the equivalent message of DataMessage_Quote in our
// DB and returns the local message id
func FindQuotedMessage(quote *signalservice.DataMessage_Quote) (error, int64) {
	var quotedMessages = []Message{}
	err := DS.Dbx.Select(&quotedMessages, findMessageBySentAtQuery, quote.GetId())
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
	err := DS.Dbx.Select(&message, findMessageByIdQuery, id)
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
	err := DS.Dbx.Select(&message, findOutgoingMessageBySendAtQuery, timestamp)
	if err != nil {
		return nil, err
	}
	if len(message) == 0 {
		return nil, errors.New("Message not found " + fmt.Sprint(timestamp))
	}
	return &message[0], nil
}

// GetUnreadMessageCounterForSession returns an int for the unread messages for a session
func GetUnreadMessageCounterForSession(id int64) (int64, error) {
	var message = []Message{}
	err := DS.Dbx.Select(&message, findUnreadMessagesForSessionQuery, id)
	if err != nil {
		return 0, err
	}
	return int64(len(message)), nil
}

func GetLastMessagesForAllSessions() ([]Message, error) {
	var messages = []Message{}
	unsafeDbx := DS.Dbx.Unsafe() // needed, because max(sentat) has no destination
	err := unsafeDbx.Select(&messages, getLastMessagesQuery)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func getMessagesForSession(id int64, limit, offset int) ([]*Message, error) {
	var messages = []*Message{}
	err := DS.Dbx.Select(&messages, findMessagesForSession, id, limit, offset)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		m := &Message{Message: "New chat created",
			SID:         id,
			Outgoing:    true,
			Source:      "",
			SourceUUID:  config.Config.UUID,
			HTime:       "Now",
			SentAt:      uint64(time.Now().UnixNano() / 1000000),
			ExpireTimer: uint32(0),
			Flags:       helpers.MsgFlagChatCreated,
		}
		SaveMessage(m)
		messages = append(messages, m)
	}
	return messages, nil
}
