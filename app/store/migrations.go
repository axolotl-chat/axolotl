package store

import (
	"fmt"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/signal-golang/textsecure/contacts"
	log "github.com/sirupsen/logrus"
)

// add support for quoted messages
func UpdateSessionTable_v_0_9_0() error {
	statement, err := DS.Dbx.Prepare("SELECT * FROM messages limit 1")
	if err != nil {
		return err
	}
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if len(col) == 16 {
		log.Infof("[axolotl] Update session schema v_0_9_0")
		_, err := DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN quoteId integer NOT NULL DEFAULT -1")
		if err != nil {
			return err
		}
	}

	return err
}

// add support uuids
func UpdateSessionTable_v_0_9_5() error {
	// add uuid column to sessions table
	statement, err := DS.Dbx.Prepare("SELECT * FROM sessions limit 1")
	if err != nil {
		return err
	}
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if err != nil {
		return err
	}
	if len(col) == 10 {
		log.Infof("[axolotl] Update sessions schema v_0_9_5")
		_, err := DS.Dbx.Exec("ALTER TABLE sessions ADD COLUMN 	type integer NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec("ALTER TABLE sessions ADD COLUMN 	uuid string NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}
	statementM, err := DS.Dbx.Prepare("SELECT * FROM messages limit 1")
	if err != nil {
		return err
	}
	resM, err := statementM.Query()
	if err != nil {
		return err
	}

	colM, err := resM.Columns()

	if len(colM) == 17 {
		log.Infof("[axolotl] Update messages schema v_0_9_5")
		_, err := DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN srcUUID string NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return err
}

func updateGroupTable_v_0_9_10() error {
	statement, err := DS.Dbx.Prepare("SELECT * FROM groups limit 1")
	if err != nil {
		return err
	}
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if err != nil {
		return err
	}
	if len(col) == 10 {
		log.Infof("[axolotl] Update groups schema v_0_9_10")
		_, err := DS.Dbx.Exec("ALTER TABLE groups ADD COLUMN 	type integer NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}
	return nil
}

func updateSessionTable_joinStatus_v_0_9_10() error {
	statement, err := DS.Dbx.Prepare("SELECT * FROM sessions limit 1")
	if err != nil {
		return err
	}
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if err != nil {
		return err
	}
	if len(col) == 12 {
		log.Infof("[axolotl] Update sessions schema join status v_0_9_10")
		_, err := DS.Dbx.Exec("ALTER TABLE sessions ADD COLUMN groupJoinStatus integer NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}
	return nil
}

func update_v_1_6_0() error {
	// check if table exists and only migrate if it does not
	_, err := DS.Dbx.Prepare("SELECT * FROM sessionsv2 limit 1")
	if err != nil {
		log.Infoln("[axolotl] update schema v_1_6_0")
		_, err = DS.Dbx.Exec(groupsV2Schema)
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec(groupV2MembersSchema)
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec(recipientsSchema)
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec(sessionsV2Schema)
		if err != nil {
			return err
		}
		var Groups []*GroupRecord
		err = DS.Dbx.Select(&Groups, groupsSelect)
		if err != nil {
			return fmt.Errorf("error loading groups: %s", err)
		}
		var sessions []*Session
		err = DS.Dbx.Select(&sessions, sessionsSelect)
		if err != nil {
			return fmt.Errorf("error loading sessions: %s", err)
		}
		for _, session := range sessions {
			if session.IsGroup && session.Type == SessionTypeGroupV2 {

				_, err = GroupV2sModel.Create(&GroupV2{
					Id:         session.UUID,
					Name:       session.Name,
					JoinStatus: session.GroupJoinStatus,
				})
				if err != nil {
					return fmt.Errorf("error creating group v2: %s", err)
				}
				_, err = SessionsV2Model.SaveSession(&SessionV2{
					ID:                       session.ID,
					DirectMessageRecipientID: int64(GroupRecipientsID),
					GroupV2ID:                session.UUID,
				})
				if err != nil {
					return fmt.Errorf("error creating session groupv2: %s", err)
				}
			} else if session.IsGroup && session.Type == SessionTypeGroupV1 {
				_, err = SessionsV2Model.SaveSession(&SessionV2{
					ID:                       session.ID,
					GroupV1ID:                session.UUID,
					DirectMessageRecipientID: int64(GroupRecipientsID),
				})
				if err != nil {
					return err
				}
			} else if session.Type == SessionTypePrivateChat {
				recipient, err := RecipientsModel.CreateRecipientWithoutProfileUpdate(&Recipient{
					UUID:             session.UUID,
					ProfileGivenName: session.Name,
					E164:             session.Tel,
				})
				if err != nil {
					return err
				}
				_, err = SessionsV2Model.SaveSession(&SessionV2{
					ID:                       session.ID,
					DirectMessageRecipientID: recipient.Id,
				})
				if err != nil {
					return err
				}
			}
		}
		// copy contacts to recipients
		registeredContacts, _ := readRegisteredContacts(config.RegisteredContactsFile)
		for _, contact := range registeredContacts {
			if contact.UUID != "" {
				c := &contacts.Contact{
					UUID: contact.UUID,
					Name: contact.Name,
					Tel:  contact.Tel,
				}
				RecipientsModel.GetOrCreateRecipientForContact(c)
			}
		}
	}
	return nil
}
