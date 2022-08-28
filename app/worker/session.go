package worker

import (
	"fmt"

	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/nanu-c/axolotl/app/ui"
)

// EndSession resets the current session.
func (Api *TextsecureAPI) EndSession(ID int64) error {
	session, err := store.SessionsV2Model.GetSessionByID(ID)
	if err != nil {
		return err
	}
	m := &store.Message{
		Message: "Secure session reset",
		SID:     store.ActiveSessionID,
		Flags:   helpers.MsgFlagResetSession,
	}
	store.SaveMessage(m)
	go sender.SendMessage(session, m, false)
	return nil
}

// MarkSessionsRead marks one or all sessions as read
func (Api *TextsecureAPI) MarkSessionRead(ID int64) error {
	if ID != -1 {
		s, err := store.SessionsV2Model.GetSessionByID(ID)
		if err != nil {
			return err
		}
		s.MarkRead()
		return nil
	}
	return fmt.Errorf("Session not found %d", ID)
}

func (Api *TextsecureAPI) DeleteSession(ID int64) {
	err := store.DeleteSession(ID)
	if err != nil {
		ui.ShowError(err)
	}
}
