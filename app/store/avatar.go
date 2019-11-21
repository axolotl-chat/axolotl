package store

import (
	"bytes"
	"image"
	"io"
)

func AvatarImageProvider(id string, width, height int) image.Image {
	var r io.Reader

	if c := GetContactForTel(id); c != nil {
		r = bytes.NewReader(c.Avatar)
	}

	if g, ok := Groups[id]; ok {
		r = bytes.NewReader(g.Avatar)
	}

	if r == nil {
		return image.NewAlpha(image.Rectangle{})
	}
	img, _, err := image.Decode(r)
	if err != nil {
		return image.NewAlpha(image.Rectangle{})

	}
	return img
}
