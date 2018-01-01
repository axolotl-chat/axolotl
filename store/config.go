package store

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/janimo/textsecure"
	"github.com/nanu-c/textsecure-qml/lang"
	"github.com/nanu-c/textsecure-qml/models"
)

var AppName = "textsecure.nanuc"

var AppVersion = "0.3.14"

// Do not allow sending attachments larger than 100M for now
var MaxAttachmentSize int64 = 100 * 1024 * 1024

var (
	IsPhone      bool
	IsPushHelper bool
	MainQml      string

	homeDir      string
	ConfigDir    string
	cacheDir     string
	ConfigFile   string
	ContactsFile string
	SettingsFile string
	LogFile      string
	dataDir      string
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
		if !models.Exists(ConfigFile) {
			cf = filepath.Join(ConfigDir, "config.yml")
		}
	}
	var err error
	if models.Exists(cf) {
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
	if models.Exists(rootCA) {
		Config.RootCA = rootCA
	}
	return Config, err
}
func SetupConfig() {
	lang.SetupTranslations(AppName)

	IsPhone = models.Exists("/home/phablet")
	IsPushHelper = filepath.Base(os.Args[0]) == "pushHelper"

	flag.Parse()
	if len(flag.Args()) == 1 {
		tsDeviceURL = flag.Arg(0)
	}

	if IsPushHelper {
		homeDir = "/home/phablet"
	} else {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		homeDir = user.HomeDir
	}
	cacheDir = filepath.Join(homeDir, ".cache/", AppName)
	LogFileName := []string{"application-click-", AppName, "_textsecure_", AppVersion, ".log"}
	LogFile = filepath.Join(homeDir, ".cache/", "upstart/", strings.Join(LogFileName, ""))
	log.Printf("LogFile: " + LogFile)
	ConfigDir = filepath.Join(homeDir, ".config/", AppName)
	ContactsFile = filepath.Join(ConfigDir, "contacts.yml")
	SettingsFile = filepath.Join(ConfigDir, "settings.yml")
	if _, err := os.Stat(SettingsFile); os.IsNotExist(err) {
		os.OpenFile(SettingsFile, os.O_RDONLY|os.O_CREATE, 0700)
	}
	os.MkdirAll(ConfigDir, 0700)
	dataDir = filepath.Join(homeDir, ".local", "share", AppName)
	attachDir = filepath.Join(dataDir, "attachments")
	os.MkdirAll(attachDir, 0700)
	StorageDir = filepath.Join(dataDir, ".storage")
	if err := SetupDB(); err != nil {
		log.Fatal(err)
	}
}
