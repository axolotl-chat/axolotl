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
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
)

func (w *WsApp) websocketSender() {
	for {
		message := <-w.Broadcast
		for client := range w.Clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Errorln("[axolotl-ws] send message", err)
				w.removeClientFromList(client)
			}
		}
	}
}
func (w *WsApp) sendRegistrationStatus() {
	log.Debugln("[axolotl-ws] getRegistrationStatus")
	if requestUsername {
		w.sendRequest("getUsername")
	} else if registered {
		w.sendRequest("registrationDone")
	} else if requestPassword {
		w.sendRequest("getEncryptionPw")
	} else if requestSmsVerificationCode {
		w.sendRequest("getVerificationCode")
	} else {
		w.sendRequest("getPhoneNumber")
	}
}
func (w *WsApp) sendChatList() {
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
	w.Broadcast <- *message
}
func (w *WsApp) sendCurrentChat(s *store.Session) {
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
			Contact:     &w.Profile,
			Group:       group,
		},
	}
	message := &[]byte{}
	*message, err = json.Marshal(currentChatEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendCurrentChat: sendRegistrationStatus", err)
		return
	}
	w.Broadcast <- *message
}

func (w *WsApp) refreshContacts(path string) {
	var err error
	contact.GetAddressBookContactsFromContentHubWithFile(path)
	err = store.RefreshContacts()
	if err != nil {
		w.ShowError(err.Error())
	}
	go w.sendContactList()
}
func (w *WsApp) recoverFromWsPanic(client *websocket.Conn) {
	client.Close()
	w.removeClientFromList(client)

}
func (w *WsApp) sendContactList() {
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
	w.Broadcast <- *message
}
func (w *WsApp) sendDeviceList() {
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
	w.Broadcast <- *message
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
func (w *WsApp) createGroup(newGroupData CreateGroupMessage) (*store.Session, error) {
	group, err := textsecure.NewGroup(newGroupData.Name, newGroupData.Members)
	if err != nil {
		w.ShowError(err.Error())
		return nil, err
	}
	members := strings.Join(newGroupData.Members, ",")
	if !strings.Contains(members, w.App.Config.TextsecureConfig.Tel) {
		members = members + "," + w.App.Config.TextsecureConfig.Tel
	}
	store.Groups[group.Hexid] = &store.GroupRecord{
		GroupID: group.Hexid,
		Name:    newGroupData.Name,
		Members: members,
	}
	_, err = store.SaveGroup(store.Groups[group.Hexid])
	if err != nil {
		w.ShowError(err.Error())
		return nil, err
	}
	session := store.SessionsModel.CreateSessionForGroup(group)
	msg := session.Add(store.GroupUpdateMsg(append(newGroupData.Members, w.App.Config.TextsecureConfig.Tel), newGroupData.Name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
	// store.UpdateSession(session) // TODO: WIP 831
	// store.Sessions.MoveToTop(session.ID)
	msg.Flags = helpers.MsgFlagGroupNew
	store.SaveMessage(msg)
	return session, nil
}
func (w *WsApp) updateGroup(updateGroupData UpdateGroupMessage) *store.Session {
	group, err := textsecure.UpdateGroup(updateGroupData.ID, updateGroupData.Name, updateGroupData.Members)
	if err != nil {
		w.ShowError(err.Error())
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
		w.ShowError(err.Error())
		return nil
	}
	msg := session.Add(store.GroupUpdateMsg(updateGroupData.Members, updateGroupData.Name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
	// store.UpdateSession(session) // TODO: WIP 831
	// store.Sessions.MoveToTop(session.ID)
	msg.Flags = helpers.MsgFlagGroupUpdate
	store.SaveMessage(msg)

	return session

}
func (w *WsApp) joinGroup(joinGroupmessage JoinGroupMessage) *store.Session {
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
	members := ""
	// members cannot be read if the group is not yet joined
	if group.JoinStatus == store.GroupJoinStatusJoined {
		log.Infoln("[axolotl] joining group was successful", group.Hexid)
		for _, member := range group.DecryptedGroup.Members {
			members = members + string(member.Uuid) + ","
		}
	}

	storeGroup := &store.GroupRecord{
		GroupID:    group.Hexid,
		Name:       group.DecryptedGroup.Title,
		Members:    members,
		JoinStatus: group.JoinStatus,
	}
	store.Groups[group.Hexid] = storeGroup
	store.SaveGroup(storeGroup)
	session, _ := store.SessionsModel.GetByUUID(group.Hexid)
	session.Name = group.DecryptedGroup.Title
	session.GroupJoinStatus = group.JoinStatus
	// Add a join message to the session when the group is joined
	if group.JoinStatus == store.GroupJoinStatusJoined {
		msg := session.Add("You accepted the invitation to the group.", "", []store.Attachment{}, "", true, store.ActiveSessionID)
		// store.UpdateSession(session) // TODO: WIP 831
		// store.Sessions.MoveToTop(session.ID)
		msg.Flags = helpers.MsgFlagGroupJoined
		store.SaveMessage(msg)
		w.MessageHandler(msg)
	}
	store.UpdateSession(session)
	w.requestEnterChat(store.ActiveSessionID)
	w.sendContactList()
	return session
}
func (w *WsApp) sendMessageList(ID int64) {
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
	w.Broadcast <- *message
}
func (w *WsApp) sendMoreMessageList(lastId string) {
	message := &[]byte{}
	err, messageList := store.SessionsModel.GetMoreMessageList(w.ActiveChat, lastId)
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
	w.Broadcast <- *message
}
func (w *WsApp) sendIdentityInfo(fingerprintNumbers []string, fingerprintQRCode []byte) {
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
	w.Broadcast <- *message
}

// UpdateChatList updates the chatlist if not entered a chat + registered
func (w *WsApp) UpdateChatList() {

	if w.ActiveChat == -1 && registered {
		w.sendChatList()
	}
}

// UpdateContactList updates the contactlist if not entered a chat + registered
func (w *WsApp) UpdateContactList() {

	if w.ActiveChat == -1 && registered {
		w.sendContactList()
	}
}

// SendGui sends to axolotl the gui to trigger ubuntu touch specific parts
type SendGui struct {
	Gui string
}

// SetGui sets the gui
func (w *WsApp) SetGui() {
	var err error
	gui := w.App.Config.Gui
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
	w.Broadcast <- *message
	// RemoveClientFromList(client)
}

type sendDarkmode struct {
	DarkMode bool
}

func (w *WsApp) SetUiDarkMode() {
	log.Debugln("[axolotl] send darkmode to client", w.App.Settings.DarkMode)

	var err error
	mode := w.App.Settings.DarkMode

	request := &sendDarkmode{
		DarkMode: mode,
	}
	message := &[]byte{}
	*message, err = json.Marshal(request)
	if err != nil {
		log.Errorln("[axolotl] set darkmode", err)
		return
	}
	w.Broadcast <- *message
}
func (w *WsApp) sendConfig() {
	var err error

	message := &[]byte{}
	configEnvelope := &ConfigEnvelope{
		Type:             "config",
		RegisteredNumber: w.App.Config.TextsecureConfig.Tel,
		Version:          config.AppVersion,
		Gui:              w.App.Config.Gui,
		LogLevel:         w.App.Config.TextsecureConfig.LogLevel,
	}
	*message, err = json.Marshal(configEnvelope)
	if err != nil {
		log.Errorln("[axolotl] sendConfig", err)
		return
	}
	w.Broadcast <- *message
}

func (w *WsApp) importVcf() {
	contact.GetAddressBookContactsFromContentHubWithFile("import.vcf")
	err := store.RefreshContacts()
	if err != nil {
		w.ShowError(err.Error())
	}
	w.sendContactList()
}
