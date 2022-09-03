package store

import (
	"github.com/nanu-c/axolotl/app/config"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	signalservice "github.com/signal-golang/textsecure/protobuf"
)

var (
	groupV2MembersSchema            = "CREATE TABLE IF NOT EXISTS groupsv2members (id INTEGER PRIMARY KEY,group_v2_id TEXT,recipient_id INTEGER,member_since DATETIME DEFAULT CURRENT_TIMESTAMP,joined_at_revision INTEGER,role INTEGER)"
	groupV2MemberInsert             = "INSERT INTO groupsV2Members (group_v2_id, recipient_id, member_since, joined_at_revision, role) VALUES (:group_v2_id, :recipient_id, :member_since, :joined_at_revision, :role)"
	groupV2MemberGetMembersForGroup = "SELECT * FROM groupsv2members WHERE group_v2_id = :id"
	groupsV2Schema                  = `CREATE TABLE IF NOT EXISTS groupsv2 (
		id TEXT PRIMARY KEY,
		name TEXT,
		master_key TEXT,
		revision INTEGER,
		invite_link_password BLOB,
		access_required_for_attributes INTEGER,
		access_required_for_members INTEGER,
		join_status INTEGER DEFAULT 0);`
	groupV2Insert = "INSERT OR REPLACE INTO groupsv2 (id, name, master_key, revision, invite_link_password, access_required_for_attributes, access_required_for_members, join_status) VALUES (:id, :name, :master_key, :revision, :invite_link_password, :access_required_for_attributes, :access_required_for_members, :join_status)"
)
var (
	GroupJoinStatusJoined  = 0
	GroupJoinStatusInvited = 1
	GroupJoinStatusDeleted = 2
)

// GroupV2 is a group, groupsv1 are deprecated
type GroupV2 struct {
	Id                          string `json:"id" db:"id"`
	Name                        string `json:"name" db:"name"`
	MasterKey                   string `json:"master_key" db:"master_key"`
	Revision                    int    `json:"revision" db:"revision"`
	InviteLinkPassword          string `json:"invite_link_password" db:"invite_link_password"`
	AccessRequiredForAttributes int    `json:"access_required_for_attributes" db:"access_required_for_attributes"`
	AccessRequiredForMembers    int    `json:"access_required_for_members" db:"access_required_for_members"`
	JoinStatus                  int    `json:"join_status" db:"join_status"`
}

// GroupV2Member represents a group member
type GroupV2Member struct {
	Id               int    `json:"id" db:"id"`
	GroupV2Id        string `json:"group_v2_id" db:"group_v2_id"`
	RecipientId      int    `json:"recipient_id" db:"recipient_id"`
	MemberSince      string `json:"member_since" db:"member_since"`
	JoinedAtRevision int    `json:"joined_at_revision" db:"joined_at_revision"`
	Role             int    `json:"role" db:"role"`
}

type GroupV2s struct {
	Groups []GroupV2 `json:"groups"`
}

var GroupV2sModel = GroupV2s{
	Groups: []GroupV2{},
}

func (g GroupV2s) Create(group *GroupV2) (*GroupV2, error) {
	_, err := DS.Dbx.NamedExec(groupV2Insert, group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// GetGroupById returns a group by id
func (g GroupV2s) GetGroupById(id string) (*GroupV2, error) {
	var group GroupV2
	err := DS.Dbx.Get(&group, "SELECT * FROM groupsv2 WHERE id = :id", id)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGroupMembers returns all members of a group
func (g *GroupV2) GetGroupMembers() ([]GroupV2Member, error) {
	var members []GroupV2Member
	err := DS.Dbx.Select(&members, "SELECT * FROM groupsv2members WHERE group_v2_id = :group_v2_id", g.Id)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetGroupMembers as recipients for a group
func (g *GroupV2) GetGroupMembersAsRecipients() ([]*Recipient, error) {
	members, err := g.GetGroupMembers()
	if err != nil {
		return nil, err
	}

	var recipients []*Recipient
	for _, member := range members {
		recipient := RecipientsModel.GetRecipientById(int64(member.RecipientId))
		if recipient != nil {
			recipients = append(recipients, recipient)
		}
	}
	return recipients, nil
}

// Get all groups
func (g GroupV2s) GetGroups() ([]GroupV2, error) {
	var groups []GroupV2
	err := DS.Dbx.Select(&groups, "SELECT * FROM groupsv2")
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// UpdateGroup updates a group
func (g *GroupV2) UpdateGroup() error {
	_, err := DS.Dbx.NamedExec(groupV2Insert, g)
	if err != nil {
		return err
	}
	return nil
}

// UpdateGroupMembers updates all group members
func (g *GroupV2) UpdateGroupMembers(members []GroupV2Member) error {
	// delete all old members
	err := g.DeleteMembers()
	if err != nil {
		return err
	}
	for _, member := range members {
		_, err := DS.Dbx.NamedExec(groupV2MemberInsert, member)
		if err != nil {
			return err
		}
	}
	return nil
}
func (g *GroupV2) AddGroupMembers(members []*signalservice.DecryptedMember) error {
	for _, member := range members {
		var err error
		id, err := uuid.FromBytes(member.Uuid)
		if err != nil {
			return err
		}
		recipient := RecipientsModel.GetRecipientByUUID(id.String())
		if recipient == nil {
			recipient, err = RecipientsModel.CreateRecipient(&Recipient{
				UUID:       id.String(),
				ProfileKey: member.ProfileKey,
				E164:       string(member.Pni),
			})
			if err != nil {
				return err
			}
		}
		if !g.IsMember(recipient) {
			err = g.AddMember(recipient)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// DeleteGroup deletes a group
func (g *GroupV2) Delete() error {
	err := g.DeleteMembers()
	if err != nil {
		return err
	}
	// delete all old members
	_, err = DS.Dbx.Exec("DELETE FROM groupsv2 WHERE id = :id", g.Id)
	if err != nil {
		return err
	}
	return nil
}

// DeleteMembers deletes all members of a group
func (g *GroupV2) DeleteMembers() error {
	// delete all old members
	_, err := DS.Dbx.Exec("DELETE FROM groupsv2members WHERE group_v2_id = :group_v2_id", g.Id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateGroupAction updates a group with a new action
func (g *GroupV2) UpdateGroupAction(action *signalservice.DecryptedGroupChange) error {
	log.Infoln("[axolotl] UpdateGroupAction to ", action.GetRevision())
	g.Revision = int(action.GetRevision())
	if action.GetNewTitle() != nil {
		g.Name = action.GetNewTitle().GetValue()
	}
	if action.GetNewInviteLinkPassword() != nil {
		g.InviteLinkPassword = string(action.GetNewInviteLinkPassword())
	}
	if len(action.NewMembers) > 0 {
		err := g.AddGroupMembers(action.NewMembers)
		if err != nil {
			return err
		}
		for i := range action.NewMembers {

			member := action.NewMembers[i]
			memberUUID := uuid.FromBytesOrNil(member.Uuid)
			log.Debugln("[axolotl] New member", memberUUID.String())
			if memberUUID.String() == config.Config.UUID {
				log.Debugln("[axolotl] I was added to group ", g.Id)
				g.JoinStatus = GroupJoinStatusJoined
			}
		}
	}
	if len(action.DeleteMembers) > 0 {
		for i := range action.DeleteMembers {
			member := action.DeleteMembers[i]
			memberUUID := uuid.FromBytesOrNil(member)
			recipient := RecipientsModel.GetRecipientByUUID(memberUUID.String())
			if recipient == nil {
				log.Debugln("[axolotl] Recipient not found:", memberUUID)
			} else {
				err := g.DeleteMember(recipient)
				if err != nil {
					return err
				}
			}
			if memberUUID.String() == config.Config.UUID {
				log.Debugln("[axolotl] I was removed from group ", g.Id)
				g.JoinStatus = GroupJoinStatusDeleted
			}
		}
	}
	// Todo: update other fields
	g.UpdateGroup()
	return nil
}

// AddRecipient adds a recipient to a group
func (g *GroupV2) AddMember(recipient *Recipient) error {
	_, err := DS.Dbx.NamedExec(groupV2MemberInsert, GroupV2Member{
		GroupV2Id:        g.Id,
		RecipientId:      int(recipient.Id),
		JoinedAtRevision: g.Revision,
		Role:             1,
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteMember deletes a member from a group
func (g *GroupV2) DeleteMember(recipient *Recipient) error {
	_, err := DS.Dbx.Exec("DELETE FROM groupsv2members WHERE group_v2_id = :group_v2_id AND recipient_id = :recipient_id", g.Id, int(recipient.Id))
	if err != nil {
		return err
	}
	return nil
}

// IsMember returns true if the user is a member of the group
func (g *GroupV2) IsMember(recipient *Recipient) bool {
	var count int
	err := DS.Dbx.Get(&count, "SELECT COUNT(*) FROM groupsv2members WHERE group_v2_id = :group_v2_id AND recipient_id = :recipient_id", g.Id, int(recipient.Id))
	if err != nil {
		return false
	}
	return count > 0
}
