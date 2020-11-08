package helpers

import (
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/signal-golang/textsecure"
)

// Model for existing chat sessions
var (
	MsgFlagGroupNew              = 1
	MsgFlagGroupUpdate           = 2
	MsgFlagGroupLeave            = 4
	MsgFlagResetSession          = 8
	MsgFlagSetTimer              = 9
	MsgFlagSticker               = 10
	MsgFlagContact               = 11
	MsgFlagExpirationTimerUpdate = 12
	MsgFlagReaction              = 13
	MsgFlagQuote                 = 14
	MsgFlagHiddenQuote           = 15
)

func HumanizeTimestamp(ts uint64) string {
	nowms := uint64(time.Now().UnixNano() / 1000000)
	if ts > nowms {
		ts = nowms
	}
	return humanize.Time(time.Unix(0, int64(1000000*ts)))
}
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// membersDiffAndUnion returns a set diff and union of two contact sets represented as
// comma separated strings.
func MembersDiffAndUnion(aa, bb string) ([]string, string) {

	if bb == "" {
		return nil, aa
	}

	as := strings.Split(aa, ",")
	bs := strings.Split(bb, ",")

	rs := []string{}

	for _, b := range bs {
		found := false
		for _, a := range as {
			if a == b {
				found = true
				break
			}
		}
		if !found {
			rs = append(rs, b)
		}
	}
	return rs, strings.Join(append(as, rs...), ",")
}

const (
	ContentTypeMessage int = iota
	ContentTypeDocuments
	ContentTypePictures
	ContentTypeMusic
	ContentTypeContacts
	ContentTypeVideos
	ContentTypeLinks
)

func MimeTypeToContentType(mt string) int {
	ct := ContentTypeMessage
	if strings.HasPrefix(mt, "image") {
		ct = ContentTypePictures
	}
	if strings.HasPrefix(mt, "video") {
		ct = ContentTypeVideos
	}
	if strings.HasPrefix(mt, "audio") {
		ct = ContentTypeMusic
	}
	return ct
}

func ContentType(att io.Reader, mt string) int {
	if att == nil {
		return ContentTypeMessage
	}
	if mt == "" {
		mt, _ = textsecure.MIMETypeFromReader(att)
	}
	return MimeTypeToContentType(mt)
}

func RandomString(length int) string {
	//Lowercase and Uppercase Both
	charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	var output strings.Builder
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		output.WriteString(string(randomChar))
	}
	return output.String()
}
