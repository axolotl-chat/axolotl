package store

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// This is needed for version .26

func (ds *DataStore) UpdateSessionTable() error {
	statement, err := ds.Dbx.Prepare("SELECT * FROM sessions limit 1")
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
		_, err := ds.Dbx.Exec("ALTER TABLE sessions ADD COLUMN notification bool NOT NULL DEFAULT 1")
		if err != nil {
			return err
		}
	}

	return err
}

// fix v.0.7.8 add SendingError + expireTimer
func (ds *DataStore) UpdateMessagesTable_v_0_7_8() error {
	statement, err := ds.Dbx.Prepare("SELECT * FROM messages limit 1")
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
		_, err := ds.Dbx.Exec("ALTER TABLE messages ADD COLUMN sendingError bool DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = ds.Dbx.Exec("ALTER TABLE messages ADD COLUMN expireTimer integer DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = ds.Dbx.Exec("ALTER TABLE messages ADD COLUMN receipt bool DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = ds.Dbx.Exec("ALTER TABLE messages ADD COLUMN statusMessage bool DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return err
}

func (ds *DataStore) UpdateSessionTable_v_0_7_8() error {
	statement, err := ds.Dbx.Prepare("SELECT * FROM sessions limit 1")
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
		_, err := ds.Dbx.Exec("ALTER TABLE sessions ADD COLUMN expireTimer bool NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return err
}

// add support for quoted messages
func (ds *DataStore) UpdateSessionTable_v_0_9_0() error {
	statement, err := ds.Dbx.Prepare("SELECT * FROM messages limit 1")
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
		_, err := ds.Dbx.Exec("ALTER TABLE messages ADD COLUMN quoteId integer NOT NULL DEFAULT -1")
		if err != nil {
			return err
		}
	}

	return err
}

// add support uuids
func (ds *DataStore) UpdateSessionTable_v_0_9_5() error {
	// add uuid column to sessions table
	statement, err := ds.Dbx.Prepare("SELECT * FROM sessions limit 1")
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
		_, err := ds.Dbx.Exec("ALTER TABLE sessions ADD COLUMN 	type integer NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = ds.Dbx.Exec("ALTER TABLE sessions ADD COLUMN 	uuid string NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}
	statementM, err := ds.Dbx.Prepare("SELECT * FROM messages limit 1")
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
		_, err := ds.Dbx.Exec("ALTER TABLE messages ADD COLUMN srcUUID string NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return err
}

// MigrateMessagesFromSessionToAnotherSession copies the messages from the
// old session to the new session and then deletes the oldSession
func (ds *DataStore) MigrateMessagesFromSessionToAnotherSession(oldSession int64, newSession int64) error {
	log.Infoln("[axolotl] migrate messages to ", newSession)

	query := fmt.Sprintf("UPDATE messages SET sid=%d WHERE sid = %d;", newSession, oldSession)

	_, err := ds.Dbx.Exec(query)
	if err != nil {
		return err
	}
	// DeleteSession(oldSession) // TODO: WIP 831
	if err != nil {
		return err
	}
	return nil
}
func (ds *DataStore) updateGroupTable_v_0_9_10() error {
	statement, err := ds.Dbx.Prepare("SELECT * FROM groups limit 1")
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
		_, err := ds.Dbx.Exec("ALTER TABLE groups ADD COLUMN 	type integer NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}
	return nil
}

func (ds *DataStore) updateSessionTable_joinStatus_v_0_9_10() error {
	statement, err := ds.Dbx.Prepare("SELECT * FROM sessions limit 1")
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
		_, err := ds.Dbx.Exec("ALTER TABLE sessions ADD COLUMN groupJoinStatus integer NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}
	return nil
}
