package store

import (
	"github.com/signal-golang/textsecure"
	"github.com/signal-golang/textsecure/profiles"
	log "github.com/sirupsen/logrus"
)

var (
	recipientsSchema   = "CREATE TABLE IF NOT EXISTS recipients (id integer PRIMARY KEY,e164 text,uuid text,username text,email text,is_blocked Bool,profile_key blob,profile_key_credential blob,profile_given_name text,profile_family_name text,profile_joined_name text,signal_profile_avatar text,profile_sharing_enabled Bool,last_profile_fetch DATETIME DEFAULT CURRENT_TIMESTAMP,unidentified_access_mode Bool,storage_service_id blob,storage_proto blob,capabilities integer,last_session_reset DATETIME DEFAULT CURRENT_TIMESTAMP);"
	recipientInsert    = "INSERT or REPLACE INTO recipients (id, e164, uuid, username, email, is_blocked, profile_key, profile_key_credential, profile_given_name, profile_family_name, profile_joined_name, signal_profile_avatar, profile_sharing_enabled, last_profile_fetch, unidentified_access_mode, storage_service_id, storage_proto, capabilities, last_session_reset) VALUES (:id, :e164, :uuid, :username, :email, :is_blocked, :profile_key, :profile_key_credential, :profile_given_name, :profile_family_name, :profile_joined_name, :signal_profile_avatar, :profile_sharing_enabled, :last_profile_fetch, :unidentified_access_mode, :storage_service_id, :storage_proto, :capabilities, :last_session_reset)"
	recipientGetById   = "SELECT * FROM recipients WHERE id = ?"
	recipientGetByUUID = "SELECT * FROM recipients WHERE uuid = ?"
	recipientGetByE164 = "SELECT * FROM recipients WHERE e164 = ?"
)

type Recipient struct {
	Id                     int64  `db:"id"`
	E164                   string `db:"e164"`
	UUID                   string `db:"uuid"`
	Username               string `db:"username"`
	Email                  string `db:"email"`
	IsBlocked              bool   `db:"is_blocked"`
	ProfileKey             []byte `db:"profile_key"`
	ProfileKeyCredential   []byte `db:"profile_key_credential"`
	ProfileGivenName       string `db:"profile_given_name"`
	ProfileFamilyName      string `db:"profile_family_name"`
	ProfileJoinedName      string `db:"profile_joined_name"`
	SignalProfileAvatar    string `db:"signal_profile_avatar"`
	ProfileSharingEnabled  bool   `db:"profile_sharing_enabled"`
	LastProfileFetch       string `db:"last_profile_fetch"`
	UnidentifiedAccessMode bool   `db:"unidentified_access_mode"`
	StorageServiceId       []byte `db:"storage_service_id"`
	StorageProto           []byte `db:"storage_proto"`
	Capabilities           int    `db:"capabilities"`
	LastSessionReset       string `db:"last_session_reset"`
	//todo add fields for profile
	// Avatar                  string `db:"avatar"`
	// About                  string `db:"about"`
	// AboutEmoji             string `db:"about_emoji"`
}

type Recipients struct {
	Recipients []*Recipient `db:"recipients"`
}

var RecipientsModel = &Recipients{
	Recipients: make([]*Recipient, 0),
}

// CreateRecipient creates a recipient
func (r *Recipients) CreateRecipient(recipient *Recipient) (*Recipient, error) {
	log.Debugln("[axolotl] Creating recipient", recipient.UUID)
	storedRecipeit := r.GetRecipientByUUID(recipient.UUID)
	if storedRecipeit != nil {
		return storedRecipeit, nil
	}
	res, err := DS.Dbx.NamedExec(recipientInsert, recipient)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	recipient.Id = id
	recipient.UpdateProfile()
	return recipient, err
}

// this is only used in the v1.6.0 migration because we migrate before we initialize the signal server
func (r *Recipients) CreateRecipientWithoutProfileUpdate(recipient *Recipient) (*Recipient, error) {
	log.Debugln("[axolotl] Creating recipient", recipient.UUID)
	storedRecipeit := r.GetRecipientByUUID(recipient.UUID)
	if storedRecipeit != nil {
		return storedRecipeit, nil
	}
	res, err := DS.Dbx.NamedExec(recipientInsert, recipient)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	recipient.Id = id
	return recipient, err
}

// GetRecipientById returns a recipient by id
func (r *Recipients) GetRecipientById(id int64) *Recipient {
	recipient := &Recipient{}
	err := DS.Dbx.Get(recipient, recipientGetById, id)
	if err != nil {
		return nil
	}
	return recipient
}

// GetRecipientByE164 returns a recipient by e164
func (r *Recipients) GetRecipientByE164(e164 string) *Recipient {
	recipient := &Recipient{}
	err := DS.Dbx.Get(recipient, recipientGetByE164, e164)
	if err != nil {
		return nil
	}
	return recipient
}

// GetRecipientByUUID returns a recipient by uuid
func (r *Recipients) GetRecipientByUUID(uuid string) *Recipient {
	recipient := &Recipient{}
	err := DS.Dbx.Get(recipient, recipientGetByUUID, uuid)
	if err != nil {
		return nil
	}
	return recipient
}

// SaveRecipient saves a recipient
func (r *Recipient) SaveRecipient() error {
	_, err := DS.Dbx.NamedExec(recipientInsert, r)
	return err
}

// UpdateProfile updates a recipient's profile
func (r *Recipient) UpdateProfile() error {
	log.Debugln("[axolotl] Updating profile for recipient", r.UUID)
	var profile *profiles.Profile
	var err error
	if r.ProfileKey == nil {
		profile, err = textsecure.GetProfileByUUID(r.UUID)
		if err != nil {
			return err
		}
		if profile != nil {
			if profile.IdentityKey != "" {
				r.ProfileKey = []byte(profile.IdentityKey)
			}
			if profile != nil && len(profile.Name) > 0 {
				r.Username = profile.Name
			}
		}
	}
	r.SaveRecipient()
	return nil
}
