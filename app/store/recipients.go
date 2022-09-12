package store

import (
	"time"

	"github.com/signal-golang/textsecure"
	"github.com/signal-golang/textsecure/contacts"
	"github.com/signal-golang/textsecure/profiles"
	log "github.com/sirupsen/logrus"
)

var (
	recipientsSchema   = "CREATE TABLE IF NOT EXISTS recipients (id integer PRIMARY KEY,e164 text,uuid text,username text,email text,is_blocked Bool,profile_key blob,profile_key_credential blob,profile_given_name text,profile_family_name text,profile_joined_name text,signal_profile_avatar text,profile_sharing_enabled Bool,last_profile_fetch DATETIME DEFAULT CURRENT_TIMESTAMP,unidentified_access_mode Bool,storage_service_id blob,storage_proto blob,capabilities integer,last_session_reset DATETIME DEFAULT CURRENT_TIMESTAMP, about text, about_emoji text);"
	recipientInsert    = "INSERT INTO recipients (id, e164, uuid, username, email, is_blocked, profile_key, profile_key_credential, profile_given_name, profile_family_name, profile_joined_name, signal_profile_avatar, profile_sharing_enabled, last_profile_fetch, unidentified_access_mode, storage_service_id, storage_proto, capabilities, last_session_reset, about, about_emoji) VALUES (:id, :e164, :uuid, :username, :email, :is_blocked, :profile_key, :profile_key_credential, :profile_given_name, :profile_family_name, :profile_joined_name, :signal_profile_avatar, :profile_sharing_enabled, :last_profile_fetch, :unidentified_access_mode, :storage_service_id, :storage_proto, :capabilities, :last_session_reset, :about, :about_emoji)"
	recipientUpdate    = "REPLACE INTO recipients (id, e164, uuid, username, email, is_blocked, profile_key, profile_key_credential, profile_given_name, profile_family_name, profile_joined_name, signal_profile_avatar, profile_sharing_enabled, last_profile_fetch, unidentified_access_mode, storage_service_id, storage_proto, capabilities, last_session_reset, about, about_emoji) VALUES (:id, :e164, :uuid, :username, :email, :is_blocked, :profile_key, :profile_key_credential, :profile_given_name, :profile_family_name, :profile_joined_name, :signal_profile_avatar, :profile_sharing_enabled, :last_profile_fetch, :unidentified_access_mode, :storage_service_id, :storage_proto, :capabilities, :last_session_reset, :about, :about_emoji)"
	recipientGetById   = "SELECT * FROM recipients WHERE id = ?"
	recipientGetByUUID = "SELECT * FROM recipients WHERE uuid = ? Limit 1"
	recipientGetByE164 = "SELECT * FROM recipients WHERE e164 = ?"
)

const TIME_FORMAT = "2006-01-02 15:04:05 -0700 MST"

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
	About                  string `db:"about"`
	AboutEmoji             string `db:"about_emoji"`
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
		log.Debugln("[axolotl] CreateRecipient: Recipient already exists", recipient.UUID)
		return storedRecipeit, nil
	}
	// get last inserted recipient id
	var lastId int64
	err := DS.Dbx.Get(&lastId, "SELECT id FROM recipients ORDER BY id DESC LIMIT 1;")
	if err != nil {
		lastId = 0
	}
	recipient.Id = lastId + 1
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

// GetOrCreateRecipient returns a recipient by uuid or creates a new one
func (r *Recipients) GetOrCreateRecipient(uuid string) *Recipient {
	var err error
	recipient := r.GetRecipientByUUID(uuid)
	if recipient == nil {
		recipient = &Recipient{
			UUID: uuid,
		}
		recipient, err = r.CreateRecipient(recipient)
		if err != nil {
			log.Errorln("[axolotl] GetOrCreateRecipient Error creating recipient", err)
		}
	}
	return recipient
}

// GetOrCreateRecipientForContact returns a recipient for a contact
func (r *Recipients) GetOrCreateRecipientForContact(contact *contacts.Contact) *Recipient {
	var err error
	recipient := r.GetRecipientByUUID(contact.UUID)
	if recipient == nil {
		recipient = &Recipient{
			UUID:             contact.UUID,
			ProfileGivenName: contact.Name,
			E164:             contact.Tel,
		}
		recipient, err = r.CreateRecipient(recipient)
		if err != nil {
			log.Errorln("[axolotl] Error creating recipient", err)
		}
	}
	return recipient
}

// CreateRecipientWithoutProfileUpdate is only used in the v1.6.0 migration because we migrate before we initialize the signal server
func (r *Recipients) CreateRecipientWithoutProfileUpdate(recipient *Recipient) (*Recipient, error) {
	log.Debugln("[axolotl] Creating recipient without profile update", recipient.UUID)
	storedRecipient := r.GetRecipientByUUID(recipient.UUID)

	if storedRecipient != nil {
		log.Debug("[axolotl] Recipient already exists")
		recipient.Id = storedRecipient.Id
		err := recipient.SaveRecipient()
		if err != nil {
			return nil, err
		}
	} else {
		// get last inserted recipient id
		var lastId int64
		err := DS.Dbx.Get(&lastId, "SELECT id FROM recipients ORDER BY id DESC LIMIT 1;")
		if err != nil {
			lastId = 0
		}
		recipient.Id = lastId + 1
		_, err = DS.Dbx.NamedExec(recipientInsert, recipient)
		if err != nil {
			return nil, err
		}
	}

	return recipient, nil
}

// GetRecipientById returns a recipient by id
func (*Recipients) GetRecipientById(id int64) *Recipient {
	recipient := Recipient{}
	err := DS.Dbx.Get(&recipient, recipientGetById, id)
	if err != nil {
		return nil
	}
	return &recipient
}

// GetRecipientByE164 returns a recipient by e164
func (*Recipients) GetRecipientByE164(e164 string) *Recipient {
	recipient := Recipient{}
	err := DS.Dbx.Get(&recipient, recipientGetByE164, e164)
	if err != nil {
		return nil
	}
	return &recipient
}

// GetRecipientByUUID returns a recipient by uuid
func (*Recipients) GetRecipientByUUID(uuid string) *Recipient {
	recipient := Recipient{}
	err := DS.Dbx.Get(&recipient, recipientGetByUUID, uuid)
	if err != nil {
		return nil
	}
	return &recipient
}

// SaveRecipient saves a recipient
func (r *Recipient) SaveRecipient() error {
	_, err := DS.Dbx.NamedExec(recipientUpdate, r)
	return err
}

// UpdateProfile updates a recipient's profile
func (r *Recipient) UpdateProfile() error {
	log.Debugln("[axolotl] Updating profile for recipient", r.UUID)
	// update profile only once an hour max
	lastFetch, _ := time.Parse(TIME_FORMAT, r.LastProfileFetch)
	if lastFetch.Unix() < time.Now().Unix()-86400 {
		return nil
	}
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
			if profile != nil {
				if len(profile.Name) > 0 {
					log.Debug("[axolotl] New profile name:", profile.Name)
					r.Username = profile.Name
				}
				if len(profile.About) > 0 {
					r.About = profile.About
				}
				if len(profile.AboutEmoji) > 0 {
					r.AboutEmoji = profile.AboutEmoji
				}
			}
		}
	} else if r.Username == "" {
		profile, err = textsecure.GetProfile(r.UUID, r.ProfileKey)
		if err != nil {
			return err
		}
		if profile != nil {
			if len(profile.Name) > 0 {
				log.Debug("[axolotl] New profile name:", profile.Name)
				r.Username = profile.Name
			}
			if len(profile.About) > 0 {
				r.About = profile.About
			}
			if len(profile.AboutEmoji) > 0 {
				r.AboutEmoji = profile.AboutEmoji
			}
		}
	}
	r.LastProfileFetch = time.Now().Format(TIME_FORMAT)
	r.SaveRecipient()
	return nil
}
