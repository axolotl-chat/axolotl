package store

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	qml "github.com/amlwwalker/qml"
	"golang.org/x/crypto/scrypt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db     *sqlx.DB
	dbDir  string
	dbFile string

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

func SetupDB(saltPath, password string) error {
	params := "_busy_timeout=5000&cache=shared"
	var err error

	dbDir = filepath.Join(DataDir, "db")
	dbFile = filepath.Join(dbDir, "db.sql")
	err = os.MkdirAll(dbDir, 0700)
	if err != nil {
		return err
	}
	// check for password
	if password != "" && saltPath != "" {
		log.Info("Connecting to encrypted data store")
		key, err := getKey(saltPath, password)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to get key")
			return err
		}

		params = fmt.Sprintf("%s&_pragma_key=x'%X'&_pragma_cipher_page_size=4096", params, key)
	}

	db, err = sqlx.Open("sqlite3", fmt.Sprintf("%s?%s", dbFile, params))
	if err != nil {
		return err
	}
	_, err = db.Exec(messagesSchema)
	if err != nil {
		return err
	}

	_, err = db.Exec(sessionsSchema)
	if err != nil {
		return err
	}

	_, err = db.Exec(groupsSchema)
	if err != nil {
		return err
	}

	return LoadMessagesFromDB()
}

// Get salt for encrypted database stored at path
func getSalt(path string) ([]byte, error) {
	salt := make([]byte, 8)

	if _, err := os.Stat(path); err == nil {
		salt, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(path, salt, 0600)
		if err != nil {
			return nil, err
		}
	}

	return salt, nil
}

// Get raw key data for use with sqlcipher
func getKey(saltPath, password string) ([]byte, error) {
	salt, err := getSalt(saltPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to get salt")
		return nil, err
	}

	return scrypt.Key([]byte(password), salt, 16384, 8, 1, 32)
}

func (s *Sessions) GetIndex(tel string) int {
	for i, ses := range s.Sess {
		if ses.Tel == tel {
			return i
		}
	}
	return -1
}

var topSession string

func (s *Session) moveToTop() {
	if topSession == s.Tel {
		return
	}

	index := SessionsModel.GetIndex(s.Tel)
	SessionsModel.Sess = append([]*Session{s}, append(SessionsModel.Sess[:index], SessionsModel.Sess[index+1:]...)...)

	// force a length change update
	SessionsModel.Len--
	qml.Changed(SessionsModel, &SessionsModel.Len)
	SessionsModel.Len++
	qml.Changed(SessionsModel, &SessionsModel.Len)

	topSession = s.Tel
}
