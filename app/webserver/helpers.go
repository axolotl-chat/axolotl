package webserver

import (
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/settings"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
	textsecureContacts "github.com/signal-golang/textsecure/contacts"
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
	chatListEnvelope := &ChatListEnvelope{
		ChatList: store.SessionsModel.Sess,
	}
	message := &[]byte{}
	*message, err = json.Marshal(chatListEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendRegistrationStatus", err)
		return
	}
	broadcast <- *message
}
func sendCurrentChat(s *store.Session) {
	var (
		err   error
		gr    *store.GroupRecord
		group *Group
	)
	if s.IsGroup {
		gr = store.GetGroupById(s.UUID)
		if gr == nil {
			log.Errorln("[axolotl] sendCurrentChat: group not found", s.UUID)
			return
		}
		group = &Group{
			HexId:   gr.GroupID,
			Name:    gr.Name,
			Members: strings.Split(gr.Members, ","),
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
func updateCurrentChat(s *store.Session) {
	var (
		err   error
		gr    *store.GroupRecord
		group *Group
		c     *textsecureContacts.Contact
	)
	if s.IsGroup {
		gr = store.GetGroupById(s.UUID)
		if gr == nil {
			log.Errorln("[axolotl] updateCurrentChat: group not found", s.UUID)
			return
		}
		group = &Group{
			HexId:   gr.GroupID,
			Name:    gr.Name,
			Members: strings.Split(gr.Members, ","),
		}
	} else {
		c = store.GetContactForTel(s.Tel)
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
func createChat(uuid string) *store.Session {
	if uuid == "0" {
		return nil
	}
	session, err := store.SessionsModel.GetByUUID(uuid)
	if err != nil {
		session = store.SessionsModel.CreateSessionForUUID(uuid)
	}
	return session
}
func createGroup(newGroupData CreateGroupMessage) (*store.Session, error) {
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
	session := store.SessionsModel.CreateSessionForGroup(group)
	msg := session.Add(store.GroupUpdateMsg(append(newGroupData.Members, config.Config.Tel), newGroupData.Name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
	msg.Flags = helpers.MsgFlagGroupNew
	store.SaveMessage(msg)
	return session, nil
}
func updateGroup(updateGroupData UpdateGroupMessage) *store.Session {
	group, err := textsecure.UpdateGroup(updateGroupData.ID, updateGroupData.Name, updateGroupData.Members)
	if err != nil {
		ShowError(err.Error())
		return nil
	}
	members := strings.Join(updateGroupData.Members, ",")
	store.Groups[updateGroupData.ID] = &store.GroupRecord{
		GroupID: updateGroupData.ID,
		Name:    updateGroupData.Name,
		Members: members,
	}
	store.SaveGroup(store.Groups[group.Hexid])
	session, err := store.SessionsModel.GetByUUID(group.Hexid)
	if err != nil {
		ShowError(err.Error())
	} else {
		msg := session.Add(store.GroupUpdateMsg(updateGroupData.Members, updateGroupData.Name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
		msg.Flags = helpers.MsgFlagGroupUpdate
		store.SaveMessage(msg)

		return session
	}
	return nil

}
func joinGroup(joinGroupmessage JoinGroupMessage) *store.Session {
	log.Infoln("[axolotl] joinGroup", joinGroupmessage.ID)
	group, err := textsecure.JoinGroup(joinGroupmessage.ID)
	if err != nil {
		if group != nil {
			session, _ := store.SessionsModel.GetByUUID(group.Hexid)
			session.Name = group.DecryptedGroup.Title
			session.GroupJoinStatus = group.JoinStatus
			store.UpdateSession(session)
		}
		log.Errorln("[axolotl] joinGroup", err)
		ShowError(err.Error())
		return nil
	}
	log.Infoln("[axolotl] joining group was succesful", group.Hexid)
	store.Groups[group.Hexid] = &store.GroupRecord{
		GroupID: group.Hexid,
		Name:    group.DecryptedGroup.Title,
		// Members: members,
		JoinStatus: store.GroupJoinStatusJoined,
	}
	store.SaveGroup(store.Groups[group.Hexid])
	session, err := store.SessionsModel.GetByUUID(group.Hexid)
	session.Name = group.DecryptedGroup.Title
	session.GroupJoinStatus = store.GroupJoinStatusJoined
	msg := session.Add("You accepted the invitation to the group.", "", []store.Attachment{}, "", true, store.ActiveSessionID)
	msg.Flags = helpers.MsgFlagGroupJoined
	store.UpdateSession(session)
	store.SaveMessage(msg)
	requestEnterChat(store.ActiveSessionID)
	sendContactList()
	return session
}
func sendMessageList(ID int64) {
	message := &[]byte{}
	log.Debugln("[axolotl] sendMessageList for conversation", ID)
	err, messageList := store.SessionsModel.GetMessageList(ID)
	if err != nil {
		log.Errorln("[axolotl] sendMessageList: ", err)
		return
	}
	messageList.Session.MarkRead()
	if push.Nh != nil {
		push.Nh.Clear(messageList.Session.Tel)
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
func sendMoreMessageList(id int64, lastId string) {
	message := &[]byte{}
	err, messageList := store.SessionsModel.GetMoreMessageList(id, lastId)
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
		s, err := store.SessionsModel.Get(activeChat)
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
	}
	*message, err = json.Marshal(configEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendConfig", err)
		return
	}
	broadcast <- *message
}
