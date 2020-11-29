package webserver

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/helpers"
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
				RemoveClientFromList(client)
			}
		}
	}
}
func sendRegistrationStatus() {
	log.Debugln("[axolotl-ws] getRegistrationStatus")
	if requestPassword {
		sendRequest("getEncryptionPw")
	} else if registered {
		sendRequest("registrationDone")
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
		fmt.Println(err)
		return
	}
	broadcast <- *message
}
func sendCurrentChat(s *store.Session) {
	var (
		err error
		gr  *textsecure.Group
		c   *textsecure.Contact
	)
	if s.IsGroup {
		gr, err = textsecure.GetGroupById(s.Tel)
	} else {
		c = store.GetContactForTel(s.Tel)
	}
	currentChatEnvelope := &CurrentChatEnvelope{
		OpenChat: &OpenChat{
			CurrentChat: s,
			Contact:     c,
			Group:       gr,
		},
	}
	message := &[]byte{}
	*message, err = json.Marshal(currentChatEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	broadcast <- *message
}
func updateCurrentChat(s *store.Session) {
	var (
		err error
		gr  *textsecure.Group
		c   *textsecure.Contact
	)
	if s.IsGroup {
		gr, err = textsecure.GetGroupById(s.Tel)
	} else {
		c = store.GetContactForTel(s.Tel)
	}
	updateCurrentChatEnvelope := &UpdateCurrentChatEnvelope{
		UpdateCurrentChat: &UpdateCurrentChat{
			CurrentChat: s,
			Contact:     c,
			Group:       gr,
		},
	}
	message := &[]byte{}
	*message, err = json.Marshal(updateCurrentChatEnvelope)
	if err != nil {
		fmt.Println(err)
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
	RemoveClientFromList(client)

}
func sendContactList() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
	var err error
	contactListEnvelope := &ContactListEnvelope{
		ContactList: store.ContactsModel.Contacts,
	}
	message := &[]byte{}
	*message, err = json.Marshal(contactListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	broadcast <- *message
}
func sendDeviceList() {
	var err error
	devices, err := textsecure.LinkedDevices()
	deviceListEnvelope := &DeviceListEnvelope{
		DeviceList: devices,
	}
	message := &[]byte{}
	*message, err = json.Marshal(deviceListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	broadcast <- *message
}
func createChat(tel string) *store.Session {
	return store.SessionsModel.Get(tel)
}
func createGroup(newGroupData CreateGroupMessage) *store.Session {
	group, err := textsecure.NewGroup(newGroupData.Name, newGroupData.Members)
	if err != nil {
		ShowError(err.Error())
		return nil
	}
	members := strings.Join(newGroupData.Members, ",")
	if !strings.Contains(members, config.Config.Tel) {
		// log.Debugln(members, config.Config.Tel)
		members = members + "," + config.Config.Tel
	}
	store.Groups[group.Hexid] = &store.GroupRecord{
		GroupID: group.Hexid,
		Name:    newGroupData.Name,
		Members: members,
	}
	store.SaveGroup(store.Groups[group.Hexid])
	session := store.SessionsModel.Get(group.Hexid)
	msg := session.Add(store.GroupUpdateMsg(append(newGroupData.Members, config.Config.Tel), newGroupData.Name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
	msg.Flags = helpers.MsgFlagGroupNew
	//qml.Changed(msg, &msg.Flags)
	store.SaveMessage(msg)

	return session
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
	session := store.SessionsModel.Get(group.Hexid)
	msg := session.Add(store.GroupUpdateMsg(updateGroupData.Members, updateGroupData.Name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
	msg.Flags = helpers.MsgFlagGroupNew
	//qml.Changed(msg, &msg.Flags)
	store.SaveMessage(msg)

	return session
}
func sendMessageList(id string) {
	message := &[]byte{}

	err, messageList := store.SessionsModel.GetMessageList(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	messageList.Session.MarkRead()
	chatListEnvelope := &MessageListEnvelope{
		MessageList: messageList,
	}
	*message, err = json.Marshal(chatListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	broadcast <- *message
}
func sendMoreMessageList(id string, lastId string) {
	message := &[]byte{}
	err, messageList := store.SessionsModel.GetMoreMessageList(id, lastId)
	if err != nil {
		fmt.Println(err)
		return
	}
	moreMessageListEnvelope := &MoreMessageListEnvelope{
		MoreMessageList: messageList,
	}
	*message, err = json.Marshal(moreMessageListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	// mu.Lock()
	// defer mu.Unlock()
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
		fmt.Println(err)
		return
	}
	broadcast <- *message
}

var test = false

func UpdateChatList() {

	if activeChat == "" && registered {
		sendChatList()
	}
}
func UpdateContactList() {

	if activeChat == "" && registered {
		sendContactList()
	}
}
func UpdateActiveChat() {
	if activeChat != "" {
		// log.Debugln("[axolotl] update activ	e chat")
		s := store.SessionsModel.Get(activeChat)
		updateCurrentChat(s)
	}
}

type SendGui struct {
	Gui string
}

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

type SendDarkmode struct {
	DarkMode bool
}

func SetUiDarkMode() {
	log.Debugln("[axolotl] send darkmode to client", settings.SettingsModel.DarkMode)

	var err error
	mode := settings.SettingsModel.DarkMode

	request := &SendDarkmode{
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
		fmt.Println(err)
		return
	}
	broadcast <- *message
}
