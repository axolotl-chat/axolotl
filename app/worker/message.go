package worker

import (
	"github.com/nanu-c/textsecure-qml/app/sender"
	"github.com/nanu-c/textsecure-qml/app/store"
)

func (Api *TextsecureAPI) SendMessage(to, message string) error {
	return sender.SendMessageHelper(to, message, "")
}
func (Api *TextsecureAPI) DeleteMessage(msg *store.Message, tel string) {
	store.DeleteMessage(msg.ID)
	s := store.SessionsModel.Get(tel)
	for i, m := range s.Messages {
		if m.ID == msg.ID {
			s.Messages = append(s.Messages[:i], s.Messages[i+1:]...)
			s.Len--
			//qml.Changed(s, &s.Len)
			return
		}
	}
}
