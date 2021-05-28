package store

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// This is needed for version .26

func UpdateSessionTable() error {
	statement, err := DS.Dbx.Prepare("SELECT * FROM sessions limit 1")
	if err != nil {
		return err
	}
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if len(col) == 8 {
		log.Infof("[axolotl] Update session schema")
		_, err := DS.Dbx.Exec("ALTER TABLE sessions ADD COLUMN notification bool NOT NULL DEFAULT 1")
		if err != nil {
			return err
		}
	}

	return err
}

// fix v.0.7.8 add SendingError + expireTimer
func UpdateMessagesTable_v_0_7_8() error {
	statement, err := DS.Dbx.Prepare("SELECT * FROM messages limit 1")
	if err != nil {
		return err
	}
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if len(col) == 12 {
		log.Infoln("[axolotl] Update messages schema for v0.7.8")
		_, err := DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN sendingError bool DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN expireTimer integer DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN receipt bool DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN statusMessage bool DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return err
}

func UpdateSessionTable_v_0_7_8() error {
	statement, err := DS.Dbx.Prepare("SELECT * FROM sessions limit 1")
	if err != nil {
		return err
	}
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if len(col) == 9 {
		log.Infof("[axolotl] Update session schema v_0_7_8")
		_, err := DS.Dbx.Exec("ALTER TABLE sessions ADD COLUMN expireTimer bool NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return err
}

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
	if len(col) == 10 {
		log.Infof("[axolotl] Update sessions schema v_0_9_5")
		_, err := DS.Dbx.Exec("ALTER TABLE sessions ADD COLUMN 	type integer NOT NULL DEFAULT 0")
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

// MigrateMessagesFromSessionToAnotherSession copies the messages from the
// old session to the new session and then deletes the oldSession
func MigrateMessagesFromSessionToAnotherSession(oldSession int64, newSession int64) error {
	log.Infoln("[axolotl] migrate messages to ", newSession)

	query := fmt.Sprintf("UPDATE messages SET sid=%d WHERE sid = %d;", newSession, oldSession)

	_, err := DS.Dbx.Exec(query)
	if err != nil {
		return err
	}
	DeleteSession(oldSession)
	if err != nil {
		return err
	}
	return nil
}
func updateGroupTable_v_0_9_10() error {
	statement, err := DS.Dbx.Prepare("SELECT * FROM groups limit 1")
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
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
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if len(col) == 12 {
		log.Infof("[axolotl] Update sessions schema join status v_0_9_10")
		_, err := DS.Dbx.Exec("ALTER TABLE sessions ADD COLUMN groupJoinStatus integer NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}
	return nil
}
