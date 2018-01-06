package worker

import (
	"strings"

	"github.com/nanu-c/textsecure-qml/app/lang"
	"github.com/nanu-c/textsecure-qml/app/store"
	"github.com/nanu-c/textsecure-qml/app/ui"
)

func (Api *TextsecureAPI) EndSession(tel string) error {
	session := store.SessionsModel.Get(tel)
	m := session.Add(lang.SessionReset, "", "", "", true, Api.ActiveSessionID)
	m.Flags = msgFlagResetSession
	store.SaveMessage(m)
	go SendMessage(session, m)
	return nil
}

// MarkSessionsRead marks one or all sessions as read
func (Api *TextsecureAPI) MarkSessionsRead(tel string) {
	if tel != "" {
		s := store.SessionsModel.Get(tel)
		s.MarkRead()
		return
	}
	for _, s := range store.SessionsModel.Sess {
		s.MarkRead()
	}
}

func (Api *TextsecureAPI) DeleteSession(tel string) {
	err := store.DeleteSession(tel)
	if err != nil {
		ui.ShowError(err)
	}
}
func (Api *TextsecureAPI) FilterSessions(sub string) {
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

	ui.Engine.Context().SetVar("sessionsModel", sm)
}
