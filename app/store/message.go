package store

import (
	"github.com/nanu-c/axolotl/app/helpers"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	ID            int64 `db:"id"`
	SID           int64
	ChatID        string
	Source        string
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
}

func SaveMessage(m *Message) (error, *Message) {

	res, err := DS.Dbx.NamedExec(messagesInsert, m)
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

func UpdateMessageSent(m *Message) error {
	if m.SendingError {
		log.Errorln("[axolotl] sending message failed ", m.SentAt)
	}
	_, err := DS.Dbx.NamedExec("UPDATE messages SET sentat = :sentat, sendingError = :sendingError, expireTimer = :expireTimer  WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}

func UpdateMessageRead(m *Message) error {
	_, err := DS.Dbx.NamedExec("UPDATE messages SET isread = :isread WHERE id = :id", m)
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
func LoadMessagesFromDB() error {
	err := LoadGroups()
	if err != nil {
		return err
	}
	log.Printf("Loading Messages")
	err = DS.Dbx.Select(&AllSessions, sessionsSelect)
	if err != nil {
		return err
	}
	for _, s := range AllSessions {
		s.When = helpers.HumanizeTimestamp(s.Timestamp)
		s.Active = !s.IsGroup || (Groups[s.Tel] != nil && Groups[s.Tel].Active)
		SessionsModel.Sess = append(SessionsModel.Sess, s)
		SessionsModel.Len++
		err = DS.Dbx.Select(&s.Messages, messagesSelectWhere, s.ID)
		s.Len = len(s.Messages)
		if err != nil {
			return err
		}
		for _, m := range s.Messages {
			m.HTime = helpers.HumanizeTimestamp(m.SentAt)
		}
	}
	//qml.Changed(SessionsModel, &SessionsModel.Len)
	return nil
}

// func LoadMessagesList(id int64) (error, *MessageList) {
// 	messageList := &MessageList{
// 		ID: id,
// 	}
// 	log.Printf("Loading Messages for " + string(id))
// 	err := DS.Dbx.Select(&messageList.Messages, messagesSelectWhere, id)
// 	if err != nil {
// 		return err, nil
// 	}
// 	return nil, messageList
// }
func DeleteMessage(id int64) error {
	_, err := DS.Dbx.Exec("DELETE FROM messages WHERE id = ?", id)
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
