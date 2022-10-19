package store

// Session is Deprecated: use store.SessionsV2Model instead
// It used to define how a session looks like
type Session struct {
	ID              int64
	UUID            string `db:"uuid"`
	Name            string
	Tel             string
	IsGroup         bool  `db:"isgroup"`
	Type            int32 // describes the type of the session, wether it's a private conversation or groupv1 or groupv2
	Last            string
	Timestamp       uint64
	When            string
	CType           int
	Messages        []*Message
	Unread          int
	Active          bool
	Len             int
	Notification    bool
	ExpireTimer     uint32 `db:"expireTimer"`
	GroupJoinStatus int    `db:"groupJoinStatus"`
}
type MessageList struct {
	ID       int64
	Session  *SessionV2
	Messages []*Message
}
type Sessions struct {
	Sess       []*Session
	ActiveChat string
	Len        int
	Type       string
}

// SessionTypes
const (
	invalidSession                  = -1
	invalidQuote                    = -1
	SessionTypePrivateChat    int32 = 0
	SessionTypeGroupV1        int32 = 1
	SessionTypeGroupV2        int32 = 2
	SessionTypeGroupV2Invited int32 = 3
)
