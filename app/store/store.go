package store

type Store struct {
	ActiveSessionID int64
	Contacts        *Contacts
	DS              *DataStore
	AllGroups       []*GroupRecord
	Groups          map[string]*GroupRecord
	LinkedDevices   *LinkedDevices
	AllSessions     []*Session
	Sessions        *Sessions
	TopSession      int64
}
