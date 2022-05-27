package store

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/nanu-c/axolotl/app/settings"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

func getSalt(path string) ([]byte, error) {
	salt := make([]byte, 8)

	if _, err := os.Stat(path); err == nil {
		salt, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(path, salt, 0600)
		if err != nil {
			return nil, err
		}
	}

	return salt, nil
}

// Get raw key data for use with sqlcipher
func getKey(saltPath, password string) ([]byte, error) {
	log.Debugf("[axolotl] get decryption key")

	salt, err := getSalt(saltPath)
	if err != nil {
		log.Errorf("Failed to get salt")
		return nil, err
	}

	return scrypt.Key([]byte(password), salt, 16384, 8, 1, 32)
}
func (ds *DataStore) Encrypt(dbFile string, password string) error {
	log.Debugf("Encrypt Database")
	err := DS.Dbx.Ping()
	if err != nil {
		log.Errorf(err.Error())
	}
	key, err := getKey(filepath.Join(filepath.Dir(dbFile), "salt"), password)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to get key")
		return err
	}

	query := fmt.Sprintf("ATTACH DATABASE '%s' AS encrypted KEY \"x'%X'\"", dbFile, key)
	log.Debugf("Encrypt db file: " + dbFile)
	_, err = DS.Dbx.Exec(query)
	if err != nil {
		log.Errorf("firstError")
		return err
	}

	_, err = DS.Dbx.Exec("PRAGMA encrypted.cipher_page_size = 4096;")
	if err != nil {
		return err
	}

	_, err = DS.Dbx.Exec("SELECT sqlcipher_export('encrypted');")
	if err != nil {
		return err
	}

	_, err = DS.Dbx.Exec("DETACH DATABASE encrypted;")
	if err != nil {
		return err
	}
	settings.SettingsModel.EncryptDatabase = true
	DS.Dbx = nil

	return nil
}
func (ds *DataStore) Convert(password string) error {
	log.Debugf("Convert Data Storage")
	if password == "" {
		return fmt.Errorf("No password given")
	}
	dbDir = GetDbDir()
	dbFile = GetDbFile()
	tmp := GetDbTmpFile()

	//create tmp file
	_, err := os.OpenFile(tmp, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	saltFile = filepath.Join(dbDir, "salt")

	encrypted := settings.SettingsModel.EncryptDatabase
	log.Debugf("Encrypt: " + strconv.FormatBool(encrypted))

	if !encrypted {
		log.Debugf("Convert: Encrypting database..")

		ds, err := NewStorage("")
		if err != nil {
			return err
		}

		err = ds.Encrypt(tmp, password)
		if err != nil {
			log.Errorf("encrypting db: " + err.Error())

			return err
		}
	} else {
		log.Info("Convert: Decrypting database..")

		ds, err := NewStorage(password)
		if err != nil {
			return err
		}

		err = ds.Decrypt(tmp)
		if err != nil {
			return err
		}
	}

	err = os.Rename(tmp, dbFile)
	if err != nil {
		return err
	}

	settings.SettingsModel.EncryptDatabase = !encrypted
	// settings.Sync()
	return nil
}

//https://github.com/mutecomm/go-sqlcipher/blob/master/sqlcipher.go#L15
var sqlite3Header = []byte("SQLite format 3\000")

// IsEncrypted returns true, if the database with the given filename is
// encrypted, and false otherwise.
// If the database header cannot be read properly an error is returned.
func IsEncrypted(filename string) (bool, error) {

	// open file
	db, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer db.Close()
	// read header
	var header [16]byte
	n, err := db.Read(header[:])
	if err != nil {
		return false, err
	}
	log.Debugf("Headercompare: " + string(header[:]))
	if n != len(header) {
		return false, errors.New("go-sqlcipher: could not read full header")
	}
	// SQLCipher encrypts also the header, the file is encrypted if the read
	// header does not equal the header string used by SQLite 3.
	encrypted := !bytes.Equal(header[:], sqlite3Header)
	return encrypted, nil
}
