package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gen2brain/beeep"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/nanu-c/axolotl/app/webserver"
	"github.com/signal-golang/textsecure"
)

//messageHandler is used on incoming message
func MessageHandler(msg *textsecure.Message, wsApp *webserver.WsApp) {
	buildAndSaveMessage(msg, false, wsApp)
}
func buildAndSaveMessage(msg *textsecure.Message, syncMessage bool, wsApp *webserver.WsApp) {
	var err error
	var attachments []store.Attachment //should be array
	mt := ""                           //
	if len(msg.Attachments()) > 0 {
		for i, a := range msg.Attachments() {
			mt = msg.Attachments()[i].MimeType
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
	//Group Message
	gr := msg.Group()

	if gr != nil && gr.Flags != 0 || gr != nil && gr.Name != store.Groups[gr.Hexid].Name {
		_, ok := store.Groups[gr.Hexid]
		members := ""
		if ok {
			members = store.Groups[gr.Hexid].Members
			if store.Groups[gr.Hexid].Name == gr.Hexid {
				textsecure.RemoveGroupKey(gr.Hexid)
				textsecure.RequestGroupInfo(gr)
			}
		}
		av := []byte{}

		if gr.Avatar != nil {
			av, err = ioutil.ReadAll(bytes.NewReader(gr.Avatar))
			if err != nil {
				log.Println("[axolotl] avatar", err)
				return
			}
		}
		store.Groups[gr.Hexid] = &store.GroupRecord{
			GroupID: gr.Hexid,
			Members: strings.Join(gr.Members, ","),
			Name:    gr.Name,
			Avatar:  av,
			Active:  true,
		}
		if ok {
			store.UpdateGroup(store.Groups[gr.Hexid])
		} else {
			store.SaveGroup(store.Groups[gr.Hexid])
		}

		if gr.Flags == textsecure.GroupUpdateFlag {
			dm, _ := helpers.MembersDiffAndUnion(members, strings.Join(gr.Members, ","))
			text = store.GroupUpdateMsg(dm, gr.Name)
			msgFlags = helpers.MsgFlagGroupUpdate
		}
		if gr.Flags == textsecure.GroupLeaveFlag {
			text = store.TelToName(msg.Source(), wsApp.App.Config.TextsecureConfig.Tel) + " has left the group."
			msgFlags = helpers.MsgFlagGroupLeave
		}
	}
	//GroupV2 Message
	grV2 := msg.GroupV2()
	if grV2 != nil {
		group := store.Groups[grV2.Hexid]
		if group != nil && grV2.DecryptedGroup != nil {
			group.Name = string(grV2.DecryptedGroup.Title)
			store.UpdateGroup(group)
		} else {
			title := "Unknown group"
			if grV2.DecryptedGroup != nil {
				title = string(grV2.DecryptedGroup.Title)
			}
			store.Groups[grV2.Hexid] = &store.GroupRecord{
				GroupID: grV2.Hexid,
				Name:    title,
				Type:    store.GroupRecordTypeGroupv2,
			}
			_, err = store.SaveGroup(store.Groups[grV2.Hexid])
			if err != nil {
				log.Errorln("[axolotl] save groupV2 ", err)
			}
		}
		if grV2.GroupAction != nil && msg.Message() == "" {
			text = "Group was changed to revision " + fmt.Sprint(grV2.GroupContext.Revision)
			msgFlags = helpers.MsgFlagGroupV2Change
		}
		// handle groupv2 updates etc
	}

	msgSource := msg.SourceUUID()
	if gr != nil {
		msgSource = gr.Hexid
	}
	if grV2 != nil {
		msgSource = grV2.Hexid
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
	session, err := store.SessionsModel.GetByUUID(msgSource)
	if gr != nil {
		if err != nil {
			log.Println("[axolotl] MessageHandler error finding group session by uuid", err)
			session = store.SessionsModel.GetByE164(msgSource)
			if session != nil {
				log.Println("[axolotl] MessageHandler update group session uuid")
				session.UUID = session.Tel
				store.UpdateSession(session)
				err = nil
			}
		}
	}
	// deduplicate sessions fix bug in 1.9.4 could be deleted later
	sessions := store.SessionsModel.GetAllSessionsByE164(msgSource)
	if len(sessions) > 1 {
		if len(sessions[0].UUID) < 32 {
			store.MigrateMessagesFromSessionToAnotherSession(sessions[0].ID, sessions[1].ID)
		} else {
			store.MigrateMessagesFromSessionToAnotherSession(sessions[1].ID, sessions[0].ID)
		}
		session, err = store.SessionsModel.GetByUUID(msgSource)
		wsApp.UpdateChatList()
	}
	if err != nil && gr == nil && grV2 == nil {
		// Session could not be found, lets try to find it by E164 aka phone number
		log.Println("[axolotl] MessageHandler: ", err)
		session = store.SessionsModel.GetByE164(msg.Source())
		if session != nil {
			// add uuid to session
			log.Println("[axolotl] Update Session to new uuid for tel", msg.Source())
			session.UUID = msgSource
			err := store.UpdateSession(session)
			if err != nil {
				log.Debugln("[axolotl] Error update Session to new uuid", err)
			}
		} else {
			// create a new session
			session = store.SessionsModel.CreateSessionForE164(msg.Source(), msg.SourceUUID())
		}
	} else if err != nil && gr != nil {
		log.Infoln("[axolotl] MessageHandler group Error ", err)
		session = store.SessionsModel.CreateSessionForGroup(gr)
		// TODO create group
	} else if err != nil && grV2 != nil {
		log.Infoln("[axolotl] MessageHandler group2 not found, lets create it ", err)
		session = store.SessionsModel.CreateSessionForGroupV2(grV2)
		// TODO create group
	}
	var m *store.Message
	if syncMessage {
		m = session.Add(text, "", attachments, mt, true, store.ActiveSessionID)
		m.IsSent = true
	} else {
		m = session.Add(text, msg.Source(), attachments, mt, false, store.ActiveSessionID)
	}
	m.ReceivedAt = uint64(time.Now().UnixNano() / 1000000)
	m.SentAt = msg.Timestamp()
	m.SourceUUID = msg.SourceUUID()
	session.ExpireTimer = msg.ExpireTimer()
	store.UpdateSession(session)
	m.ExpireTimer = msg.ExpireTimer()
	m.HTime = helpers.HumanizeTimestamp(m.SentAt)
	if msg.Quote() != nil {
		msgFlags = helpers.MsgFlagQuote
		text = msg.Message()
		err, id := store.FindQuotedMessage(msg.Quote())
		if err != nil || id == -1 {
			// create quoted message
			quoteMessage := session.Add(text, msg.Quote().GetAuthorE164(), nil, msg.Quote().GetText(), false, store.ActiveSessionID)
			quoteMessage.Flags = helpers.MsgFlagHiddenQuote
			err, savedQuoteMessage := store.SaveMessage(quoteMessage)
			id = savedQuoteMessage.ID
			if err != nil {
				log.Debugln("[axolotl] Error saving quote message", err)
			}
		}
		m.QuoteID = id
	}
	session.Timestamp = m.SentAt
	session.When = m.HTime
	if gr != nil && gr.Flags == textsecure.GroupUpdateFlag {
		session.Name = gr.Name
	}
	if msgFlags != 0 {
		m.Flags = msgFlags
	}
	if msgFlags == helpers.MsgFlagProfileKeyUpdated {
		m.IsRead = true
	}
	if session.Notification && !syncMessage && msgFlags != helpers.MsgFlagProfileKeyUpdated {
		if wsApp.App.Settings.EncryptDatabase {
			text = "Encrypted message"
		}
		//only send a notification, when it's not the current chat
		// if session.ID != store.Sessions.GetActiveChat {
		if session.ID != store.ActiveSessionID {
			if wsApp.App.Config.Gui == "ut" {
				n := push.Nh.NewStandardPushMessage(
					session.Name,
					text, "", msgSource)
				push.Nh.Send(n)
			} else {
				err := beeep.Notify("Axolotl: "+session.Name, text, "axolotl-web/dist/public/axolotl.png")
				if err != nil {
					log.Errorln("[axolotl] notification ", err)
				}
			}
		}
	}
	err, msgSend := store.SaveMessage(m)
	if err != nil {
		log.Printf("[axolotl] MessageHandler: Error saving message: %s\n", err.Error())
	}
	store.UpdateSession(session)
	// webserver.UpdateChatList()
	wsApp.MessageHandler(msgSend)
}
func CallMessageHandler(msg *textsecure.Message, wsApp *webserver.WsApp) {
	log.Debugln("[axolotl] CallMessageHandler", msg)
	session := store.SessionsModel.GetByE164(msg.Source())
	var f []store.Attachment
	m := session.Add(msg.Message(), "", f, "", true, store.ActiveSessionID)
	store.SaveMessage(m)
	wsApp.UpdateChatList()
	wsApp.UpdateChatList()
}
func TypingMessageHandler(msg *textsecure.Message, wsApp *webserver.WsApp) {
	wsApp.UpdateChatList()
}

// ReceiptHandler handles receipts for outgoing messages
func ReceiptHandler(source string, devID uint32, timestamp uint64, wsApp *webserver.WsApp) {
	m, err := store.FindOutgoingMessage(timestamp)
	if err != nil {
		log.Printf("[axolotl] ReceiptHandler: Message with timestamp %d not found\n", timestamp)
	} else {
		log.Println("[axolotl] ReceiptHandler for message ", timestamp, m.SourceUUID)
		m.IsSent = true
		m.Receipt = true
		store.UpdateMessageReceipt(m)
		wsApp.UpdateMessageHandlerWithSource(m)
		return
	}
}

// ReceiptMessageHandler handles outgoing message receipts and marks messages as read
func ReceiptMessageHandler(msg *textsecure.Message, wsApp *webserver.WsApp) {
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
		wsApp.UpdateMessageHandlerWithSource(m)
		return
	}
}

// SyncSentHandler handle sync messages from signal desktop
func SyncSentHandler(msg *textsecure.Message, timestamp uint64, wsApp *webserver.WsApp) {
	log.Debugln("[axolotl] handle sync message", msg.Timestamp(), msg.SourceUUID())
	// use the same routine to save sync messages as incoming messages
	buildAndSaveMessage(msg, true, wsApp)
}
