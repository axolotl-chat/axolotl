package config

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/signal-golang/textsecure"
)

var AppName = "textsecure.nanuc"

var AppVersion = "0.8.1"

// Do not allow sending attachments larger than 100M for now
var MaxAttachmentSize int64 = 100 * 1024 * 1024

var (
	IsPhone                bool
	IsPushHelper           bool
	MainQml                string
	Gui                    string
	ElectronDebug          bool
	HomeDir                string
	ConfigDir              string
	CacheDir               string
	ConfigFile             string
	ContactsFile           string
	RegisteredContactsFile string
	SettingsFile           string
	LogFile                string
	LogLevel               string
	DataDir                string
	StorageDir             string
	AttachDir              string
	TsDeviceURL            string
	VcardPath              string
)

var Config *textsecure.Config

func GetConfig() (*textsecure.Config, error) {
	ConfigFile = filepath.Join(ConfigDir, "config.yml")
	cf := ConfigFile
	if IsPhone {
		ConfigDir = filepath.Join("/opt/click.ubuntu.com", AppName, "current")
		if !helpers.Exists(ConfigFile) {
			cf = filepath.Join(ConfigDir, "config.yml")
		}
	}
	var err error
	if helpers.Exists(cf) {
		Config, err = textsecure.ReadConfig(cf)
	} else {
		Config = &textsecure.Config{}
	}
	Config.StorageDir = StorageDir
	log.Debugln("[axolotl] config path: ", ConfigDir)
	Config.UserAgent = fmt.Sprintf("TextSecure %s for Ubuntu Phone", AppVersion)
	Config.UnencryptedStorage = true

	Config.LogLevel = "debug"
	Config.AlwaysTrustPeerID = true
	rootCA := filepath.Join(ConfigDir, "rootCA.crt")
	if helpers.Exists(rootCA) {
		Config.RootCA = rootCA
	}
	return Config, err
}
func SetupConfig() {

	IsPhone = helpers.Exists("/home/phablet")
	IsPushHelper = filepath.Base(os.Args[0]) == "pushHelper"

	flag.Parse()
	if len(flag.Args()) == 1 {
		TsDeviceURL = flag.Arg(0)
	}

	if IsPushHelper {
		log.Printf("[axolotl] use push helper")
		HomeDir = "/home/phablet"
	} else {
		user, err := user.Current()
		if err != nil {
			// log.Fatal(err)
			HomeDir = "/home/phablet"
		} else {
			//if in a snap environment
			snapPath := os.Getenv("SNAP_USER_DATA")
			if len(snapPath) > 0 {
				HomeDir = snapPath
			} else {

				HomeDir = user.HomeDir
			}
		}
	}
	CacheDir = filepath.Join(HomeDir, ".cache/", AppName)
	LogFileName := []string{"application-click-", AppName, "_textsecure_", AppVersion, ".log"}
	LogFile = filepath.Join(HomeDir, ".cache/", "upstart/", strings.Join(LogFileName, ""))
	ConfigDir = filepath.Join(HomeDir, ".config/", AppName)
	ContactsFile = filepath.Join(ConfigDir, "contacts.yml")
	RegisteredContactsFile = filepath.Join(ConfigDir, "registeredContacts.yml")
	SettingsFile = filepath.Join(ConfigDir, "settings.yml")
	if _, err := os.Stat(SettingsFile); os.IsNotExist(err) {
		os.OpenFile(SettingsFile, os.O_RDONLY|os.O_CREATE, 0700)
	}
	os.MkdirAll(ConfigDir, 0700)
	DataDir = filepath.Join(HomeDir, ".local", "share", AppName)
	AttachDir = filepath.Join(DataDir, "attachments")
	os.MkdirAll(AttachDir, 0700)
	StorageDir = filepath.Join(DataDir, ".storage")

}
func Unregister() {
	err := os.Remove(HomeDir + "/.local/share/textsecure.nanuc/db/db.sql")
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(ContactsFile)
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(SettingsFile)
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(ConfigFile)
	if err != nil {
		log.Error(err)
	}
	err = os.RemoveAll(HomeDir + "/.cache/textsecure.nanuc/qmlcache")
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(HomeDir + "/.config/textsecure.nanuc/config.yml")
	if err != nil {
		log.Error(err)
	}
	err = os.RemoveAll(StorageDir)
	if err != nil {
		log.Error(err)
	}
	err = os.RemoveAll(DataDir + AppName)
	if err != nil {
		log.Error(err)
	}
	err = os.RemoveAll(CacheDir + AppName)
	if err != nil {
		log.Error(err)
	}
	os.Exit(1)
}
