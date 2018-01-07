package store

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/morph027/textsecure"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/snapcore/snapd/osutil"
)

func SaveAttachment(a *textsecure.Attachment) (string, error) {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		return "", err
	}

	ext := ""
	if strings.HasPrefix(a.MimeType, "video/") {
		ext = strings.Replace(a.MimeType, "video/", ".", 1)
	}

	fn := filepath.Join(config.AttachDir, hex.EncodeToString(id)+ext)
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, a.R)
	if err != nil {
		return "", err

	}

	return fn, nil
}

// copyAttachment makes a copy of a file that is in the volatile content hub cache
func CopyAttachment(src string) (string, error) {
	_, b := filepath.Split(src)
	dest := filepath.Join(config.AttachDir, b)
	err := osutil.CopyFile(src, dest, osutil.CopyFlagOverwrite)
	if err != nil {
		return "", err
	}
	return dest, nil
}
