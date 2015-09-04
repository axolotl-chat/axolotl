package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/janimo/textsecure"
	"gopkg.in/qml.v1"
)

var appName = "textsecure.jani"

var appVersion = "0.1.0"

var (
	phone   bool
	mainQml string
)

var (
	homeDir      string
	configDir    string
	configFile   string
	contactsFile string
	dataDir      string
	storageDir   string
)

func init() {
	flag.BoolVar(&phone, "phone", false, "Indicate the app runs on the Ubuntu phone")
	flag.StringVar(&mainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
}

func messageHandler(msg *textsecure.Message) {
	session := sessionsModel.Get(msg.Source())
	var r io.Reader
	if len(msg.Attachments()) > 0 {
		r = msg.Attachments()[0]
	}
	session.Add(msg.Message(), msg.Source(), r, false)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var config *textsecure.Config

func getConfig() (*textsecure.Config, error) {
	configFile = filepath.Join(configDir, "config.yml")
	cf := configFile
	if phone {
		configDir = filepath.Join("/opt/click.ubuntu.com", appName, "current")
		if !exists(configFile) {
			cf = filepath.Join(configDir, "config.yml")
		}
	}
	var err error
	if exists(cf) {
		config, err = textsecure.ReadConfig(cf)
	} else {
		config = &textsecure.Config{}
	}
	config.StorageDir = storageDir
	config.UserAgent = fmt.Sprintf("TextSecure %s for Ubuntu Phone", appVersion)
	rootCA := filepath.Join(configDir, "rootCA.crt")
	if exists(rootCA) {
		config.RootCA = rootCA
	}
	return config, err
}

func registrationDone() {
	log.Println("Registered")
	win.Root().Call("registered")
	textsecure.WriteConfig(configFile, config)
}

func showError(err error) {
	win.Root().Call("error", err.Error())
}

func setupEnvironment() {
	user, err := user.Current()
	if err != nil {
		showError(err)
	}
	homeDir = user.HomeDir
	configDir = filepath.Join(homeDir, ".config/", appName)
	contactsFile = filepath.Join(configDir, "contacts.yml")
	os.MkdirAll(configDir, 0700)
	dataDir = filepath.Join(homeDir, ".local", "share", appName)
	storageDir = filepath.Join(dataDir, ".storage")
}

func runBackend() {
	setupEnvironment()

	client := &textsecure.Client{
		GetConfig:           getConfig,
		GetPhoneNumber:      getPhoneNumber,
		GetVerificationCode: getVerificationCode,
		GetStoragePassword:  getStoragePassword,
		MessageHandler:      messageHandler,
		RegistrationDone:    registrationDone,
	}

	if phone {
		client.GetLocalContacts = getAddressBookContactsFromContentHub
	} else {
		client.GetLocalContacts = getDesktopContacts
	}

	err := textsecure.Setup(client)
	if err != nil {
		showError(err)
		return
	}

	if exists(contactsFile) {
		api.HasContacts = true
		refreshContacts()
	}

	if err := textsecure.StartListening(); err != nil {
		showError(err)
	}
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 1 {
		log.Println("URL", flag.Arg(0))
	}
	err := qml.Run(runUI)
	if err != nil {
		log.Fatal(err)
	}
}

var engine *qml.Engine
var win *qml.Window

type textsecureAPI struct {
	HasContacts bool
}

var api = &textsecureAPI{}

func (api *textsecureAPI) SendMessage(to, message string) error {
	session := sessionsModel.Get(to)
	session.Add(message, "", nil, true)
	return textsecure.SendMessage(to, message)
}

func (api *textsecureAPI) SendAttachment(to, message string, file string) error {
	session := sessionsModel.Get(to)
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()
	session.Add(message, "", r, true)
	r, err = os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()
	return textsecure.SendAttachment(to, message, r)
}

var vcardPath string

func (api *textsecureAPI) ContactsImported(path string) {
	vcardPath = path
	refreshContacts()
}

// FIXME: receive members as splice, blocked by https://github.com/go-qml/qml/issues/137
func (api *textsecureAPI) NewGroup(name string, members string) error {
	m := strings.Split(members, ":")
	return textsecure.NewGroup(name, m)
}

func runUI() error {
	engine = qml.NewEngine()

	initModels()
	engine.Context().SetVar("textsecure", api)

	component, err := engine.LoadFile(mainQml)
	if err != nil {
		return err
	}
	win = component.CreateWindow(nil)
	win.Show()

	go runBackend()
	win.Wait()
	return nil
}
