package worker

import (
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/store"
)

func (Api *TextsecureAPI) SendMessage(to, message string) error {
	err, _ := sender.SendMessageHelper(to, message, "", nil)
	return err
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
