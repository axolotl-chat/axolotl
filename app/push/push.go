package push

import (
	"encoding/json"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// run as the application push helper
func PushHelperProcess() {
	in, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	err = PushHelperProcessMessage(in, out)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

type AppMessageCard struct {
	Summary   string   `json:"summary"`
	Body      string   `json:"body"`
	Actions   []string `json:"actions"`
	Popup     bool     `json:"popup"`
	Persist   bool     `json:"persist"`
	Timestamp int64    `json:"timestamp"`
}

type AppMessageEmblemCounter struct {
	Count   int  `json:"count"`
	Visible bool `json:"visible"`
}

type AppMessageNotification struct {
	Tag           string                  `json:"tag"`
	Card          AppMessageCard          `json:"card"`
	Sound         bool                    `json:"sound"`
	Vibrate       bool                    `json:"vibrate"`
	EmblemCounter AppMessageEmblemCounter `json:"emblem-counter"`
}

type AppMessage struct {
	Notification AppMessageNotification `json:"notification"`
}

func PushHelperProcessMessage(in io.Reader, out io.Writer) error {
	pushMsg := &PushMessage{}
	dec := json.NewDecoder(in)
	err := dec.Decode(pushMsg)
	if err != nil {
		return err
	}

	appMsg := &AppMessage{
		Notification: AppMessageNotification{
			Card: AppMessageCard{
				Summary:   "New message",
				Body:      "",
				Actions:   []string{"appid://textsecure.nanuc/textsecure/current-user-version"},
				Popup:     true,
				Persist:   true,
				Timestamp: time.Now().Unix(),
			},
			Sound:   true,
			Vibrate: true,
		},
	}

	enc := json.NewEncoder(out)
	err = enc.Encode(appMsg)
	return err
}
