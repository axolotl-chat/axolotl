package config

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/signal-golang/textsecure"
	textsecureConfig "github.com/signal-golang/textsecure/config"
)

const AppName = "textsecure.nanuc"

const AppVersion = "1.5.0"

// Do not allow sending attachments larger than 100M for now
const MaxAttachmentSize int64 = 100 * 1024 * 1024

var (
	IsPhone                bool
	IsPushHelper           bool
	MainQml                string
	Gui                    string
	ElectronDebug          bool
	PrintVersion           bool
	HomeDir                string
	ConfigDir              string
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
	ServerHost             string
	ServerPort             string
	AxolotlWebDir          string
	ElectronFlag           string
)

var Config *textsecureConfig.Config

func GetConfig() (*textsecureConfig.Config, error) {
	ConfigFile = filepath.Join(ConfigDir, "config.yml")
	cf := ConfigFile
	if IsPhone {
		ConfigDir = filepath.Join("/home/phablet/.config/textsecure.nanuc/")
		if !helpers.Exists(ConfigFile) {
			cf = filepath.Join(ConfigDir, "config.yml")
		}
	}
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		log.Debugln("[axolotl] create config file")
		_, err := os.OpenFile(ConfigFile, os.O_RDONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Errorln("[axolotl] creating config file", err.Error())
		}
	}
	var err error
	if helpers.Exists(cf) {
		Config, err = textsecure.ReadConfig(cf)
	} else {
		Config = &textsecureConfig.Config{}
	}
	Config.StorageDir = StorageDir
	log.Debugln("[axolotl] config path: ", ConfigDir)
	Config.UserAgent = fmt.Sprintf("TextSecure %s for Ubuntu Phone", AppVersion)
	Config.UnencryptedStorage = true

	if Config.LogLevel == "" {
		Config.LogLevel = "info"
	}
	Config.CrayfishSupport = true
	Config.AlwaysTrustPeerID = true
	rootCA := filepath.Join(ConfigDir, "rootCA.crt")
	if helpers.Exists(rootCA) {
		Config.RootCA = rootCA
	}
	return Config, err
}
func SetupConfig() {
	log.Debugln("[axolotl] setup config")

	IsPhone = helpers.Exists("/home/phablet")
	IsPushHelper = filepath.Base(os.Args[0]) == "pushHelper"
	flag.Parse()
	if len(flag.Args()) == 1 {
		TsDeviceURL = flag.Arg(0)
	}

	if PrintVersion {
		fmt.Printf("%s %s %s %s %s\n",
			AppName, AppVersion, runtime.GOOS, runtime.GOARCH, runtime.Version())
		os.Exit(0)
	}

	if IsPushHelper || IsPhone {
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
	LogFileName := []string{"application-click-", AppName, "_textsecure_", AppVersion, ".log"}
	ConfigDir = filepath.Join(HomeDir, ".config/", AppName)
	ContactsFile = filepath.Join(ConfigDir, "contacts.yml")
	RegisteredContactsFile = filepath.Join(ConfigDir, "registeredContacts.yml")
	SettingsFile = filepath.Join(ConfigDir, "settings.yml")
	if _, err := os.Stat(SettingsFile); os.IsNotExist(err) {
		_, err := os.OpenFile(SettingsFile, os.O_RDONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Errorln("[axolotl] creating settings file", err.Error())
		}
	}
	os.MkdirAll(ConfigDir, 0700)
	DataDir = filepath.Join(HomeDir, ".local", "share", AppName)
	if IsPushHelper || IsPhone {
		LogFile = filepath.Join(HomeDir, ".cache/", "upstart/", strings.Join(LogFileName, ""))
	} else {
		LogFile = filepath.Join(DataDir, strings.Join(LogFileName, ""))

	}
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
	os.Exit(1)
}
func SetLogLevel(loglevel string) {
	if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "warn" {
		log.SetLevel(log.WarnLevel)
	} else if loglevel == "error" {
		log.SetLevel(log.ErrorLevel)
	} else if loglevel == "fatal" {
		log.SetLevel(log.FatalLevel)
	} else if loglevel == "panic" {
		log.SetLevel(log.PanicLevel)
	} else {
		log.SetLevel(log.InfoLevel)
		loglevel = "info"
	}
	Config.LogLevel = loglevel
	textsecure.WriteConfig(ConfigFile, Config)
	textsecure.RefreshConfig()
}
