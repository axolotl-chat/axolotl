package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/qml.v1"
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

func setupDB() error {
	var err error

	dbDir = filepath.Join(dataDir, "db")
	dbFile = filepath.Join(dbDir, "db.sql")
	err = os.MkdirAll(dbDir, 0700)
	if err != nil {
		return err
	}

	db, err = sqlx.Open("sqlite3", dbFile)
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

	migrations()

	return loadMessagesFromDB()
}

func saveSession(s *Session) error {
	res, err := db.NamedExec(sessionsInsert, s)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	s.ID = id
	return err
}

func saveGroup(g *GroupRecord) error {
	res, err := db.NamedExec(groupsInsert, g)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	g.ID = id
	return nil
}

func saveMessage(m *Message) error {
	res, err := db.NamedExec(messagesInsert, m)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	m.ID = id
	return nil
}

func updateMessageSent(m *Message) error {
	_, err := db.NamedExec("UPDATE messages SET issent = :issent, sentat = :sentat WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}

func updateMessageRead(m *Message) error {
	_, err := db.NamedExec("UPDATE messages SET isread = :isread WHERE id = :id", m)
	if err != nil {
		return err
	}
	return err
}

func updateSession(s *Session) error {
	_, err := db.NamedExec("UPDATE sessions SET name = :name, timestamp = :timestamp, ctype = :ctype, last = :last, unread = :unread WHERE id = :id", s)
	if err != nil {
		return err
	}
	return err
}

func updateGroup(g *GroupRecord) error {
	_, err := db.NamedExec(groupsUpdate, g)
	if err != nil {
		return err
	}
	return err
}

func deleteGroup(hexid string) error {
	_, err := db.Exec(groupsDelete, hexid)
	return err
}

var allSessions []*Session
var allGroups []*GroupRecord

func loadMessagesFromDB() error {
	err := db.Select(&allGroups, groupsSelect)
	if err != nil {
		return err
	}
	for _, g := range allGroups {
		groups[g.GroupID] = g
	}

	err = db.Select(&allSessions, sessionsSelect)
	if err != nil {
		return err
	}
	for _, s := range allSessions {
		s.When = humanizeTimestamp(s.Timestamp)
		s.Active = !s.IsGroup || (groups[s.Tel] != nil && groups[s.Tel].Active)
		sessionsModel.sessions = append(sessionsModel.sessions, s)
		sessionsModel.Len++
		err = db.Select(&s.messages, messagesSelectWhere, s.ID)
		s.Len = len(s.messages)
		if err != nil {
			return err
		}
		for _, m := range s.messages {
			m.HTime = humanizeTimestamp(m.SentAt)
		}
	}
	return nil
}

func deleteSession(tel string) error {
	s := sessionsModel.Get(tel)
	_, err := db.Exec("DELETE FROM messages WHERE sid = ?", s.ID)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM sessions WHERE id = ?", s.ID)
	if err != nil {
		return err
	}
	index := sessionsModel.GetIndex(s.Tel)
	sessionsModel.sessions = append(sessionsModel.sessions[:index], sessionsModel.sessions[index+1:]...)
	sessionsModel.Len--
	qml.Changed(sessionsModel, &sessionsModel.Len)
	return nil
}

func addFlagsColumnToMessages() error {
	_, err := db.Exec("ALTER TABLE messages ADD COLUMN flags INTEGER DEFAULT 0")
	return err
}

func addUnreadColumnToSessions() error {
	_, err := db.Exec("ALTER TABLE sessions ADD COLUMN unread INTEGER DEFAULT 0")
	return err
}

func migrateOnce(name string, f func()) error {
	path := filepath.Join(dbDir, "migrated_to_"+name)
	if !exists(path) {
		f()
	}
	_, err := os.Create(path)
	return err
}

// Columns messages.flags and sessions.unread were introduced in 0.3.7
func migrate_to_0_3_7() {
	err := addFlagsColumnToMessages()
	if err != nil {
		log.Println(err)
	}
	err = addUnreadColumnToSessions()
	if err != nil {
		log.Println(err)
	}
}

func migrations() {
	err := migrateOnce("0_3_7", migrate_to_0_3_7)
	if err != nil {
		log.Println(err)
	}
}
