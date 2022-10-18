package sender

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
)

// SendMessageHelper sends the message and returns the updated message
func SendMessageHelper(ID int64, message, file string,
	updateMessageChannel chan *store.Message,
	voiceMessage bool) (*store.Message, error) {
	if ID >= 0 {
		var err error
		attachments := []store.Attachment{}
		if file != "" {
			file, err = store.CopyAttachment(file)
			log.Debugln("[axolotl] attachment: " + file)
			if err != nil {
				log.Errorln("[axolotl] Error Attachment:" + err.Error())
				return nil, err
			}
			strParts := strings.Split(file, "/")
			filename := strParts[len(strParts)-1]
			if voiceMessage {
				attachments = []store.Attachment{
					store.Attachment{
						File:     file,
						CType:    3,
						FileName: filename,
					},
				}
			} else {
				attachments = []store.Attachment{
					store.Attachment{
						File:     file,
						FileName: filename,
					},
				}
			}

		}
		session, err := store.SessionsV2Model.GetSessionByID(ID)
		if err != nil {
			log.Errorln("[axolotl] SendMessageHelper: get session " + err.Error())
			return nil, err
		}
		fJSON, ctype, err := prepareAttachment(attachments)
		if err != nil {
			log.Errorln("[axolotl] SendMessageHelper: attachment " + err.Error())
			return nil, err
		}
		// m := session.Add(message, "", attachments, "", true, ID)
		m := &store.Message{Message: message,
			SID:         session.ID,
			Outgoing:    true,
			Source:      "",
			SourceUUID:  config.Config.UUID,
			CType:       ctype,
			Attachment:  string(fJSON),
			HTime:       "Now",
			SentAt:      uint64(time.Now().UnixNano() / 1000000),
			ExpireTimer: uint32(session.ExpireTimer),
		}
		log.Debugln("[axolotl] SendMessageHelper sentAt ", m.SentAt)
		savedM, err := store.SaveMessage(m)
		if err != nil {
			log.Errorln("[axolotl] SendMessageHelper: save Message" + err.Error())
			return nil, err
		}

		go func() {
			mNew, err := SendMessage(session, m, voiceMessage)
			if err != nil {
				log.Errorln("[axolotl] SendMessageHelper: send message" + err.Error())
				mNew.IsSent = false
				mNew.SendingError = true
				mNew, err = store.SaveMessage(mNew)
				if err != nil {
					log.Errorln("[axolotl] SendMessageHelper: save message" + err.Error())
					return
				}

			}
			if updateMessageChannel != nil {
				updateMessageChannel <- mNew
			}
		}()
		return savedM, nil
	}
	log.Errorln("[axolotl] send to is empty")
	return nil, errors.New("send to is empty")
}

// prepareAttachment prepares the attachment for sending
func prepareAttachment(file []store.Attachment) ([]byte, int, error) {
	var files []store.Attachment

	ctype := helpers.ContentTypeMessage
	if len(file) > 0 {
		for _, fi := range file {
			f, _ := os.Open(fi.File)
			if fi.CType == 0 {
				ctype = helpers.ContentType(f, "")
			} else {
				ctype = fi.CType
			}
			files = append(files, store.Attachment{File: fi.File, CType: ctype, FileName: fi.FileName})
		}
	}
	fJson, err := json.Marshal(files)
	return fJson, ctype, err

}

func SendMessage(s *store.SessionV2, m *store.Message, isVoiceMessage bool) (*store.Message, error) {
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
	var recipient string
	// check if user uuid exists and if yes send it to the uuid instead of the phone number
	if s.IsGroup() {
		recipient = s.GroupV2ID
	} else {
		r := store.RecipientsModel.GetRecipientById(s.DirectMessageRecipientID)
		if r == nil {
			return nil, errors.New("recipient not found")
		}
		if r.UUID != "" {
			recipient = r.UUID
		} else {
			recipient = r.E164
		}
	}
	ts := SendMessageLoop(recipient, m.Message, s.IsGroup(), att, m.Flags, uint32(s.ExpireTimer), isVoiceMessage)
	log.Debugln("[axolotl] SendMessage", recipient, ts)
	m.SentAt = ts
	m.IsSent = true
	if ts == 0 {
		m.SendingError = true
		m.IsSent = false
		m.SentAt = uint64(time.Now().UnixNano() / 1000000)
		m.ExpireTimer = 0
	}
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	store.UpdateMessageSent(m)
	if m.SendingError {
		return m, errors.New("message sending failed")
	}
	return m, nil
}

// SendMessageLoop sends a single message and also loops over groups in order to send it to each participant of the group
func SendMessageLoop(to, message string, group bool, att io.Reader, flags int, timer uint32, isVoiceMessage bool) uint64 {
	var err error
	var ts uint64
	var count int
	for {
		err = nil
		if flags == helpers.MsgFlagResetSession {
			ts, err = textsecure.EndSession(to, "TERMINATE")
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
				log.Debugln("[axolotl] send message sent", ts)

			}
		} else {
			if group {
				if isVoiceMessage {
					ts, err = textsecure.SendGroupVoiceNote(to, message, att, timer)
				} else {
					ts, err = textsecure.SendGroupAttachment(to, message, att, timer)
				}
			} else {
				if isVoiceMessage {
					ts, err = textsecure.SendVoiceNote(to, message, att, timer)
					log.Printf("[axolotl] SendMessageLoop send voice note")
				} else {
					log.Printf("[axolotl] SendMessageLoop sendAttachment")
					ts, err = textsecure.SendAttachment(to, message, att, timer)
				}
			}
		}
		if err == nil {
			break
		}
		log.Println("[axolotl]", err)
		// If sending failed, try again after a while
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
	// for _, s := range store.SessionsV2Model.Sess {
	// 	for _, m := range s.Messages {
	// 		if m.Outgoing && !m.IsSent {
	// 			go SendMessage(s, m)
	// 		}
	// 	}
	// }
}
