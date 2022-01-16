package store

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/signal-golang/textsecure"
	log "github.com/sirupsen/logrus"
)

type Attachment struct {
	File     string
	FileName string
	CType    int
}

func SaveAttachment(a *textsecure.Attachment) (Attachment, error) {

	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		return Attachment{}, err
	}
	dt := time.Now()

	// ext := ""
	// if strings.HasPrefix(a.MimeType, "video/") {
	// 	ext = strings.Replace(a.MimeType, "video/", ".", 1)
	// }
	fileName := a.FileName
	fileName = strings.Replace(fileName, "/", "-", -1)
	if fileName == "" {
		fileName = helpers.RandomString(10)
		extension, err := mime.ExtensionsByType(a.MimeType)
		if err != nil {
			log.Debugln("[axolotl] could not detect file extension", a.MimeType)
		} else if len(extension) > 0 {
			fileName += extension[0]
		}
	}
	log.Debugln("[axolotl] save attachment to",
		dt.Format("01-02-2006-15-04-05")+fileName)
	fn := filepath.Join(config.AttachDir, dt.Format("01-02-2006-15-04-05")+fileName)
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return Attachment{}, err
	}
	defer f.Close()

	_, err = io.Copy(f, a.R)
	if err != nil {
		return Attachment{}, err

	}

	return Attachment{File: fn, FileName: a.FileName}, nil
}

// CopyAttachment makes a copy of a file that is in the volatile content hub cache
func CopyAttachment(src string) (string, error) {
	_, b := filepath.Split(src)
	dest := filepath.Join(config.AttachDir, b)
	input, err := ioutil.ReadFile(src)
	if err != nil {
		log.Errorln("[axolotl] CopyAttachment ", err)
		return "", err
	}

	err = ioutil.WriteFile(dest, input, 0644)
	if err != nil {
		log.Errorln("[axolotl] CopyAttachment Error creating ", dest, err)
		return "", err
	}
	return dest, nil
}
func deleteAttachmentForMessage(id int64) error {
	m, err := GetMessageById(id)
	if err != nil {
		return err
	}
	if m.Attachment != "null" {
		attachments := []Attachment{}
		json.Unmarshal([]byte(m.Attachment), &attachments)
		log.Debugln("[axolotl] Delete attachment for message", id)
		for _, attachment := range attachments {
			e := os.Remove(attachment.File)
			if e != nil {
				return e
			}
		}

	}
	return nil
}
