package webserver

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/nanu-c/textsecure-qml/app/store"
	log "github.com/sirupsen/logrus"
)

func attachmentsHandler(w http.ResponseWriter, r *http.Request) {
	Filename := r.URL.Query().Get("file")
	if Filename == "" {
		//Get not set, send a 400 bad request
		http.Error(w, "Get 'file' not specified in url.", 400)
		return
	}
	log.Debugln("[axolotl] open file: " + Filename)

	//Check if file exists and open
	Openfile, err := os.Open(Filename)
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
	log.Debugln("[axolotl] open avatar file: " + Filename)
	if Filename == "" {

		//Get not set, send a 400 bad request
		http.Error(w, "Get 'file' not specified in url.", 400)
		return
	}
	//handle group abvatars
	if len(Filename) > 30 {
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
	} else {
		log.Debugln("[axolotl] non group")
		Filename = "+" + Filename[1:]
		avatar := store.GetContactForTel(Filename)
		if avatar == nil {
			http.Error(w, "File not found.", 404)
			return
		}
		if len(avatar.Avatar) == 0 {
			//File not found, send 404
			http.Error(w, "File not found.", 404)
			return
		}
		log.Debugln("avatar len >0")
		FileContentType := http.DetectContentType(avatar.Avatar)
		log.Debugln(FileContentType)
		w.Header().Set("Content-Disposition", "attachment; filename="+Filename+".png")
		w.Header().Set("Content-Type", FileContentType)
		w.Write(avatar.Avatar)

	}
	//
	// //Check if file exists and open
	// Openfile, err := os.Open(Filename)
	// defer Openfile.Close() //Close after function return
	// if err != nil {
	// 	//File not found, send 404
	// 	http.Error(w, "File not found.", 404)
	// 	return
	// }
	// //File is found, create and send the correct headers
	//
	// //Get the Content-Type of the file
	// //Create a buffer to store the header of the file in
	// FileHeader := make([]byte, 512)
	// //Copy the headers into the FileHeader buffer
	// Openfile.Read(FileHeader)
	// //Get content type of file
	// FileContentType := http.DetectContentType(FileHeader)
	//
	// //Get the file size
	// FileStat, _ := Openfile.Stat()                     //Get info from file
	// FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string
	//
	// //Send the headers
	// w.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	// w.Header().Set("Content-Type", FileContentType)
	// w.Header().Set("Content-Length", FileSize)
	//
	// //Send the file
	// //We read 512 bytes from the file already, so we reset the offset back to 0
	// Openfile.Seek(0, 0)
	// io.Copy(w, Openfile) //'Copy' the file to the client
	// return
}
func attachmentServer() {
}
