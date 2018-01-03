package worker

import (
	"fmt"
	"strings"

	qml "github.com/amlwwalker/qml"
	"github.com/morph027/textsecure"
	"github.com/nanu-c/textsecure-qml/lang"
	"github.com/nanu-c/textsecure-qml/models"
	"github.com/nanu-c/textsecure-qml/store"
	"github.com/nanu-c/textsecure-qml/ui"
)

type GroupRecord struct {
	ID      int64
	GroupID string
	Name    string
	Members string
	Avatar  []byte
	Active  bool
}

// FIXME: receive members as splice, blocked by https://github.com/amlwwalker/qml/issues/137
func (Api *TextsecureAPI) NewGroup(name string, members string) error {
	m := strings.Split(members, ",")
	group, err := textsecure.NewGroup(name, m)
	if err != nil {
		ui.ShowError(err)
		return err
	}

	members = members + "," + store.Config.Tel
	store.Groups[group.Hexid] = &models.GroupRecord{
		GroupID: group.Hexid,
		Name:    name,
		Members: members,
	}
	store.SaveGroup(store.Groups[group.Hexid])
	session := store.SessionsModel.Get(group.Hexid)
	msg := session.Add(GroupUpdateMsg(append(m, store.Config.Tel), name), "", "", "", true, Api.ActiveSessionID)
	msg.Flags = msgFlagGroupNew
	qml.Changed(msg, &msg.Flags)
	store.SaveMessage(msg)

	return nil

}
func (Api *TextsecureAPI) UpdateGroup(hexid, name string, members string) error {
	g, ok := store.Groups[hexid]
	if !ok {
		return fmt.Errorf("Unknown group id %s\n", hexid)
	}
	dm, members := models.MembersDiffAndUnion(g.Members, members)
	store.Groups[hexid] = &models.GroupRecord{
		GroupID: hexid,
		Name:    name,
		Members: members,
		Active:  g.Active,
		Avatar:  g.Avatar,
	}
	store.UpdateGroup(store.Groups[hexid])
	session := store.SessionsModel.Get(hexid)
	msg := session.Add(ui.GroupUpdateMsg(dm, name), "", "", "", true, Api.ActiveSessionID)
	msg.Flags = msgFlagGroupUpdate
	qml.Changed(msg, &msg.Flags)
	store.SaveMessage(msg)
	session.Name = name
	qml.Changed(session, &session.Name)
	store.UpdateSession(session)
	go SendMessage(session, msg)
	return nil
}

func (Api *TextsecureAPI) LeaveGroup(hexid string) error {
	session := store.SessionsModel.Get(hexid)
	msg := session.Add(lang.YouLeftGroup, "", "", "", true, Api.ActiveSessionID)
	msg.Flags = msgFlagGroupLeave
	qml.Changed(msg, &msg.Flags)
	store.SaveMessage(msg)
	session.Active = false
	qml.Changed(session, &session.Active)
	store.Groups[hexid].Active = false
	err := store.UpdateGroup(store.Groups[hexid])
	go SendMessage(session, msg)
	return err
}
func GroupUpdateMsg(tels []string, title string) string {
	s := ""
	if len(tels) > 0 {
		for _, t := range tels {
			s += store.TelToName(t) + ", "
		}
		s = s[:len(s)-2] + " joined the group. "
	}

	return s + "Title is now '" + title + "'."
}

func (Api *TextsecureAPI) GroupInfo(hexid string) string {
	s := ""
	if g, ok := store.Groups[hexid]; ok {
		for _, t := range strings.Split(g.Members, ",") {
			s += store.TelToName(t) + "\n\n"
		}
	}
	return s
}
