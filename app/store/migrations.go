package store

import (
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
func UpdateMessagesTable() error {
	statement, err := DS.Dbx.Prepare("SELECT * FROM sessions limit 1")
	if err != nil {
		return err
	}
	res, err := statement.Query()
	if err != nil {
		return err
	}

	col, err := res.Columns()
	if len(col) == 12 {
		log.Infof("[axolotl] Update messages schema")
		_, err := DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN sendingError bool NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN expireTimer integer NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
		_, err = DS.Dbx.Exec("ALTER TABLE messages ADD COLUMN receipt bool NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return err
}
