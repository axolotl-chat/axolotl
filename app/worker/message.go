package worker

import (
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/store"
)

func (Api *TextsecureAPI) SendMessage(to int64, message string) error {
	_, err := sender.SendMessageHelper(to, message, "", nil, false)
	return err
}
func (Api *TextsecureAPI) DeleteMessage(msg *store.Message) {
	store.DeleteMessage(msg.ID)
}
