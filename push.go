package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
)

// run as the application push helper
func pushHelperProcess() {
	in, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	err = pushHelperProcessMessage(in, out)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

type pushMessage struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

type appMessageCard struct {
	Summary   string   `json:"summary"`
	Body      string   `json:"body"`
	Actions   []string `json:"actions"`
	Popup     bool     `json:"popup"`
	Persist   bool     `json:"persist"`
	Timestamp int64    `json:"timestamp"`
}

type appMessageEmblemCounter struct {
	Count   int  `json:"count"`
	Visible bool `json:"visible"`
}

type appMessageNotification struct {
	Tag           string                  `json:"tag"`
	Card          appMessageCard          `json:"card"`
	Sound         bool                    `json:"sound"`
	Vibrate       bool                    `json:"vibrate"`
	EmblemCounter appMessageEmblemCounter `json:"emblem-counter"`
}

type appMessage struct {
	Message      string                 `json:"message"`
	Notification appMessageNotification `json:"notification"`
}

func pushHelperProcessMessage(in io.Reader, out io.Writer) error {
	pushMsg := &pushMessage{}
	dec := json.NewDecoder(in)
	err := dec.Decode(pushMsg)
	if err != nil {
		return err
	}

	appMsg := &appMessage{
		Message: pushMsg.Message,
		Notification: appMessageNotification{
			Tag: pushMsg.Sender,
			Card: appMessageCard{
				Summary:   pushMsg.Message,
				Body:      pushMsg.Message + " body",
				Actions:   []string{"appid://textsecure.jani/textsecure/current-user-version"},
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
