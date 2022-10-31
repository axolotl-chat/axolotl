package store

import (
	"fmt"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/signal-golang/textsecure"
	"github.com/signal-golang/textsecure/contacts"
	log "github.com/sirupsen/logrus"
)

// UpdateSessionTable_v_0_9_0 adds support for quoted messages
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

// UpdateSessionTable_v_0_9_5 adds support uuids
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

// updateGroupTable_v_0_9_10 adds support for group types
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

// updateSessionTable_joinStatus_v_0_9_10 adds support for groupJoinStatus
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

// update_v_1_6_0 introduces the new data structure with sessionsv2, groupsv2, recipients
func update_v_1_6_0() error {
	// check if table exists and only migrate if it does not
	_, err := DS.Dbx.Prepare("SELECT * FROM sessionsv2 limit 1")
	if err != nil {
		log.Infoln("[axolotl] update schema v_1_6_0")
		log.Infoln("[axolotl][update v_1_6_0] create groupsv2 table")
		_, err = DS.Dbx.Exec(groupsV2Schema)
		if err != nil {
			return err
		}
		log.Infoln("[axolotl][update v_1_6_0] create groupv2 member table")
		_, err = DS.Dbx.Exec(groupV2MembersSchema)
		if err != nil {
			return err
		}
		log.Infoln("[axolotl][update v_1_6_0] create recipients table")
		_, err = DS.Dbx.Exec(recipientsSchema)
		if err != nil {
			return err
		}
		log.Infoln("[axolotl][update v_1_6_0] create sessionsv2 table")

		_, err = DS.Dbx.Exec(sessionsV2Schema)
		if err != nil {
			return err
		}
		var sessions []*Session
		err = DS.Dbx.Select(&sessions, sessionsSelect)
		if err != nil {
			return fmt.Errorf("error loading sessions: %s", err)
		}
		// copy contacts to recipients
		log.Infoln("[axolotl][update v_1_6_0] create recipients for contacts")
		registeredContacts, _ := readRegisteredContacts(config.RegisteredContactsFile)
		for i := range registeredContacts {
			contact := registeredContacts[i]
			if contact.UUID != "" {
				c := &contacts.Contact{
					UUID: contact.UUID,
					Name: contact.Name,
					Tel:  contact.Tel,
				}
				RecipientsModel.GetOrCreateRecipientForContact(c)
			}
		}
		log.Infoln("[axolotl][update v_1_6_0] migrate sessionsv1 to sessionsv2")
		for _, session := range sessions {
			if session.IsGroup && session.Type == SessionTypeGroupV2 {
				log.Infoln("[axolotl][update v_1_6_0] migrate groupv2 session")

				group, err := GroupV2sModel.Create(&GroupV2{
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
				log.Infoln("[axolotl][update v_1_6_0] migrate groupv2 session: members")
				groupMembers, err := textsecure.GetGroupV2MembersForGroup(session.UUID)
				if err != nil {
					log.Errorf("[axolotl][update v_1_6_0] error getting group members: %s", err)
				} else {
					err = group.AddGroupMembers(groupMembers)
					if err != nil {
						log.Errorf("[axolotl][update v_1_6_0] error adding group members: %s", err)
					}
				}
			} else if session.IsGroup && session.Type == SessionTypeGroupV1 {
				log.Infoln("[axolotl][update v_1_6_0] migrate groupv1 session")

				_, err = SessionsV2Model.SaveSession(&SessionV2{
					ID:                       session.ID,
					GroupV1ID:                session.UUID,
					DirectMessageRecipientID: int64(GroupRecipientsID),
				})
				if err != nil {
					return err
				}
			} else if session.Type == SessionTypePrivateChat {
				log.Infoln("[axolotl][update v_1_6_0] migrate private chat session")

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

	}
	return nil
}

// update_v_1_6_1 fixes the message histroy by introducing the renaming column sid to sv1id an reintrodrucing column sid in messages
func update_v_1_6_1() error {
	err := sessionsV1toSessionsV2()
	if err != nil {
		return err
	}
	return migrateMessageIds()
}

func migrateMessageIds() error {
	// check if new column exists and only migrate if it does not
	_, err := DS.Dbx.Prepare("SELECT sv1id FROM messages limit 1")
	if err == nil {
		return nil
	}
	log.Infoln("[axolotl][update v_1_6_1] add column sv1id")
	_, err = DS.Dbx.Exec("ALTER TABLE messages ADD sv1id integer;")
	if err != nil {
		return err
	}
	log.Infoln("[axolotl][update v_1_6_1] copy sid into sv1id")
	_, err = DS.Dbx.Exec("UPDATE messages SET sv1id = sid;")
	if err != nil {
		return err
	}
	log.Infoln("[axolotl][update v_1_6_1] delete sid of all messages")
	_, err = DS.Dbx.Exec("UPDATE messages SET sid = null;")
	if err != nil {
		return err
	}
	log.Infoln("[axolotl][update v_1_6_1] set sid for group messages")
	_, err = DS.Dbx.Exec("UPDATE messages SET sid = (SELECT v2.ID from sessions v1 JOIN sessionsv2 v2 ON v1.uuid = v2.groupV2Id where v1.ID = messages.sv1id) WHERE sid IS null;")
	if err != nil {
		return err
	}
	log.Infoln("[axolotl][update v_1_6_1] set sid for direct messages")
	_, err = DS.Dbx.Exec("UPDATE messages SET sid = (SELECT ID from sessionsv2 WHERE directMessageRecipientId = (SELECT r.id from recipients r JOIN sessions v1 ON r.uuid = v1.uuid WHERE v1.id = messages.sv1id)) WHERE sid IS null;")
	if err != nil {
		return err
	}
	log.Infoln("[axolotl][update v_1_6_1] set sid for messages of new sessions")
	_, err = DS.Dbx.Exec("UPDATE messages SET sid = sv1id WHERE sid IS null;")
	return err
}

func sessionV1ToGroupV2(session *Session) error {
	group, err := GroupV2sModel.GetGroupById(session.UUID)
	if group != nil && err == nil {
		//allready migrated
		return nil
	}
	log.Infoln("[axolotl][update v_1_6_1] migrate groupv2 session")

	group, err = GroupV2sModel.Create(&GroupV2{
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
	log.Infoln("[axolotl][update v_1_6_1] migrate groupv2 session: members")
	groupMembers, err := textsecure.GetGroupV2MembersForGroup(session.UUID)
	if err != nil {
		return fmt.Errorf("[axolotl][update v_1_6_1] error getting group members: %s", err)
	}
	err = group.AddGroupMembers(groupMembers)
	if err != nil {
		return fmt.Errorf("[axolotl][update v_1_6_1] error adding group members: %s", err)
	}
	return nil
}

func sessionV1ToSessionV2(session *Session) error {
	if session.IsGroup && session.Type == SessionTypeGroupV2 {
		err := sessionV1ToGroupV2(session)
		if err != nil {
			return err
		}
	} else if session.IsGroup && session.Type == SessionTypeGroupV1 {
		sessionV2, err := SessionsV2Model.GetSessionByID(session.ID)
		if sessionV2 != nil && err == nil {
			//allready migrated
			return nil
		}
		log.Infoln("[axolotl][update v_1_6_1] migrate groupv1 session")
		_, err = SessionsV2Model.SaveSession(&SessionV2{
			ID:                       session.ID,
			GroupV1ID:                session.UUID,
			DirectMessageRecipientID: int64(GroupRecipientsID),
		})
		if err != nil {
			return err
		}
	} else if session.Type == SessionTypePrivateChat {
		sessionV2, err := SessionsV2Model.GetSessionByID(session.ID)
		if sessionV2 != nil && err == nil {
			//allready migrated
			return nil
		}
		recipient := RecipientsModel.GetRecipientByUUID(session.UUID)
		log.Infoln("[axolotl][update v_1_6_1] migrate private chat session")
		if recipient == nil {
			recipient, err = RecipientsModel.CreateRecipientWithoutProfileUpdate(&Recipient{
				UUID:             session.UUID,
				ProfileGivenName: session.Name,
				E164:             session.Tel,
			})
			if err != nil {
				return err
			}
		}
		_, err = SessionsV2Model.SaveSession(&SessionV2{
			ID:                       session.ID,
			DirectMessageRecipientID: recipient.Id,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func sessionsV1toSessionsV2() error {
	var sessions []*Session
	err := DS.Dbx.Select(&sessions, sessionsSelect)
	if err != nil {
		return fmt.Errorf("error loading sessions: %s", err)
	}
	log.Infoln("[axolotl][update v_1_6_1] migrate sessionsv1 to sessionsv2")
	for _, session := range sessions {
		err = sessionV1ToSessionV2(session)
		if err != nil {
			log.Errorf("failed to migrate session %s. Error: %s", session.UUID, err)
		}
	}
	return nil
}
