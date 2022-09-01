package webserver

import (
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"
	textsecureContacts "github.com/signal-golang/textsecure/contacts"
)

type MessageListEnvelope struct {
	MessageList []*store.Message
}
type MoreMessageListEnvelope struct {
	MoreMessageList *store.MessageList
}
type ChatListEnvelope struct {
	ChatList     []*store.SessionV2
	LastMessages []store.Message
	SessionNames []store.SessionV2Name
}
type ContactListEnvelope struct {
	ContactList []textsecureContacts.Contact
}
type DeviceListEnvelope struct {
	DeviceList []textsecure.DeviceInfo
}
type Group struct {
	HexId   string
	Name    string
	Members []*store.Recipient
}
type OpenChat struct {
	CurrentChat *store.SessionV2
	Contact     *textsecureContacts.Contact
	Group       *Group
}
type CurrentChatEnvelope struct {
	OpenChat *OpenChat
}
type UpdateCurrentChat struct {
	CurrentChat *store.SessionV2
	Contact     *store.Recipient
	Group       *Group
}
type UpdateCurrentChatEnvelope struct {
	UpdateCurrentChat *UpdateCurrentChat
}
type IdentityEnvelope struct {
	FingerprintNumbers []string
	FingerprintQRCode  []int
}
type ConfigEnvelope struct {
	Type             string
	Version          string
	RegisteredNumber string
	Name             string
	Notifications    bool
	Encryption       bool
	Gui              string
	LogLevel         string
}
type Message struct {
	Type string                 `json:"request"`
	Data map[string]interface{} `json:"-"` // Rest of the fields should go here.
}
type DelMessageMessage struct {
	Type string `json:"request"`
	ID   int64  `json:"id"`
}
type GetMessageListMessage struct {
	Type string `json:"request"`
	ID   int64  `json:"id"`
}
type GetMoreMessages struct {
	Type   string `json:"request"`
	LastID string `json:"lastId"`
}
type SendMessageMessage struct {
	Type    string `json:"request"`
	To      int64  `json:"to"`
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
type SendUsernameMessage struct {
	Type     string `json:"request"`
	Username string `json:"username"`
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
type SendPinMessage struct {
	Type string `json:"request"`
	Pin  string `json:"pin"`
}
type SendCaptchaTokenMessage struct {
	Type  string `json:"request"`
	Token string `json:"token"`
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
	UUID  string `json:"uuid"`
}
type EditContactMessage struct {
	Type  string `json:"request"`
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	ID    int    `json:"id"`
}
type DelContactMessage struct {
	Type string `json:"request"`
	ID   string `json:"id"`
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
	ID   int64  `json:"id"`
}
type CreateChatMessage struct {
	Type string `json:"request"`
	UUID string `json:"uuid"`
}
type OpenChatMessage struct {
	Type string `json:"request"`
	Id   int64  `json:"id"`
}
type CreateGroupMessage struct {
	Type    string   `json:"request"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}
type JoinGroupMessage struct {
	Type string `json:"request"`
	ID   string `json:"id"`
}
type SendAttachmentMessage struct {
	Type    string `json:"request"`
	AType   string `json:"type"`
	Path    string `json:"path"`
	To      int64  `json:"to"`
	Message string `json:"message"`
}
type SendVoiceNoteMessage struct {
	Type      string `json:"request"`
	To        int64  `json:"to"`
	VoiceNote string `json:"voiceNote"`
}
type UploadAttachmentMessage struct {
	Type       string `json:"request"`
	To         int64  `json:"to"`
	Attachment string `json:"attachment"`
	Message    string `json:"message"`
}
type toggleNotificationsMessage struct {
	Type string `json:"request"`
	Chat int64  `json:"chat"`
}
type ResetEncryptionMessage struct {
	Type string `json:"request"`
	Chat int64  `json:"chat"`
}
type verifyIdentityMessage struct {
	Type string `json:"request"`
	Chat int64  `json:"chat"`
}
type SetDarkMode struct {
	Type     string `json:"request"`
	DarkMode bool   `json:"darkMode"`
}
type SetLogLevelMessage struct {
	Type  string `json:"request"`
	Level string `json:"level"`
}
type GetProfileMessage struct {
	Type string `json:"request"`
	ID   int64  `json:"id"`
}
type UpdateProfileNameMessage struct {
	Type string `json:"request"`
	Name string `json:"name"`
	ID   int64  `json:"id"`
}
type CreateChatForRecipientMessage struct {
	Type string `json:"request"`
	ID   int64  `json:"id"`
}
type CreateRecipientMessage struct {
	Type      string `json:"request"`
	Recipient string `json:"recipient"`
}
type CreateRecipientAndAddToGroupMessage struct {
	Type      string `json:"request"`
	Recipient string `json:"recipient"`
	Group     string `json:"group"`
}
type UpdateMessage struct {
	UpdateMessage *store.Message
}

type ProfileMessage struct {
	Recipient *store.Recipient
	Contact   *textsecureContacts.Contact
}

type ProfileMessageEnvelope struct {
	ProfileMessage *ProfileMessage
}

type SendEnterChatRequest struct {
	Type string
	Chat int64
}

type SendRequest struct {
	Type string
}

type SendError struct {
	Type  string
	Error string
}

type MessageRecieved struct {
	MessageRecieved *store.Message
}
