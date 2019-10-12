package webserver

import (
	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/store"
)

type MessageListEnvelope struct {
	MessageList *store.MessageList
}
type MoreMessageListEnvelope struct {
	MoreMessageList *store.MessageList
}
type ChatListEnvelope struct {
	ChatList []*store.Session
}
type ContactListEnvelope struct {
	ContactList []textsecure.Contact
}
type DeviceListEnvelope struct {
	DeviceList []textsecure.DeviceInfo
}
type CurrentChatEnvelope struct {
	CurrentChat *store.Session
}

type Message struct {
	Type string                 `json:"request"`
	Data map[string]interface{} `json:"-"` // Rest of the fields should go here.
}
type GetMessageListMessage struct {
	Type string `json:"request"`
	ID   string `json:"id"`
}
type GetMoreMessages struct {
	Type   string `json:"request"`
	LastID string `json:"lastId"`
}
type SendMessageMessage struct {
	Type    string `json:"request"`
	To      string `json:"to"`
	Message string `json:"message"`
}
type RequestCodeMessage struct {
	Type string `json:"request"`
	Tel  string `json:"tel"`
}
type SendPasswordMessage struct {
	Type string `json:"request"`
	Pw   string `json:"pw"`
}
type SetPasswordMessage struct {
	Type      string `json:"request"`
	Pw        string `json:"pw"`
	CurrentPw string `json:"CurrentPw"`
}
type SendCodeMessage struct {
	Type string `json:"request"`
	Code string `json:"code"`
}
type AddDeviceMessage struct {
	Type string `json:"request"`
	Url  string `json:"url"`
}
type DelDeviceMessage struct {
	Type string `json:"request"`
	Id   int    `json:"id"`
}
type AddContactMessage struct {
	Type  string `json:"request"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}
type EditContactMessage struct {
	Type  string `json:"request"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	ID    int    `json:"id"`
}
type DelContactMessage struct {
	Type string `json:"request"`
	ID   int    `json:"id"`
}
type RefreshContactsMessage struct {
	Type string `json:"request"`
	Url  string `json:"url"`
}
type UploadVcf struct {
	Type string `json:"request"`
	Vcf  string `json:"vcf"`
}
type DelChatMessage struct {
	Type string `json:"request"`
	ID   string `json:"id"`
}
type CreateChatMessage struct {
	Type string `json:"request"`
	Tel  string `json:"tel"`
}
type CreateGroupMessage struct {
	Type    string   `json:"request"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}
type SendAttachmentMessage struct {
	Type    string `json:"request"`
	AType   string `json:"type"`
	Path    string `json:"path"`
	To      string `json:"to"`
	Message string `json:"message"`
}
type ToggleNotificationsMessage struct {
	Type string `json:"request"`
	Chat string `json:"chat"`
}
