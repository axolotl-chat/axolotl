package store

import log "github.com/sirupsen/logrus"

type GroupRecord struct {
	ID      int64
	Uuid    string
	GroupID string
	Name    string
	Members string
	Avatar  []byte
	Active  bool
	Type    int
}

const (
	GroupRecordTypeGroupv1 = 0
	GroupRecordTypeGroupv2 = 1
)

var AllGroups []*GroupRecord
var Groups = map[string]*GroupRecord{}

func UpdateGroup(g *GroupRecord) (*GroupRecord, error) {
	res, err := DS.Dbx.NamedExec(groupsUpdate, g)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	g.ID = id
	return g, err
}

// DeleteGroup deletes a group from the database
func DeleteGroup(hexid string) error {
	_, err := DS.Dbx.Exec(groupsDelete, hexid)
	return err
}
func SaveGroup(g *GroupRecord) (*GroupRecord, error) {
	log.Debugln("[axolotl] saveGroup ", g.Uuid)
	res, err := DS.Dbx.NamedExec(groupsInsert, g)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	g.ID = id
	return g, nil
}
func FetchAllGroups() error {
	return nil
}
func GetGroupById(hexid string) *GroupRecord {
	return Groups[hexid]
}
func GroupUpdateMsg(tels []string, title string) string {
	s := ""
	if len(tels) > 0 {
		for _, t := range tels {
			s += TelToName(t) + ", "
		}
		s = s[:len(s)-2] + " joined the group. "
	}

	return s + "Title is now '" + title + "'."
}
