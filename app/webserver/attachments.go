package webserver

import (
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
		//Get not set, send a 400 bad request
		http.Error(w, "Get 'file' not specified in url.", 400)
		return
	}

	//Check if file exists and open
	filename := strings.Split(Filename, "/")
	path := config.AttachDir + "/" + filename[len(filename)-1]
	log.Debugln("[axolotl] open file: " + path)
	Openfile, err := os.Open(path)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(w, "File not found.", 404)
		return
	}
	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(w, Openfile) //'Copy' the file to the client
	return
}
func avatarsHandler(w http.ResponseWriter, r *http.Request) {
	Filename := r.URL.Query().Get("file")
	// log.Debugln("[axolotl] open avatar file: " + Filename)
	if Filename == "" {

		//Get not set, send a 400 bad request
		http.Error(w, "Get 'file' not specified in url.", 400)
		return
	}
	//handle group abvatars
	if len(Filename)  != 32 {
		Filename = strings.ReplaceAll(Filename, " ", "+")
		profile, err := textsecure.GetProfile(Filename)
		if err != nil || len(profile.Avatar) == 0 {
			http.Error(w, "File not found.", 404)
			return
		}
		avatar := []byte(profile.Avatar)
		FileContentType := http.DetectContentType(avatar)
		w.Header().Set("Content-Disposition", "attachment; filename="+Filename+".png")
		w.Header().Set("Content-Type", FileContentType)
		w.Write(avatar)
	} else {
		group := store.GetGroupById(Filename)
		if group == nil {
			//File not found, send 404
			http.Error(w, "File not found.", 404)
			return
		}
		FileContentType := http.DetectContentType(group.Avatar)
		w.Header().Set("Content-Disposition", "attachment; filename="+Filename+".png")
		w.Header().Set("Content-Type", FileContentType)
		w.Write(group.Avatar)
	}
}
