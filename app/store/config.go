package store

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/morph027/textsecure"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/lang"
)

var AppName = "textsecure.nanuc"

var AppVersion = "0.3.16"

// Do not allow sending attachments larger than 100M for now
var MaxAttachmentSize int64 = 100 * 1024 * 1024

var (
	IsPhone      bool
	IsPushHelper bool
	MainQml      string

	HomeDir      string
	ConfigDir    string
	CacheDir     string
	ConfigFile   string
	ContactsFile string
	SettingsFile string
	LogFile      string
	DataDir      string
	StorageDir   string
	attachDir    string
	tsDeviceURL  string
	VcardPath    string
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
	lang.SetupTranslations(AppName)

	IsPhone = helpers.Exists("/home/phablet")
	IsPushHelper = filepath.Base(os.Args[0]) == "pushHelper"

	flag.Parse()
	if len(flag.Args()) == 1 {
		tsDeviceURL = flag.Arg(0)
	}

	if IsPushHelper {
		log.Printf("isPushhelper")
		HomeDir = "/home/phablet"
	} else {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		HomeDir = user.HomeDir
	}
	CacheDir = filepath.Join(HomeDir, ".cache/", AppName)
	LogFileName := []string{"application-click-", AppName, "_textsecure_", AppVersion, ".log"}
	LogFile = filepath.Join(HomeDir, ".cache/", "upstart/", strings.Join(LogFileName, ""))
	log.Printf("LogFile: " + LogFile)
	ConfigDir = filepath.Join(HomeDir, ".config/", AppName)
	ContactsFile = filepath.Join(ConfigDir, "contacts.yml")
	SettingsFile = filepath.Join(ConfigDir, "settings.yml")
	if _, err := os.Stat(SettingsFile); os.IsNotExist(err) {
		os.OpenFile(SettingsFile, os.O_RDONLY|os.O_CREATE, 0700)
	}
	os.MkdirAll(ConfigDir, 0700)
	DataDir = filepath.Join(HomeDir, ".local", "share", AppName)
	attachDir = filepath.Join(DataDir, "attachments")
	os.MkdirAll(attachDir, 0700)
	StorageDir = filepath.Join(DataDir, ".storage")
	if err := SetupDB(); err != nil {
		log.Fatal(err)
	}
	RefreshContacts()
}
