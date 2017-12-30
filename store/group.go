package store

import "github.com/nanu-c/textsecure-qml/models"

func UpdateGroup(g *models.GroupRecord) error {
	_, err := db.NamedExec(groupsUpdate, g)
	if err != nil {
		return err
	}
	return err
}

func DeleteGroup(hexid string) error {
	_, err := db.Exec(groupsDelete, hexid)
	return err
}
func SaveGroup(g *models.GroupRecord) error {
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
func FetchAllGroups() error {
	return nil
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
