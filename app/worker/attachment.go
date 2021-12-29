package worker

import (
	"errors"
	"os"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/contact"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/ui"
	log "github.com/sirupsen/logrus"
)

// SendContactAttachment extracts the phone number from a contact and sends it as number
func (Api *TextsecureAPI) SendContactAttachment(to int64, message string, file string) error {
	phone, err := contact.PhoneFromVCardFile(file)
	if err != nil {
		log.Errorln("[axolotl] SendContactAttachment: ", err)
		return err
	}
	return Api.SendMessage(to, phone)
}

func (Api *TextsecureAPI) SendAttachmentToApi(to int64, message string, file string) error {
	// log.Printf("SendAttachmentApi")
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if fi.Size() > config.MaxAttachmentSize {
		ui.ShowError(errors.New("Attachment too large, not sending"))
		return nil
	}

	go sender.SendMessageHelper(to, message, file, nil, false)
	return nil
}
