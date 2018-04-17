package store

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/mutecomm/go-sqlcipher"
	qml "github.com/nanu-c/qml-go"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/settings"
)

var DS *DataStore

type DataStore struct {
	Dbx *sqlx.DB
}

var (
	dbDir    string
	dbFile   string
	saltFile string

	sessionsSchema = "CREATE TABLE IF NOT EXISTS sessions (id INTEGER PRIMARY KEY, name text, tel text, isgroup boolean, last string, timestamp integer, ctype integer, unread integer default 0)"
	sessionsInsert = "INSERT OR REPLACE INTO sessions (name, tel, isgroup, last, ctype, timestamp) VALUES (:name, :tel, :isgroup, :last, :ctype, :timestamp)"
	sessionsSelect = "SELECT * FROM sessions ORDER BY timestamp DESC"

	messagesSchema      = "CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY, sid integer, source text, message text, outgoing boolean, sentat integer, receivedat integer, ctype integer, attachment string, issent boolean, isread boolean, flags integer default 0)"
	messagesInsert      = "INSERT INTO messages (sid, source, message, outgoing, sentat, receivedat, ctype, attachment, issent, isread, flags) VALUES (:sid, :source, :message, :outgoing, :sentat, :receivedat, :ctype, :attachment, :issent, :isread, :flags)"
	messagesSelectWhere = "SELECT * FROM messages WHERE sid = ?"

	groupsSchema = "CREATE TABLE IF NOT EXISTS groups (id INTEGER PRIMARY KEY, groupid TEXT, name TEXT, members TEXT, avatar BLOB, avatarid INTEGER, avatar_key BLOB, avatar_content_type TEXT, relay TEXT, active INTEGER DEFAULT 1)"
	groupsInsert = "INSERT OR REPLACE INTO groups (groupid, name, members, avatar) VALUES (:groupid, :name, :members, :avatar)"
	groupsUpdate = "UPDATE groups SET members = :members, name = :name, avatar = :avatar, active = :active WHERE groupid = :groupid"
	groupsSelect = "SELECT groupid, name, members, avatar, active FROM groups"
	groupsDelete = "DELETE FROM groups WHERE groupid = ?"
)

// func SetupDB(password string) error {
// 	params := "_busy_timeout=5000&cache=shared"
// 	var err error
//
// 	dbDir = filepath.Join(config.DataDir, "db")
// 	dbFile = filepath.Join(dbDir, "db.sql")
// 	saltFile = filepath.Join(dbDir, "salt")
//
// 	err = os.MkdirAll(dbDir, 0700)
// 	if err != nil {
// 		return err
// 	}
// 	// check for password
// 	if password != "" {
// 		log.Info("Connecting to encrypted data store")
// 		key, err := getKey(dbDir, password)
// 		if err != nil {
// 			log.WithFields(log.Fields{
// 				"error": err,
// 			}).Error("Failed to get key")
// 			return err
// 		}
//
// 		params = fmt.Sprintf("%s&_pragma_key=x'%X'&_pragma_cipher_page_size=4096", params, key)
// 	}
//
// 	log.Printf(dbFile)
// 	log.Printf(fmt.Sprintf("%s?%s", dbFile, params))
// 	db, err := sqlx.Open("sqlite3", fmt.Sprintf("%s?%s", dbFile, params))
// 	if err != nil {
// 		return err
// 	}
// 	_, err = db.Exec(messagesSchema)
// 	if err != nil {
// 		return err
// 	}
//
// 	_, err = db.Exec(sessionsSchema)
// 	if err != nil {
// 		return err
// 	}
//
// 	_, err = db.Exec(groupsSchema)
// 	if err != nil {
// 		return err
// 	}
// 	DS = &DataStore{Dbx: db}
// 	// RefreshContacts()
// 	return nil
// 	// return LoadMessagesFromDB()
// }

// Get salt for encrypted database stored at path

// Decrypt database and closes connection
func (ds *DataStore) Decrypt(dbPath string) error {
	log.Debugf("Decrypt Db")
	query := fmt.Sprintf("ATTACH DATABASE '%s' AS plaintext KEY '';", dbPath)
	if DS.Dbx == nil {
		log.Errorf("Dbx is nil")
	}
	_, err := DS.Dbx.Exec(query)
	if err != nil {
		return err
	}

	_, err = DS.Dbx.Exec("SELECT sqlcipher_export('plaintext');")
	if err != nil {
		return err
	}

	_, err = DS.Dbx.Exec("DETACH DATABASE plaintext;")
	if err != nil {
		return err
	}
	settings.SettingsModel.EncryptDatabase = false
	DS.Dbx = nil

	return nil
}

func (ds *DataStore) DBX() *sqlx.DB {
	return DS.Dbx
}
func (ds *DataStore) SetupDb(password string) bool {
	var err error
	dbDir = filepath.Join(config.DataDir, "db")
	err = os.MkdirAll(dbDir, 0700)
	DS, err = NewStorage(password)
	if err != nil {
		log.Debugf("Couldn't open db: " + err.Error())
		return false
	}
	LoadMessagesFromDB()
	qml.Changed(SessionsModel, &SessionsModel.Len)
	log.Printf("Db setup finished")

	return true
}
func (ds *DataStore) ResetDb() {
	dbDir = filepath.Join(config.DataDir, "db")
	dbFile = filepath.Join(dbDir, "db.sql")
	err := os.Remove(dbFile)
	if err != nil {
		log.Errorf(err.Error())
	}
	settings.SettingsModel.EncryptDatabase = false

}
func (ds *DataStore) DecryptDb(password string) bool {
	log.Info("DecryptDb: Decrypting database..")
	dbDir = filepath.Join(config.DataDir, "db")
	dbFile = filepath.Join(dbDir, "db.sql")
	tmp := filepath.Join(dbDir, "tmp.db")

	ds, err := NewStorage(password)
	if err != nil {
		return true
	}

	err = ds.Decrypt(tmp)
	if err != nil {
		log.Errorf(err.Error())
		return true
	}
	err = os.Remove(dbFile)
	if err != nil {
		log.Errorf(err.Error())

		return true
	}
	err = os.Rename(tmp, dbFile)
	if err != nil {
		log.Errorf(err.Error())
		return true
	}
	settings.SettingsModel.EncryptDatabase = false
	DS.Dbx = nil
	DS, err = NewStorage("")
	if err != nil {
		log.Debugf("Couldn't open db: " + err.Error())
		return false
	}
	return false
}
func (ds *DataStore) EncryptDb(password string) bool {
	log.Info("EncryptDb: Encrypting database..")
	dbDir = filepath.Join(config.DataDir, "db")
	dbFile = filepath.Join(dbDir, "db.sql")
	tmp := filepath.Join(dbDir, "tmp.db")

	ds, err := NewStorage("")
	if err != nil {
		return true
	}

	err = ds.Encrypt(tmp, password)
	if err != nil {
		log.Errorf("encrypting db: " + err.Error())

		return true
	}
	err = os.Remove(dbFile)
	if err != nil {
		log.Errorf(err.Error())

		return true
	}
	err = os.Rename(tmp, dbFile)
	if err != nil {
		log.Errorf(err.Error())
		return true
	}
	err = os.Remove(tmp)
	if err != nil {
		log.Errorf(err.Error())

		return true
	}
	settings.SettingsModel.EncryptDatabase = true
	DS.Dbx = nil
	DS.SetupDb(password)
	return false
}
func NewStorage(password string) (*DataStore, error) {
	log.Debugf("New Storage")
	// Set more restrictive umask to ensure database files are created 0600
	syscall.Umask(0077)

	dbDir = filepath.Join(config.DataDir, "db")
	err := os.MkdirAll(dbDir, 0700)
	if err != nil {
		log.Debugf(err.Error())

		return nil, err
	}

	dbFile := filepath.Join(dbDir, "db.sql")
	saltFile := ""

	if password != "" {
		log.Debugf("New Storage: Password not empty")

		saltFile = filepath.Join(dbDir, "salt")
	}

	ds, err := NewDataStore(dbFile, saltFile, password)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
func NewDataStore(dbPath, saltPath, password string) (*DataStore, error) {
	log.Debugf("NewDataStore")

	params := "_busy_timeout=5000&cache=shared"

	if password != "" && saltPath != "" {
		log.Info("Connecting to encrypted data store")
		key, err := getKey(saltPath, password)
		if err != nil {
			log.Errorf("Failed to get key: " + err.Error())
			return nil, err
		}
		log.Debugf("Connecting to encrypted data store finished")

		params = fmt.Sprintf("%s&_pragma_key=x'%X'&_pragma_cipher_page_size=4096", params, key)
	}

	db, err := sqlx.Open("sqlite3", fmt.Sprintf("%s?%s", dbPath, params))
	if err != nil {
		log.Errorf("Failed to open DB")
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Errorf("Failed to ping db")

		return nil, err
	}

	_, err = db.Exec(messagesSchema)

	if err != nil {
		log.Debugf("Failed exec messageSchema (Happens also on encrypted db)")

		return nil, err
	}

	_, err = db.Exec(sessionsSchema)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(groupsSchema)
	if err != nil {
		return nil, err
	}
	log.Debugf("NewDataStore finished")

	return &DataStore{Dbx: db}, nil
}
