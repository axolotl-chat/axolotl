package worker

import (
	"fmt"
	"strings"

	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/sender"
	"github.com/nanu-c/axolotl/app/store"
	"github.com/nanu-c/axolotl/app/ui"
	"github.com/signal-golang/textsecure"
	log "github.com/sirupsen/logrus"
)

type GroupRecord struct {
	ID      int64
	GroupID string
	Name    string
	Members string
	Avatar  []byte
	Active  bool
}

// FIXME: receive members as splice, blocked by https://github.com/nanu-c/qml-go/issues/137
func (a *TextsecureAPI) NewGroup(name string, members string) error {
	m := strings.Split(members, ",")
	group, err := textsecure.NewGroup(name, m)
	if err != nil {
		ui.ShowError(err, a.Websocket)
		return err
	}
	tel := a.Websocket.App.Config.TsConfig.Tel

	members = members + "," + tel
	store.Groups[group.Hexid] = &store.GroupRecord{
		GroupID: group.Hexid,
		Name:    name,
		Members: members,
	}
	g, err := store.SaveGroup(store.Groups[group.Hexid])
	if err != nil {
		ui.ShowError(err, a.Websocket)
		return err
	}
	session, err := store.SessionsModel.Get(g.ID)
	if err != nil {
		ui.ShowError(err, a.Websocket)
		return err
	}
	msg := session.Add(GroupUpdateMsg(append(m, tel), name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
	msg.Flags = helpers.MsgFlagGroupNew
	store.SaveMessage(msg)

	return nil

}
func (a *TextsecureAPI) UpdateGroup(hexid, name string, members string) error {
	g, ok := store.Groups[hexid]
	if !ok {
		log.Errorf("[textsecure] Update group: Unknown group id %s\n", hexid)
		return fmt.Errorf("[textsecure] Update group: Unknown group id %s\n", hexid)
	}
	dm, members := helpers.MembersDiffAndUnion(g.Members, members)
	store.Groups[hexid] = &store.GroupRecord{
		GroupID: hexid,
		Name:    name,
		Members: members,
		Active:  g.Active,
		Avatar:  g.Avatar,
	}
	g, err := store.UpdateGroup(store.Groups[hexid])
	if err != nil {
		ui.ShowError(err, a.Websocket)
		return err
	}
	session, err := store.SessionsModel.Get(g.ID)
	if err != nil {
		ui.ShowError(err, a.Websocket)
		return err
	}
	msg := session.Add(ui.GroupUpdateMsg(dm, name), "", []store.Attachment{}, "", true, store.ActiveSessionID)
	msg.Flags = helpers.MsgFlagGroupUpdate
	store.SaveMessage(msg)
	session.Name = name
	store.UpdateSession(session)
	go sender.SendMessage(session, msg, false)
	return nil
}

func (a *TextsecureAPI) LeaveGroup(hexid string) error {
	store.Groups[hexid].Active = false
	g, err := store.UpdateGroup(store.Groups[hexid])
	if err != nil {
		ui.ShowError(err, a.Websocket)
		return err
	}
	session, err := store.SessionsModel.Get(g.ID)
	if err != nil {
		ui.ShowError(err, a.Websocket)
		return err
	}
	msg := session.Add("You have left the group.", "", []store.Attachment{}, "", true, store.ActiveSessionID)
	msg.Flags = helpers.MsgFlagGroupLeave
	store.SaveMessage(msg)
	session.Active = false
	go sender.SendMessage(session, msg, false)
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

func (a *TextsecureAPI) GroupInfo(hexid string) string {
	s := ""
	if g, ok := store.Groups[hexid]; ok {
		for _, t := range strings.Split(g.Members, ",") {
			s += store.TelToName(t) + "\n\n"
		}
	}
	return s
}
