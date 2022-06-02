package worker

import (
	"fmt"
	"strings"

	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/nanu-c/axolotl/app/ui"
)

// EndSession resets the current session.
func (a *TextsecureAPI) EndSession(ID int64) error {
	session, err := store.SessionsModel.Get(ID)
	if err != nil {
		return err
	}
	m := session.Add("Secure session reset.", "", []store.Attachment{}, "", true, store.ActiveSessionID)
	m.Flags = helpers.MsgFlagResetSession
	store.SaveMessage(m)
	go sender.SendMessage(session, m, false)
	return nil
}

// MarkSessionsRead marks one or all sessions as read
func (a *TextsecureAPI) MarkSessionRead(ID int64) error {
	if ID != -1 {
		s, err := store.SessionsModel.Get(ID)
		if err != nil {
			return err
		}
		s.MarkRead()
		return nil
	}
	return fmt.Errorf("Session not found %d", ID)
}

func (a *TextsecureAPI) DeleteSession(ID int64) {
	err := store.DeleteSession(ID)
	if err != nil {
		ui.ShowError(err, a.Websocket)
	}
}
func (a *TextsecureAPI) FilterSessions(sub string) {
	sub = strings.ToUpper(sub)

	sm := &store.Sessions{
		Sess: make([]*store.Session, 0),
	}

	for _, s := range store.SessionsModel.Sess {
		if strings.Contains(strings.ToUpper(store.TelToName(s.Tel)), sub) {
			sm.Sess = append(sm.Sess, s)
			sm.Len++
		}
	}

	// ui.Engine.Context().SetVar("sessionsModel", sm)
}
