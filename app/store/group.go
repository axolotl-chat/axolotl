package store

type GroupRecord struct {
	ID      int64
	GroupID string
	Name    string
	Members string
	Avatar  []byte
	Active  bool
}

var AllGroups []*GroupRecord
var Groups = map[string]*GroupRecord{}

func UpdateGroup(g *GroupRecord) error {
	_, err := DS.Dbx.NamedExec(groupsUpdate, g)
	if err != nil {
		return err
	}
	return err
}

func DeleteGroup(hexid string) error {
	_, err := DS.Dbx.Exec(groupsDelete, hexid)
	return err
}
func SaveGroup(g *GroupRecord) error {
	res, err := DS.Dbx.NamedExec(groupsInsert, g)
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
