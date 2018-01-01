package attachment

import (
	"errors"
	"log"
	"os"

	"github.com/nanu-c/textsecure-qml/contact"
	"github.com/nanu-c/textsecure-qml/ui"
	"github.com/nanu-c/textsecure-qml/worker"
)

type Attachments struct {
	Len int
}

var AttachmentsModel *Attachments = &Attachments{}

func (a *Attachments) SendContactAttachment(to, message string, file string) error {
	phone, err := contact.PhoneFromVCardFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	return worker.Api.SendMessage(to, phone)
}

// Do not allow sending attachments larger than 100M for now
var maxAttachmentSize int64 = 100 * 1024 * 1024

func (a *Attachments) SendAttachment(to, message string, file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if fi.Size() > maxAttachmentSize {
		ui.ShowError(errors.New("Attachment too large, not sending"))
		return nil
	}

	go worker.SendMessageHelper(to, message, file)
	return nil
}
