package webserver

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/signal-golang/textsecure"

	log "github.com/sirupsen/logrus"
)

func attachmentsHandler(w http.ResponseWriter, r *http.Request) {
	Filename := r.URL.Query().Get("file")
	if Filename == "" {
		// Get not set, send a 400 bad request
		http.Error(w, "Get 'file' not specified in url.", 400)
		return
	}

	// Check if file exists and open
	filename := strings.Split(Filename, "/")
	path := config.AttachDir + "/" + filename[len(filename)-1]
	log.Debugln("[axolotl] open file: " + path)
	Openfile, err := os.Open(path)
	defer Openfile.Close() // Close after function return
	if err != nil {
		// File not found, send 404
		http.Error(w, "File not found.", 404)
		return
	}
	// File is found, create and send the correct headers

	// Get the Content-Type of the file
	// Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	// Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	// Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	// Get the file size
	FileStat, _ := Openfile.Stat()                     // Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) // Get file size as a string

	// Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	// Send the file
	// We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(w, Openfile) // 'Copy' the file to the client
}
func avatarsHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := strconv.ParseInt(r.URL.Query().Get("session"), 10, 64)
	if err == nil {
		avatar, err := getAvatarForSession(sessionId)
		if err != nil {
			http.Error(w, "File not found.", 404)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprint(sessionId)+".png")

		w.Header().Set("Content-Length", strconv.Itoa(len(avatar)))
		w.Write(avatar)
		return

	}
	recipientId, err := strconv.ParseInt(r.URL.Query().Get("recipient"), 10, 64)
	if err == nil {
		avatar, err := getAvatarForRecipient(recipientId)
		if err != nil {
			http.Error(w, "File not found.", 404)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprint(recipientId)+".png")
		w.Header().Set("Content-Length", strconv.Itoa(len(avatar)))
		w.Write(avatar)
		return
	}
	recipientE164 := r.URL.Query().Get("e164")
	if recipientE164 != "" {
		e164 := "+" + strings.ReplaceAll(recipientE164, " ", "")
		avatar, err := getAvatarForE164(e164)
		if err != nil {
			http.Error(w, "File not found.", 404)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "attachment; filename="+recipientE164+".png")
		w.Header().Set("Content-Length", strconv.Itoa(len(avatar)))
		w.Write(avatar)
		return
	}

	Filename := r.URL.Query().Get("group")

	// log.Debugln("[axolotl] open avatar file: " + Filename)
	if Filename == "" {

		// Get not set, send a 400 bad request
		http.Error(w, "Parameter wrong not specified in url.", 400)
		return
	}
	// handle group abvatars
	if len(Filename) > 30 {
		group := store.GetGroupById(Filename)
		if group == nil {
			// File not found, send 404
			http.Error(w, "File not found.", 404)
			return
		}
		FileContentType := http.DetectContentType(group.Avatar)
		w.Header().Set("Content-Disposition", "attachment; filename="+Filename+".png")
		w.Header().Set("Content-Type", FileContentType)
		w.Write(group.Avatar)
		return
	}

	http.Error(w, "File not found.", 404)
}

func getAvatarForSession(sessionId int64) ([]byte, error) {
	session, err := store.SessionsV2Model.GetSessionByID(sessionId)
	if err != nil {
		return nil, err
	}
	return getAvatarForRecipient(session.DirectMessageRecipientID)
}

func getAvatarForRecipient(recipientId int64) ([]byte, error) {
	recipient := store.RecipientsModel.GetRecipientById(recipientId)
	if recipient == nil {
		return nil, nil
	}
	avatar, err := textsecure.GetAvatar(recipient.UUID)
	if err != nil {
		return nil, err
	}
	if avatar == nil {
		return nil, nil
	}
	avatarBytes, err := io.ReadAll(avatar)
	if err != nil {
		return nil, err
	}
	return avatarBytes, nil
}

func getAvatarForE164(e164 string) ([]byte, error) {
	recipient := store.RecipientsModel.GetRecipientByE164(e164)
	if recipient == nil {
		return nil, nil
	}
	avatar, err := textsecure.GetAvatar(recipient.UUID)
	if err != nil {
		return nil, err
	}
	if avatar == nil {
		return nil, nil
	}
	avatarBytes, err := io.ReadAll(avatar)
	if err != nil {
		return nil, err
	}
	return avatarBytes, nil
}
