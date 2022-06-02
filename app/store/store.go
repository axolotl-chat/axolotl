package store

type Store struct {
	ActiveSessionID int64
	Contacts        *Contacts
	DS              *DataStore
	AllGroups       []*GroupRecord
	Groups          map[string]*GroupRecord
	LinkedDevices   *LinkedDevices
	AllSessions     []*Session
	topSession      int64
}
