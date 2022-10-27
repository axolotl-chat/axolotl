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

// MarkSessionRead marks the session with this id as read
func (Api *TextsecureAPI) MarkSessionRead(ID int64) error {
	if ID != -1 {
		s, err := store.SessionsV2Model.GetSessionByID(ID)
		if err != nil {
			return err
		}
		s.MarkRead()
		return nil
	}
	return fmt.Errorf("session not found %d", ID)
}

func (Api *TextsecureAPI) DeleteSession(ID int64) error {
	if ID != -1 {
		session, err := store.SessionsV2Model.GetSessionByID(ID)
		if err != nil {
			return err
		}
		err = store.SessionsV2Model.DeleteSession(session)
		if err != nil {
			ui.ShowError(err)
		}
		return nil
	}
	return fmt.Errorf("session not found %d", ID)
}
