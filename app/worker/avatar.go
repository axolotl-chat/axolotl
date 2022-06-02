package worker

import (
	"github.com/nanu-c/axolotl/app/store"
)

func (a *TextsecureAPI) GetAvatarImage(id string) string {
	url := ""

	if c := store.GetContactForTel(id); c != nil {
		if c.HasAvatar {
			url = "image://avatar/" + id
		}
	}
	if g, ok := store.Groups[id]; ok {
		if len(g.Avatar) > 0 {
			url = "image://avatar/" + id
		}
	}
	return url
}
