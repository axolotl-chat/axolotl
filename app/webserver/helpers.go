package webserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/settings"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
)

func websocketSender() {
	for {
		message := <-broadcast
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Errorln("[axolotl-ws] send message", err)
				removeClientFromList(client)
			}
		}
	}
}
func sendRegistrationStatus() {
	log.Debugln("[axolotl-ws] getRegistrationStatus")
	if requestUsername {
		sendRequest("getUsername")
	} else if registered {
		sendRequest("registrationDone")
	} else if requestPassword {
		sendRequest("getEncryptionPw")
	} else if requestSmsVerificationCode {
		sendRequest("getVerificationCode")
	} else {
		sendRequest("getPhoneNumber")
	}
}
func sendChatList() {
	var err error
	sessions, err := store.SessionsV2Model.GetAllSessions()
	if err != nil {
		log.Errorln("[axolotl] sendChatList1", err)
		return
	}
	lastMessages, err := store.SessionsV2Model.GetLastMessagesForAllSessions()
	if err != nil {
		log.Errorln("[axolotl] sendChatList2", err)
		return
	}
	sessionNames, err := store.SessionsV2Model.GetSessionNames()
	if err != nil {
		log.Errorln("[axolotl] sendChatList3", err)
		return
	}
	chatListEnvelope := &ChatListEnvelope{
		ChatList:     sessions,
		LastMessages: lastMessages,
		SessionNames: sessionNames,
	}
	message := &[]byte{}
	*message, err = json.Marshal(chatListEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendRegistrationStatus", err)
		return
	}
	broadcast <- *message
}
func sendCurrentChat(s *store.SessionV2) {
	var (
		err   error
		group *Group
	)
	if s.IsGroup() {
		gr, err := store.GroupV2sModel.GetGroupById(s.GroupV2ID)
		if err != nil {
			log.Errorln("[axolotl] sendCurrentChat: get group ", err)
			return
		}
		if gr == nil {
			log.Errorln("[axolotl] sendCurrentChat: group not found", s.GroupV2ID)
			return
		}
		members, err := gr.GetGroupMembersAsRecipients()
		if err != nil {
			log.Errorln("[axolotl] sendCurrentChat: get members ", err)
			return
		}
		group = &Group{
			HexId:      s.GroupV2ID,
			Name:       gr.Name,
			Members:    members,
			JoinStatus: gr.JoinStatus,
		}
	}
	currentChatEnvelope := &CurrentChatEnvelope{
		OpenChat: &OpenChat{
			CurrentChat: s,
			Contact:     &profile,
			Group:       group,
		},
	}
	message := &[]byte{}
	*message, err = json.Marshal(currentChatEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendCurrentChat: sendRegistrationStatus", err)
		return
	}
	broadcast <- *message
}
func updateCurrentChat(s *store.SessionV2) {
	var (
		err   error
		group *Group
		c     *store.Recipient
	)
	if s.IsGroup() {
		gr, err := store.GroupV2sModel.GetGroupById(s.GroupV2ID)
		if err != nil {
			log.Errorln("[axolotl] updateCurrentChat: ", err)
			return
		}
		if gr == nil {
			log.Errorln("[axolotl] updateCurrentChat: group not found", s.GroupV2ID)
			return
		}
		members, err := gr.GetGroupMembersAsRecipients()
		if err != nil {
			log.Errorln("[axolotl] updateCurrentChat: ", err)
			return
		}
		group = &Group{
			HexId:   gr.Id,
			Name:    gr.Name,
			Members: members,
		}
	} else {
		c = store.RecipientsModel.GetRecipientById(s.DirectMessageRecipientID)
	}
	updateCurrentChatEnvelope := &UpdateCurrentChatEnvelope{
		UpdateCurrentChat: &UpdateCurrentChat{
			CurrentChat: s,
			Contact:     c,
			Group:       group,
		},
	}
	message := &[]byte{}
	*message, err = json.Marshal(updateCurrentChatEnvelope)
	if err != nil {
		log.Errorln("[axolotl] updateCurrentChat", err)
		return
	}
	broadcast <- *message
}
func refreshContacts(path string) {
	var err error
	config.VcardPath = path
	contact.GetAddressBookContactsFromContentHub()
	err = store.RefreshContacts()
	if err != nil {
		ShowError(err.Error())
	}
	go sendContactList()
}
func recoverFromWsPanic(client *websocket.Conn) {
	client.Close()
	removeClientFromList(client)

}
func sendContactList() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorln("[axolotl] sendContactList panic occurred:", err)
		}
	}()
	var err error
	contactListEnvelope := &ContactListEnvelope{
		ContactList: store.ContactsModel.Contacts,
	}
	message := &[]byte{}
	*message, err = json.Marshal(contactListEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendContactList ", err)
		return
	}
	broadcast <- *message
}
func sendDeviceList() {
	var err error
	devices, err := textsecure.LinkedDevices()
	if err != nil {
		log.Errorln("[axolotl] sendDeviceList", err)
		return
	}
	deviceListEnvelope := &DeviceListEnvelope{
		DeviceList: devices,
	}
	message := &[]byte{}
	*message, err = json.Marshal(deviceListEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendDeviceList", err)
		return
	}
	broadcast <- *message
}
func createDirectRecipientChat(uuid string) (*store.SessionV2, error) {
	var err error
	if uuid == "0" {
		return nil, fmt.Errorf("createDirectRecipientChat: uuid is 0")
	}
	contact := store.GetContactForUUID(uuid)

	recipient := store.RecipientsModel.GetOrCreateRecipientForContact(contact)
	if recipient == nil {
		recipient, err = store.RecipientsModel.CreateRecipient(&store.Recipient{
			UUID: uuid,
		})
		if err != nil {
			log.Errorln("[axolotl] createDirectRecipientChat", err)
			return nil, err
		}
	}
	session, err := store.SessionsV2Model.GetOrCreateSessionForDirectMessageRecipient(recipient.Id)
	if err != nil {
		log.Errorln("[axolotl] createDirectRecipientChat", err)
		return nil, err
	}
	// ensure that a message exists
	messages, err := session.GetMessageList(1, 0)
	if err != nil || len(messages) == 0 {
		m := &store.Message{Message: "New chat created",
			SID:         session.ID,
			Outgoing:    true,
			Source:      "",
			SourceUUID:  config.Config.UUID,
			HTime:       "Now",
			SentAt:      uint64(time.Now().UnixNano() / 1000000),
			ExpireTimer: uint32(session.ExpireTimer),
		}
		go MessageHandler(m)
	}
	return session, nil
}

// createGroup creates a group chat session and returns the session id. -> Deprectated
func createGroup(newGroupData CreateGroupMessage) (*store.SessionV2, error) {
	group, err := textsecure.NewGroup(newGroupData.Name, newGroupData.Members)
	if err != nil {
		ShowError(err.Error())
		return nil, err
	}
	members := strings.Join(newGroupData.Members, ",")
	if !strings.Contains(members, config.Config.Tel) {
		members = members + "," + config.Config.Tel
	}
	store.Groups[group.Hexid] = &store.GroupRecord{
		GroupID: group.Hexid,
		Name:    newGroupData.Name,
		Members: members,
	}
	_, err = store.SaveGroup(store.Groups[group.Hexid])
	if err != nil {
		ShowError(err.Error())
		return nil, err
	}
	session, err := store.SessionsV2Model.CreateSessionForGroupV1(group.Hexid)
	if err != nil {
		ShowError(err.Error())
		return nil, err
	}
	// TODO:
	// msg := session.Add(store.GroupUpdateMsg(append(newGroupData.Members, config.Config.Tel), newGroupData.Name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
	// msg.Flags = helpers.MsgFlagGroupNew
	// store.SaveMessage(msg)
	return session, nil
}
func joinGroupV2(joinGroupmessage JoinGroupMessage) *store.SessionV2 {
	log.Infoln("[axolotl] joinGroup", joinGroupmessage.ID)
	group, err := textsecure.JoinGroup(joinGroupmessage.ID)
	if err != nil {
		log.Warnln("[axolotl] error while joining group", err)
		if group == nil {
			log.Errorln("[axolotl] joinGroup failed")
			return nil
		}
		// in this case the group is already joined. Its join status has to be updated
	}
	// members cannot be read if the group is not yet joined
	if group.JoinStatus == store.GroupJoinStatusJoined {
		log.Infoln("[axolotl] joining group was successful", group.Hexid)
		for _, member := range group.DecryptedGroup.Members {
			// todo: add recipients and store them as group members
			log.Debugln("[axolotl] member", member)
		}
	}
	groupV2, err := store.GroupV2sModel.GetGroupById(group.Hexid)
	if err != nil {
		log.Errorln("[axolotl] joinGroup", err)
		return nil
	}
	// create groupV2 if it doesn't exist
	if groupV2 == nil {
		groupV2, err = store.GroupV2sModel.Create(&store.GroupV2{
			Id:         group.Hexid,
			Name:       group.DecryptedGroup.Title,
			JoinStatus: group.JoinStatus,
		})
		if err != nil {
			log.Errorln("[axolotl] joinGroup", err)
			return nil
		}
	} else {
		// update GroupV2
		groupV2.JoinStatus = group.JoinStatus
		groupV2.Name = group.DecryptedGroup.Title
		err = groupV2.UpdateGroup()
		if err != nil {
			log.Errorln("[axolotl] joinGroup", err)
			return nil
		}
	}
	// Create or get session for group
	session, _ := store.SessionsV2Model.GetSessionByGroupV2ID(group.Hexid)
	if session == nil {
		session, _ = store.SessionsV2Model.CreateSessionForGroupV2(groupV2.Id)
	}

	// Add a join message to the session when the group is joined
	if group.JoinStatus == store.GroupJoinStatusJoined {
		msg, err := store.SaveMessage(&store.Message{
			Message: "You accepted the invitation to the group.",
			Flags:   helpers.MsgFlagGroupJoined,
			SID:     session.ID,
		})
		if err != nil {
			log.Errorln("[axolotl] joinGroup", err)
			return nil
		}
		MessageHandler(msg)
	}
	requestEnterChat(store.ActiveSessionID)
	sendContactList()
	return session
}
func sendMessageList(ID int64) {
	message := &[]byte{}
	log.Debugln("[axolotl] sendMessageList for conversation", ID)
	session, err := store.SessionsV2Model.GetSessionByID(ID)
	if err != nil {
		log.Errorln("[axolotl] sendMessageList", err)
		return
	}
	messageList, err := session.GetMessageList(20, 0)
	if err != nil {
		log.Errorln("[axolotl] sendMessageList: ", err)
		return
	}
	session.MarkRead()
	if push.Nh != nil {
		push.Nh.Clear(session.ID)
	}
	chatListEnvelope := &MessageListEnvelope{
		MessageList: messageList,
	}
	*message, err = json.Marshal(chatListEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendMessageList: ", err)
		return
	}
	broadcast <- *message
}
func sendMoreMessageList(id int64, sentat uint64) {
	message := &[]byte{}
	err, messageList := store.SessionsV2Model.GetMoreMessageList(id, sentat)
	if err != nil {
		log.Errorln("[axolotl] sendMoreMessageList: ", err)
		return
	}
	moreMessageListEnvelope := &MoreMessageListEnvelope{
		MoreMessageList: messageList,
	}
	*message, err = json.Marshal(moreMessageListEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendMoreMessageList: ", err)
		return
	}
	broadcast <- *message
}
func sendIdentityInfo(fingerprintNumbers []string, fingerprintQRCode []byte) {
	var err error
	r := make([]int, 0)
	for _, i := range fingerprintQRCode {
		r = append(r, int(i))
	}
	message := &[]byte{}
	identityEnvelope := &IdentityEnvelope{
		FingerprintNumbers: fingerprintNumbers,
		FingerprintQRCode:  r,
	}
	*message, err = json.Marshal(identityEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendIdentityInfo: ", err)
		return
	}
	broadcast <- *message
}

// UpdateChatList updates the chatlist if not entered a chat + registered
func UpdateChatList() {

	if activeChat == -1 && registered {
		sendChatList()
	}
}

// UpdateContactList updates the contactlist if not entered a chat + registered
func UpdateContactList() {

	if activeChat == -1 && registered {
		sendContactList()
	}
}

// UpdateActiveChat checks if there is an active chat an if yes it updates it on axolotl web
func UpdateActiveChat() {
	if activeChat != -1 {
		// log.Debugln("[axolotl] update activ	e chat")
		s, err := store.SessionsV2Model.GetSessionByID(activeChat)
		if err != nil {
			log.Errorln("[axolotl-ws] UpdateActiveChat", err)
		} else {
			updateCurrentChat(s)
		}
	}
}

// SendGui sends to axolotl the gui to trigger ubuntu touch specific parts
type SendGui struct {
	Gui string
}

// SetGui sets the gui
func SetGui() {
	var err error
	gui := config.Gui
	if gui == "" {
		gui = "standard"
	}
	request := &SendGui{
		Gui: gui,
	}
	message := &[]byte{}
	*message, err = json.Marshal(request)
	if err != nil {
		log.Errorln("[axolotl] set gui", err)
		return
	}
	broadcast <- *message
	// RemoveClientFromList(client)
}

type sendDarkmode struct {
	DarkMode bool
}

func SetUiDarkMode() {
	log.Debugln("[axolotl] send darkmode to client", settings.SettingsModel.DarkMode)

	var err error
	mode := settings.SettingsModel.DarkMode

	request := &sendDarkmode{
		DarkMode: mode,
	}
	message := &[]byte{}
	*message, err = json.Marshal(request)
	if err != nil {
		log.Errorln("[axolotl] set darkmode", err)
		return
	}
	broadcast <- *message
}
func sendConfig() {
	var err error

	message := &[]byte{}
	configEnvelope := &ConfigEnvelope{
		Type:             "config",
		RegisteredNumber: config.Config.Tel,
		Version:          config.AppVersion,
		Gui:              config.Gui,
		LogLevel:         config.Config.LogLevel,
	}
	*message, err = json.Marshal(configEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendConfig", err)
		return
	}
	broadcast <- *message
}

func importVcf() {
	config.VcardPath = "import.vcf"
	contact.GetAddressBookContactsFromContentHub()
	err := store.RefreshContacts()
	if err != nil {
		ShowError(err.Error())
	}
	sendContactList()
}
