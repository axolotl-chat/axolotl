package handler

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gen2brain/beeep"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/settings"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/nanu-c/axolotl/app/webserver"
	"github.com/signal-golang/textsecure"
)

// messageHandler is used on incoming message
func MessageHandler(msg *textsecure.Message) {
	buildAndSaveMessage(msg, false)
}
func buildAndSaveMessage(msg *textsecure.Message, syncMessage bool) {
	var err error
	var attachments []store.Attachment // should be array
	if len(msg.Attachments()) > 0 {
		for _, a := range msg.Attachments() {
			file, err := store.SaveAttachment(a)
			if err != nil {
				log.Printf("[axolotl] MessageHandler Error saving attachments %s\n", err.Error())
			}
			attachments = append(attachments, file)
		}
	}

	msgFlags := 0

	text := msg.Message()
	if msg.Flags() == textsecure.EndSessionFlag {
		text = "Secure session reset."
		msgFlags = helpers.MsgFlagResetSession
	}
	if msg.Flags() == 2 {
		text = "Message timer update."
		msgFlags = helpers.MsgFlagExpirationTimerUpdate
	}
	if msg.Flags() == uint32(textsecure.ProfileKeyUpdatedFlag) {
		msgFlags = helpers.MsgFlagProfileKeyUpdated
	}
	var session *store.SessionV2
	// GroupV2 Message
	var recipient *store.Recipient
	grV2 := msg.GroupV2()
	if grV2 != nil {
		group, err := store.GroupV2sModel.GetGroupById(grV2.Hexid)
		if err != nil {
			log.Println("[axolotl] GroupV2sModel.GetGroupById", err)
			return
		}
		if group == nil {
			// create new group
			name := "Unknown group"
			if grV2.DecryptedGroup != nil {
				name = string(grV2.DecryptedGroup.Title)
			}
			group, err = store.GroupV2sModel.Create(&store.GroupV2{
				Id:   grV2.Hexid,
				Name: name,
			})
			if err != nil {
				log.Println("[axolotl] GroupV2sModel.Create", err)
				return
			}
		}

		if grV2.GroupAction != nil && msg.Message() == "" {
			err = group.UpdateGroupAction(grV2.GroupAction)
			if err != nil {
				log.Println("[axolotl] GroupV2sModel.UpdateGroupAction", err)
			}
			text = "Group was updated"
			msgFlags = helpers.MsgFlagGroupV2Change
		}
		session, err = store.SessionsV2Model.GetSessionByGroupV2ID(grV2.Hexid)
		if err != nil {
			log.Println("[axolotl] SessionsV2Model.GetSessionByGroupV2ID", err)
		}
		// check if recipient exists and is in group
		recipient = store.RecipientsModel.GetRecipientByUUID(msg.SourceUUID())
		log.Debugln("[axolotl] GroupV2 Message ", msg.SourceUUID(), grV2.Hexid)
		if recipient == nil {
			log.Debugln("[axolotl] Recipient not found, creating new one for " + msg.SourceUUID())
			recipient, err = store.RecipientsModel.CreateRecipient(&store.Recipient{
				UUID: msg.SourceUUID(),
			})
			if err != nil {
				log.Errorln("[axolotl] RecipientsModel.Create", err)
				return
			}
		}
		if !group.IsMember(recipient) {
			err = group.AddMember(recipient)
			if err != nil {
				log.Errorln("[axolotl] GroupV2.AddMember", err)
				return
			}
		}
		if recipient.Username == "" {
			err = recipient.UpdateProfile()
			if err != nil {
				log.Errorln("[axolotl] Recipient.UpdateProfile", err)
			}
		}
	} else {
		recipient = store.RecipientsModel.GetRecipientByUUID(msg.SourceUUID())
		if recipient == nil {
			// todo get recipient profile from signal server
			recipient, err = store.RecipientsModel.CreateRecipient(&store.Recipient{
				UUID: msg.SourceUUID(),
			})
			if err != nil {
				log.Println("[axolotl] RecipientsModel.Create", err)
				return
			}
			session, err = store.SessionsV2Model.CreateSessionForDirectMessageRecipient(recipient.Id)
			if err != nil {
				log.Println("[axolotl] SessionsV2Model.CreateSessionForDirectMessageRecipient", err)
			}
		} else {
			session, err = store.SessionsV2Model.GetSessionByDirectMessageRecipientID(recipient.Id)
			if err != nil {
				log.Println("[axolotl] SessionsV2Model.GetSessionByRecipientID", err)
			}
		}
		if err != nil {
			log.Println("[axolotl] SessionsModel.GetSessionByID", err)
		}
		if recipient.Username == "" {
			recipient.UpdateProfile()
			if err != nil {
				log.Errorln("[axolotl] Recipient.UpdateProfile", err)
			}
		}
	}

	if msg.Sticker() != nil {
		msgFlags = helpers.MsgFlagSticker
		text = "Unsupported Message: sticker"
	}
	if msg.Contact() != nil {
		msgFlags = helpers.MsgFlagContact
		c := msg.Contact()[0]
		text = c.GetName().GetDisplayName() + " " + c.GetNumber()[0].GetValue()
	}
	if msg.Reaction() != nil {
		msgFlags = helpers.MsgFlagReaction
		text = msg.Reaction().GetEmoji()
	}
	mAttachment, ctype, err := prepareAttachment(attachments)
	if err != nil {
		log.Println("[axolotl] prepareAttachment", err)
		return
	}

	var m = &store.Message{
		Message:    text,
		Attachment: string(mAttachment),
		CType:      ctype,
		SID:        session.ID,
	}
	if syncMessage {
		m.Outgoing = true
		m.IsSent = true
	} else {
		m.Outgoing = false
		m.Source = msg.Source()
	}
	m.ReceivedAt = uint64(time.Now().UnixNano() / 1000000)
	m.SentAt = msg.Timestamp()
	m.SourceUUID = msg.SourceUUID()
	session.ExpireTimer = int64(msg.ExpireTimer())
	m.ExpireTimer = msg.ExpireTimer()
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	if msg.Quote() != nil {
		msgFlags = helpers.MsgFlagQuote
		text = msg.Message()
		err, id := store.FindQuotedMessage(msg.Quote())
		if err != nil || id == -1 {
			// create quoted message
			// TODO implement quoted message when not exists
			// quoteMessage := session.Add(text, msg.Quote().GetAuthorUuid(), nil, msg.Quote().GetText(), false, store.ActiveSessionID)
			// quoteMessage.Flags = helpers.MsgFlagHiddenQuote
			// err, savedQuoteMessage := store.SaveMessage(quoteMessage)
			// id = savedQuoteMessage.ID
			if err != nil {
				log.Debugln("[axolotl] Error saving quote message", err)
			}
		}
		m.QuoteID = id
	}
	if msgFlags != 0 {
		m.Flags = msgFlags
	}
	if msgFlags == helpers.MsgFlagProfileKeyUpdated {
		m.IsRead = true
	}
	if !session.IsMuted && !syncMessage && msgFlags != helpers.MsgFlagProfileKeyUpdated {
		if settings.SettingsModel.EncryptDatabase {
			text = "Encrypted message"
		}
		// only send a notification, when it's not the current chat
		if session.ID != store.ActiveSessionID {
			name, err := session.GetName()
			if err != nil {
				log.Println("[axolotl] session.getName", err)
			}
			var icon string
			if recipient != nil {
				avatar, err := textsecure.GetAvatarPath(recipient.UUID)
				if err == nil && avatar != "" {
					icon = avatar
				}
			} else {
				path := config.AxolotlWebDir
				axolotlWebDirEnv := os.Getenv("AXOLOTL_WEB_DIR")
				if len(axolotlWebDirEnv) > 0 {
					path = axolotlWebDirEnv
				}

				snapEnv := os.Getenv("SNAP")
				if len(snapEnv) > 0 && !strings.Contains(snapEnv, "/snap/go/") {
					path = os.Getenv("SNAP") + "/bin/axolotl-web/"
				}

				icon = path + "/public/axolotl.png"
			}
			if config.Gui == "ut" {
				n := push.Nh.NewStandardPushMessage(
					name,
					text, icon, strconv.FormatInt(session.ID, 10))
				push.Nh.Send(n)
			} else {

				err := beeep.Notify("Axolotl: "+name, text, icon)
				if err != nil {
					log.Errorln("[axolotl] notification ", err)
				}
			}
		}
	}
	// for now ignore empty messages and profile key updates
	if helpers.MsgFlagProfileKeyUpdated != msgFlags && (syncMessage || m.Message != "" || m.Attachment != "") {

		msgSend, err := store.SaveMessage(m)
		if err != nil {
			log.Printf("[axolotl] MessageHandler: Error saving message: %s\n", err.Error())
		}
		webserver.UpdateChatList()
		webserver.MessageHandler(msgSend)
	} else {
		log.Println("[axolotl] MessageHandler: Empty message")
	}
}
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
func CallMessageHandler(msg *textsecure.Message) {
	// TODO
	// log.Debugln("[axolotl] CallMessageHandler", msg)
	// session := store.SessionsV2Model.GetSessionBy(msg.SourceUUID())
	// var f []store.Attachment
	// m := session.Add(msg.Message(), "", f, "", true, store.ActiveSessionID)
	// store.SaveMessage(m)
	// webserver.UpdateChatList()
	// webserver.UpdateChatList()
}
func TypingMessageHandler(msg *textsecure.Message) {
	webserver.UpdateChatList()
}

// ReceiptHandler handles receipts for outgoing messages
func ReceiptHandler(source string, devID uint32, timestamp uint64) {

	m, err := store.FindOutgoingMessage(timestamp)
	if err != nil {
		log.Printf("[axolotl] ReceiptHandler: Message with timestamp %d not found\n", timestamp)
	} else {
		log.Println("[axolotl] ReceiptHandler for message ", timestamp, m.SourceUUID)
		m.IsSent = true
		m.Receipt = true
		store.UpdateMessageReceipt(m)
		webserver.UpdateMessageHandlerWithSource(m)
		return
	}
}

// ReceiptMessageHandler handles outgoing message receipts and marks messages as read
func ReceiptMessageHandler(msg *textsecure.Message) {
	log.Println("[axolotl] ReceiptMessageHandler for message ", msg.Timestamp())
	m, err := store.FindOutgoingMessage(msg.Timestamp())
	if err != nil {
		log.Printf("[axolotl] ReceiptMessageHandler: Message with timestamp %d not found\n", msg.Timestamp())
		return
	} else {
		if msg.Message() == "readReceiptMessage" {
			m.IsRead = true
			m.IsSent = true
			store.UpdateMessageRead(m)
		} else if msg.Message() == "deliveryReceiptMessage" {
			log.Debugln("[axolotl] unhandeld receipt message type for message ", msg.Timestamp(), msg.Message())
			m.IsSent = true
			store.UpdateMessageReceiptSent(m)
		}
		webserver.UpdateMessageHandlerWithSource(m)
		return
	}
}

// SyncSentHandler handle sync messages from signal desktop
func SyncSentHandler(msg *textsecure.Message, timestamp uint64) {
	log.Debugln("[axolotl] handle sync message", msg.Timestamp(), msg.SourceUUID())
	// use the same routine to save sync messages as incoming messages
	buildAndSaveMessage(msg, true)
}
