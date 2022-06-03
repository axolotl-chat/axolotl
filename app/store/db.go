package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/settings"
	log "github.com/sirupsen/logrus"
)

//var DS *DataStore // TODO

type DataStore struct {
	Dbx *sqlx.DB
}

var (
	dbDir    string // TODO
	dbFile   string // TODO
	saltFile string // TODO

	sessionsSchema = "CREATE TABLE IF NOT EXISTS sessions (id INTEGER PRIMARY KEY, name text, tel text, isgroup boolean, last string, timestamp integer, ctype integer, unread integer default 0, notification boolean default 1, expireTimer integer default 0, type integer NOT NULL DEFAULT 0, uuid string  NOT NULL DEFAULT 0, groupJoinStatus integer NOT NULL DEFAULT 0)"
	sessionsInsert = "INSERT OR REPLACE INTO sessions (name, tel, isgroup, last, ctype, timestamp, notification, expireTimer, type, uuid, groupJoinStatus ) VALUES ( :name, :tel, :isgroup, :last, :ctype, :timestamp, :notification, :expireTimer, :type, :uuid, :groupJoinStatus)"
	sessionsSelect = "SELECT * FROM sessions ORDER BY timestamp DESC"

	messagesSchema                 = "CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY, sid integer, source text, srcUUID string NOT NULL DEFAULT 0, message text, outgoing boolean, sentat integer, receivedat integer, ctype integer, attachment string, issent boolean, isread boolean, flags integer default 0, sendingError boolean, expireTimer integer default 0, receipt boolean default 0, statusMessage boolean default 0, quoteId integer NOT NULL default -1)"
	messagesInsert                 = "INSERT INTO messages (sid, source, srcUUID, message, outgoing, sentat, receivedat, ctype, attachment, issent, isread, flags, sendingError, expireTimer, statusMessage, quoteID) VALUES (:sid, :source, :srcUUID, :message, :outgoing, :sentat, :receivedat, :ctype, :attachment, :issent, :isread, :flags, :sendingError, :expireTimer, :statusMessage, :quoteId)"
	messagesSelectWhere            = "SELECT * FROM messages WHERE sid = ? ORDER BY sentat DESC LIMIT 20"
	messagesSelectWhereMore        = "SELECT * FROM messages WHERE sid = ? AND id< ? ORDER BY sentat DESC LIMIT 20"
	messagesSelectWhereLastMessage = "SELECT * FROM messages WHERE sid = ? ORDER BY sentat DESC LIMIT 1"
	messagesDelete                 = "DELETE FROM messages WHERE id = ?"
	messagesReceiptSent            = "UPDATE messages SET isSent=1 WHERE sentat = ?"

	groupsSchema = "CREATE TABLE IF NOT EXISTS groups (id INTEGER PRIMARY KEY, groupid TEXT, name TEXT, members TEXT, avatar BLOB, avatarid INTEGER, avatar_key BLOB, avatar_content_type TEXT, relay TEXT, active INTEGER DEFAULT 1, type INTEGER DEFAULT 1, joinStatus INTEGER DEFAULT 1)"
	groupsInsert = "INSERT OR REPLACE INTO groups (groupid, name, members, avatar, type) VALUES (:groupid, :name, :members, :avatar, :type)"
	groupsUpdate = "UPDATE groups SET members = :members, name = :name, avatar = :avatar, active = :active,  type = :type WHERE groupid = :groupid"
	groupsSelect = "SELECT groupid, name, members, avatar, active FROM groups"
	groupsDelete = "DELETE FROM groups WHERE groupid = ?"
)

// Get salt for encrypted database stored at path

// Decrypt database and closes connection
func (ds *DataStore) Decrypt(dbPath string) error {
	log.Debugf("[axolotl] Decrypt Db")
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

// SetupDb tries to decrypt the database and runs the migrations
func (ds *DataStore) SetupDb(password string) bool {
	var err error
	dbDir = GetDbDir()
	log.Debugln("[axolotl] openDb: " + dbDir)

	err = os.MkdirAll(dbDir, 0700)
	if err != nil {
		log.Debugln("[axolotl] setupDb: Couldn't make dir: " + err.Error())
		return false
	}
	DS, err = NewStorage(password)
	if err != nil {
		log.Debugln("[axolotl] setupDb: Couldn't open db: " + err.Error())
		return false
	}
	UpdateSessionTable()
	UpdateMessagesTable_v_0_7_8()
	UpdateSessionTable_v_0_7_8()
	UpdateSessionTable_v_0_9_0()
	UpdateSessionTable_v_0_9_5()
	updateGroupTable_v_0_9_10()
	updateSessionTable_joinStatus_v_0_9_10()

	err = LoadChats()
	if err != nil {
		log.Errorln("[axolotl]  SetupDB: ", err)

	}
	//qml.Changed(SessionsModel, &SessionsModel.Len)
	log.Printf("[axolotl] Db setup finished")

	return true
}

// ResetDb removes the database file and resets the config for encrypted database.
func (ds *DataStore) ResetDb() {
	dbDir = GetDbDir()
	dbFile = GetDbFile()
	err := os.Remove(dbFile)
	if err != nil {
		log.Errorf(err.Error())
	}
	settings.SettingsModel.EncryptDatabase = false

}
func (ds *DataStore) DecryptDb(password string) bool {
	log.Info("DecryptDb: Decrypting database..")
	dbDir = GetDbDir()
	dbFile = GetDbFile()
	tmp := GetDbTmpFile()

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
	settings.SaveSettings(settings.SettingsModel)

	DS.Dbx = nil
	DS, err = NewStorage("")
	if err != nil {
		log.Debugf("Couldn't open db: " + err.Error())
		return false
	}
	return false
}
func (ds *DataStore) EncryptDb(password string) bool {
	log.Info("[axolotl] EncryptDb: Encrypting database..")
	dbDir = GetDbDir()
	dbFile = GetDbFile()
	tmp := GetDbTmpFile()

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
	settings.SaveSettings(settings.SettingsModel)
	DS.Dbx = nil
	DS.SetupDb(password)
	return false
}

// NewStorage
func NewStorage(password string) (*DataStore, error) {
	// Set more restrictive umask to ensure database files are created 0600
	// syscall.Umask(0077)

	dbDir = GetDbDir()
	err := os.MkdirAll(dbDir, 0700)
	if err != nil {
		log.Debugln("[axolotl] error open db ", err.Error())

		return nil, err
	}

	dbFile := GetDbFile()
	saltFile := ""

	if password != "" {
		saltFile = filepath.Join(dbDir, "salt")
	}

	ds, err := NewDataStore(dbFile, saltFile, password)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
func NewDataStore(dbPath, saltPath, password string) (*DataStore, error) {
	log.Debugf("[axolotl] NewDataStore")

	params := "_busy_timeout=5000&cache=shared"

	if password != "" && saltPath != "" {
		log.Info("[axolotl] Connecting to encrypted data store")
		key, err := getKey(saltPath, password)
		if err != nil {
			log.Errorf("[axolotl] Failed to get key: " + err.Error())
			return nil, err
		}
		log.Debugf("[axolotl] Connecting to encrypted data store finished")

		params = fmt.Sprintf("%s&_pragma_key=x'%X'&_pragma_cipher_page_size=4096", params, key)
	}

	db, err := sqlx.Open("sqlite3", fmt.Sprintf("%s?%s", dbPath, params))
	if err != nil {
		log.Errorf("[axolotl] Failed to open DB")
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Errorf("[axolotl] Failed to ping db")

		return nil, err
	}

	_, err = db.Exec(messagesSchema)

	if err != nil {
		log.Debugln("[axolotl] Failed exec messageSchema (Happens also on encrypted db):", err)

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
	log.Debugf("[axolotl] NewDataStore finished")

	return &DataStore{Dbx: db}, nil
}

func GetDbDir() string {
	dataDir := config.GetDataDir()
	return filepath.Join(dataDir, "db")
}

func GetDbFile() string {
	dbDir := GetDbDir()
	return filepath.Join(dbDir, "db.sql")
}

func GetDbTmpFile() string {
	dbDir := GetDbDir()
	return filepath.Join(dbDir, "tmp.db")
}
