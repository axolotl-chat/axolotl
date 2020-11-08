package sender

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
)

func SendMessageHelper(to, message, file string, updateMessageChannel chan *store.Message) (error, *store.Message) {
	if to != "" {
		var err error
		attachments := []store.Attachment{}
		if file != "" {
			file, err = store.CopyAttachment(file)
			log.Debugln("[axolotl] attachment: " + file)
			if err != nil {
				log.Errorln("[axolotl] Error Attachment:" + err.Error())
				return err, nil
			}
			strParts := strings.Split(file, "/")
			attachments = []store.Attachment{store.Attachment{File: file, FileName: strParts[len(strParts)-1]}}

		}
		session := store.SessionsModel.Get(to)

		m := session.Add(message, "", attachments, "", true, store.ActiveSessionID)
		m.Source = to
		m.ExpireTimer = session.ExpireTimer
		_, savedM := store.SaveMessage(m)

		go func() {
			mNew, _ := SendMessage(session, m)
			if updateMessageChannel != nil {
				updateMessageChannel <- mNew
			}
		}()
		return nil, savedM
	} else {
		log.Errorln("[axolotl] send to is empty")
		return errors.New("send to is empty"), nil
	}
}

func SendMessage(s *store.Session, m *store.Message) (*store.Message, error) {
	var att io.Reader
	var err error
	if len(m.Attachment) > 0 && m.Attachment != "null" {

		files := []store.Attachment{}
		json.Unmarshal([]byte(m.Attachment), &files)
		att, err = os.Open(files[0].File)
		if err != nil {
			log.Errorln("[axolotl] SendMessage FileOpend")
			return nil, err
		} else {
			log.Printf("[axolotl] SendMessage FileOpend")
		}
	}
	ts := SendMessageLoop(s.Tel, m.Message, s.IsGroup, att, m.Flags, s.ExpireTimer)
	log.Debugln("[axolotl] SendMessage", s.Tel, ts)
	m.SentAt = ts
	m.ExpireTimer = s.ExpireTimer
	s.Timestamp = m.SentAt
	m.IsSent = true
	if ts == 0 {
		m.SendingError = true
		m.IsSent = false
		m.SentAt = uint64(time.Now().UnixNano() / 1000000)
		m.ExpireTimer = 0
	}
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	s.When = m.HTime
	store.UpdateMessageSent(m)
	store.UpdateSession(s)
	if m.SendingError {
		return m, errors.New("message sending failed")
	}
	return m, nil
}

// SendMessageLoop sends a single message and also loops over groups in order to send it to each participant of the group
func SendMessageLoop(to, message string, group bool, att io.Reader, flags int, timer uint32) uint64 {
	var err error
	var ts uint64
	var count int
	for {
		err = nil
		if flags == helpers.MsgFlagResetSession {
			ts, err = textsecure.EndSession(to, "TERMINATE")
		} else if flags == helpers.MsgFlagGroupLeave {
			err = textsecure.LeaveGroup(to)
		} else if flags == helpers.MsgFlagGroupUpdate {
			_, err = textsecure.UpdateGroup(to, store.Groups[to].Name, strings.Split(store.Groups[to].Members, ","))
		} else if att == nil {
			if group {
				ts, err = textsecure.SendGroupMessage(to, message, timer)
				if err != nil {
					log.Errorln("[axolotl] send to group ", err)
				}
				log.Debugln("[axolotl] send to group ")
			} else {
				ts, err = textsecure.SendMessage(to, message, timer)
				if err != nil {
					log.Errorln("[axolotl] send message error", err.Error(), ts)
				}

			}
		} else {
			if group {
				ts, err = textsecure.SendGroupAttachment(to, message, att, timer)
			} else {
				log.Printf("[axolotl] SendMessageLoop sendAttachment")
				ts, err = textsecure.SendAttachment(to, message, att, timer)
			}
		}
		if err == nil {
			break
		}
		log.Println("[axolotl]", err)
		//If sending failed, try again after a while
		time.Sleep(3 * time.Second)
		count++
		if count == 2 {
			// return nil, new Error("sending")
			break

		}
	}
	return ts
}
func SendUnsentMessages() {
	// for _, s := range store.SessionsModel.Sess {
	// 	for _, m := range s.Messages {
	// 		if m.Outgoing && !m.IsSent {
	// 			go SendMessage(s, m)
	// 		}
	// 	}
	// }
}
