package worker

import (
	"strings"

	"github.com/nanu-c/axolotl/app/store"
)

type GroupRecord struct {
	ID      int64
	GroupID string
	Name    string
	Members string
	Avatar  []byte
	Active  bool
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
