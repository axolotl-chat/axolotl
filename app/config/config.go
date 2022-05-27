package config

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/signal-golang/textsecure"
	textsecureConfig "github.com/signal-golang/textsecure/config"
)

const AppName = "textsecure.nanuc"

const AppVersion = "1.2.0"

const LogFileName = "application-click-" + AppName + "_textsecure_" + AppVersion + ".log"

// Do not allow sending attachments larger than 100M for now
const MaxAttachmentSize int64 = 100 * 1024 * 1024

type Config struct {
	TsConfig               *textsecureConfig.Config
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
}

func (c *Config) GetConfig() (*textsecureConfig.Config, error) {
	c.ConfigFile = filepath.Join(c.ConfigDir, "config.yml")
	cf := c.ConfigFile
	if c.IsPhone {
		c.ConfigDir = filepath.Join("/home/phablet/.config/textsecure.nanuc/")
		if !helpers.Exists(c.ConfigFile) {
			cf = filepath.Join(c.ConfigDir, "config.yml")
		}
	}
	if _, err := os.Stat(c.ConfigFile); os.IsNotExist(err) {
		log.Debugln("[axolotl] create config file")
		_, err := os.OpenFile(c.ConfigFile, os.O_RDONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Errorln("[axolotl] creating config file", err.Error())
		}
	}
	var err error
	if helpers.Exists(cf) {
		c.TsConfig, err = textsecure.ReadConfig(cf)
	} else {
		c.TsConfig = &textsecureConfig.Config{}
	}
	c.TsConfig.StorageDir = c.StorageDir
	log.Debugln("[axolotl] config path: ", c.ConfigDir)
	c.TsConfig.UserAgent = fmt.Sprintf("TextSecure %s for Ubuntu Phone", AppVersion)
	c.TsConfig.UnencryptedStorage = true

	if c.TsConfig.LogLevel == "" {
		c.TsConfig.LogLevel = "info"
	}
	c.TsConfig.CrayfishSupport = true
	c.TsConfig.AlwaysTrustPeerID = true
	rootCA := filepath.Join(c.ConfigDir, "rootCA.crt")
	if helpers.Exists(rootCA) {
		c.TsConfig.RootCA = rootCA
	}
	return c.TsConfig, err
}
func SetupConfig() *Config {
	c := &Config{}

	log.Debugln("[axolotl] setup config")

	c.IsPhone = GetIsPhone()
	c.IsPushHelper = GetIsPushHelper()

	flag.StringVar(&c.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
	flag.StringVar(&c.Gui, "e", "", "Specify runtime environment. Use either electron, ut, lorca, qt or server")
	flag.StringVar(&c.AxolotlWebDir, "axolotlWebDir", "./axolotl-web/dist", "Specify the directory to use for axolotl-web")
	flag.BoolVar(&c.ElectronDebug, "eDebug", false, "Open electron development console")
	flag.BoolVar(&c.PrintVersion, "version", false, "Print version info")
	flag.StringVar(&c.ServerHost, "host", "127.0.0.1", "Host to serve UI from.")
	flag.StringVar(&c.ServerPort, "port", "9080", "Port to serve UI from.")
	flag.StringVar(&c.ElectronFlag, "electron-flag", "", "Specify electron flag. Use no-ozone to disable Ozone/Wayland platform")

	flag.Parse()

	if len(flag.Args()) == 1 {
		c.TsDeviceURL = flag.Arg(0)
	}

	if c.PrintVersion {
		fmt.Printf("%s %s %s %s %s\n",
			AppName, AppVersion, runtime.GOOS, runtime.GOARCH, runtime.Version())
		os.Exit(0)
	}

	c.HomeDir = GetHomeDir()

	c.ConfigDir = GetConfigDir()
	c.ContactsFile = GetContactsFile()
	c.RegisteredContactsFile = GetRegisteredContactsFile()
	c.SettingsFile = GetSettingsFile()
	if _, err := os.Stat(c.SettingsFile); os.IsNotExist(err) {
		_, err := os.OpenFile(c.SettingsFile, os.O_RDONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Errorln("[axolotl] creating settings file", err.Error())
		}
	}
	os.MkdirAll(c.ConfigDir, 0700)
	c.DataDir = GetDataDir()
	c.LogFile = GetLogFile()
	c.AttachDir = GetAttachDir()
	os.MkdirAll(c.AttachDir, 0700)
	c.StorageDir = GetStorageDir()

	return c
}
func (c *Config) Unregister() {
	err := os.Remove(c.HomeDir + "/.local/share/textsecure.nanuc/db/db.sql")
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(c.ContactsFile)
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(c.SettingsFile)
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(c.ConfigFile)
	if err != nil {
		log.Error(err)
	}
	err = os.RemoveAll(c.HomeDir + "/.cache/textsecure.nanuc/qmlcache")
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(c.HomeDir + "/.config/textsecure.nanuc/config.yml")
	if err != nil {
		log.Error(err)
	}
	err = os.RemoveAll(c.StorageDir)
	if err != nil {
		log.Error(err)
	}
	err = os.RemoveAll(c.DataDir + AppName)
	if err != nil {
		log.Error(err)
	}
	os.Exit(1)
}
func (c *Config) SetLogLevel(loglevel string) {
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
	c.TsConfig.LogLevel = loglevel
	textsecure.WriteConfig(c.ConfigFile, c.TsConfig)
	textsecure.RefreshConfig()
}

func GetIsPhone() bool {
	return helpers.Exists("/home/phablet")
}

func GetIsPushHelper() bool {
	return filepath.Base(os.Args[0]) == "pushHelper"
}

func GetHomeDir() string {
	homeDir := ""
	if GetIsPushHelper() || GetIsPhone() {
		log.Printf("[axolotl] use push helper")
		homeDir = "/home/phablet"
	} else {
		user, err := user.Current()
		if err != nil {
			// log.Fatal(err)
			homeDir = "/home/phablet"
		} else {
			//if in a snap environment
			snapPath := os.Getenv("SNAP_USER_DATA")
			if len(snapPath) > 0 {
				homeDir = snapPath
			} else {
				homeDir = user.HomeDir
			}
		}
	}

	return homeDir
}

func GetConfigDir() string {
	homeDir := GetHomeDir()
	return filepath.Join(homeDir, ".config/", AppName)
}

func GetContactsFile() string {
	configDir := GetConfigDir()
	return filepath.Join(configDir, "contacts.yml")
}

func GetRegisteredContactsFile() string {
	configDir := GetConfigDir()
	return filepath.Join(configDir, "registeredContacts.yml")
}

func GetSettingsFile() string {
	configDir := GetConfigDir()
	return filepath.Join(configDir, "settings.yml")
}

func GetDataDir() string {
	homeDir := GetHomeDir()
	return filepath.Join(homeDir, ".local", "share", AppName)
}

func GetLogFile() string {
	logFile := ""
	if GetIsPushHelper() || GetIsPhone() {
		logFile = filepath.Join(GetHomeDir(), ".cache/", "upstart/", LogFileName)
	} else {
		logFile = filepath.Join(GetDataDir(), LogFileName)
	}
	return logFile
}

func GetAttachDir() string {
	dataDir := GetDataDir()
	return filepath.Join(dataDir, "attachments")
}

func GetStorageDir() string {
	dataDir := GetDataDir()
	return filepath.Join(dataDir, ".storage")
}
