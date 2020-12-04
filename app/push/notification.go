package push

/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 * Aaron Kimmig: aaron@nanu-c.org
 *
 * ciborium is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * nuntium is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"encoding/json"
	"sync"
	"time"

	dbus "github.com/z3ntu/go-dbus"
)

var (
	sessionBus        *dbus.Connection
	err               error
	Nh                *NotificationHandler
	useNotifications  bool
	mu                sync.Mutex
	lastNotifications map[string]int64
)

func NotificationInit() {
	if sessionBus, err = dbus.Connect(dbus.SessionBus); err != nil {
		//log.Fatal("Connection error: ", err)
		useNotifications = false
	} else {
		useNotifications = true
		Nh = NewLegacyHandler(sessionBus, "textsecure.nanuc_textsecure")
		lastNotifications = make(map[string]int64)
	}

}

// Notifications lives on a well-knwon bus.Address

const (
	dbusName        = "com.ubuntu.Postal"
	dbusInterface   = "com.ubuntu.Postal"
	dbusPathPart    = "/com/ubuntu/Postal/"
	dbusPostMethod  = "Post"
	dbusClearMethod = "ClearPersistent"
)

type VariantMap map[string]dbus.Variant

type NotificationHandler struct {
	dbusObject  *dbus.ObjectProxy
	application string
}

func NewLegacyHandler(conn *dbus.Connection, application string) *NotificationHandler {
	return &NotificationHandler{
		dbusObject:  conn.Object(dbusName, "/com/ubuntu/Postal/textsecure_2enanuc"),
		application: application,
	}
}
func (n *NotificationHandler) Clear(tag string) error {
	_, err := n.dbusObject.Call(dbusInterface, dbusClearMethod, "textsecure.nanuc_textsecure", tag)
	return err
}

func (n *NotificationHandler) Send(m *PushMessage) error {
	if useNotifications {
		var pushMessage string
		// check if there was a notification in the same chat less than 10s and
		//if yes just replace the notification but don't tripper a popup
		if timestamp, found := lastNotifications[m.Notification.Tag]; found {

			if m.Notification.Card.Timestamp-timestamp < 10 {
				m.Notification.Card.Popup = false
				m.Notification.RawSound = json.RawMessage(`false`)
				m.Notification.RawVibration = json.RawMessage(`false`)
			}
			lastNotifications[m.Notification.Tag] = m.Notification.Card.Timestamp
		} else {
			lastNotifications[m.Notification.Tag] = m.Notification.Card.Timestamp
		}
		if out, err := json.Marshal(m); err == nil {
			pushMessage = string(out)
		} else {
			return err
		}
		_, err := n.dbusObject.Call(dbusInterface, dbusClearMethod, "textsecure.nanuc_textsecure", m.Notification.Tag)
		_, err = n.dbusObject.Call(dbusInterface, dbusPostMethod, "textsecure.nanuc_textsecure", pushMessage)

		return err
	}
	return nil
}

// NewStandardPushMessage creates a base Notification with common
// components (members) setup.
func (n *NotificationHandler) NewStandardPushMessage(summary, body, icon string, tag string) *PushMessage {
	pm := &PushMessage{
		Message: summary,
		Notification: Notification{
			Card: &Card{
				Summary: summary,
				Body:    body,
				Actions: []string{"appid://textsecure.nanuc/textsecure/current-user-version"},
				// Icon:    icon,
				Popup:     true,
				Persist:   true,
				Timestamp: time.Now().Unix(),
			},
			Tag:          tag,
			RawSound:     json.RawMessage(`"sounds/ubuntu/notifications/Slick.ogg"`),
			RawVibration: json.RawMessage(`{"pattern": [100, 100], "repeat": 2}`),
		},
	}
	return pm
}

// PushMessage represents a data structure to be sent over to the
// Post Office. It consists of a Notification and a Message.
type PushMessage struct {
	// Notification (optional) describes the user-facing notifications
	// triggered by this push message.
	Notification Notification `json:"notification,omitempty"`
	Message      string       `json:"message,omitempty"`
}

// Notification (optional) describes the user-facing notifications
// triggered by this push message.
type Notification struct {
	Card         *Card           `json:"card,omitempty"`
	RawSound     json.RawMessage `json:"sound"`
	RawVibration json.RawMessage `json:"vibrate"`
	Tag          string          `json:"tag,omitempty"`
}

// Card is part of a notification and represents the user visible hints for
// a specific notification.
type Card struct {
	// Summary is a required title. The card will not be presented if this is missing.
	Summary string `json:"summary"`
	// Body is the longer text.
	Body string `json:"body,omitempty"`
	// Whether to show a bubble. Users can disable this, and can easily miss
	// them, so don’t rely on it exclusively.
	Popup bool `json:"popup,omitempty"`
	// Actions provides actions for the bubble's snap decissions.
	Actions []string `json:"actions,omitempty"`
	// Icon is a path to an icon to display with the notification bubble.
	// Icon string `json:"icon,omitempty"`
	// Whether to show in notification centre.
	Persist   bool  `json:"persist,omitempty"`
	Timestamp int64 `json:"timestamp"`
}
