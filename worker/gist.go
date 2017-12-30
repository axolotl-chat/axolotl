package worker

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/nanu-c/textsecure-qml/store"
)

type entry struct {
	Content string `json:"content"`
}

type gist struct {
	Description string           `json:"description"`
	Public      bool             `json:"public"`
	Files       map[string]entry `json:"files"`
}

type gistResponse struct {
	HTML_URL string `json:"html_url"`
}

var (
	apiURL = "https://api.github.com/gists"
)

var re = regexp.MustCompile("/\\+[0-9]+")

// filterLogs removes potentially sensitive information from the logfile about to be submitted as part of a bug report.
func filterLogs(logs string) string {
	return re.ReplaceAllString(logs, "/+XXXXXXXXX")
}

func (api *TextsecureAPI) SubmitDebugLog() (string, error) {
	b, err := ioutil.ReadFile(store.LogFile)
	if err != nil {
		return "", err
	}

	content := filterLogs(string(b))
	f := make(map[string]entry)

	f["fname"] = entry{content}

	g := gist{
		Description: "Debug log",
		Public:      true,
		Files:       f,
	}
	b, err = json.MarshalIndent(&g, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	r := gistResponse{}
	err = dec.Decode(&r)
	if err != nil {
		return "", err
	}

	return r.HTML_URL, nil
}
