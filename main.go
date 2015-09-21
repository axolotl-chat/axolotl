package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/janimo/textsecure"
	"gopkg.in/qml.v1"
)

var appName = "textsecure.jani"

var appVersion = "0.2.5"

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
	s := msg.Group()
	if s == "" {
		s = msg.Source()
	}
	session := sessionsModel.Get(s)
	var r io.Reader
	if len(msg.Attachments()) > 0 {
		r = msg.Attachments()[0]
	}
	session.Add(msg.Message(), msg.Source(), r, false)
}

func receiptHandler(source string, devID uint32, timestamp uint64) {
	s := sessionsModel.Get(source)
	for i := len(s.messages) - 1; i >= 0; i-- {
		m := s.messages[i]
		if m.Timestamp == timestamp {
			m.IsRead = true
			qml.Changed(m, &m.IsRead)
			return
		}
	}
	log.Println("Message with timestamp %d not found\n", timestamp)
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
	config.UnencryptedStorage = true
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
		ReceiptHandler:      receiptHandler,
		RegistrationDone:    registrationDone,
	}

	if phone {
		client.GetLocalContacts = getAddressBookContactsFromContentHub
	} else {
		client.GetLocalContacts = getDesktopContacts
	}

	err := textsecure.Setup(client)
	if _, ok := err.(*strconv.NumError); ok {
		showError(errors.New("Switching to unencrypted session store for now.\n On the phone rm -Rf /home/phablet/.local/share/textsecure.jani/.storage/\n This will reset your sessions and reregister your phone."))
		return
	}
	if err != nil {
		showError(err)
		return
	}

	if exists(contactsFile) {
		api.HasContacts = true
		refreshContacts()
	}

	for {
		if err := textsecure.StartListening(); err != nil {
			log.Println(err)
			time.Sleep(3 * time.Second)
		}
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

func sendMessage(to, message string, group bool, att io.Reader) uint64 {
	var err error
	var ts uint64
	for {
		err = nil
		if att == nil {
			if group {
				err = textsecure.SendGroupMessage(to, message)
			} else {
				err, ts = textsecure.SendMessage(to, message)
			}
		} else {
			if group {
				err = textsecure.SendGroupAttachment(to, message, att)
			} else {
				err, ts = textsecure.SendAttachment(to, message, att)
			}
		}
		if err == nil {
			break
		}
		log.Println(err)
		//If sending failed, try again after a while
		time.Sleep(3 * time.Second)
	}
	return ts
}

func (api *textsecureAPI) SendMessage(to, message string) error {
	session := sessionsModel.Get(to)
	m := session.Add(message, "", nil, true)
	go func() {
		ts := sendMessage(to, message, session.IsGroup, nil)
		m.IsSent = true
		m.Timestamp = ts
		qml.Changed(m, &m.IsSent)
	}()
	return nil
}

func (api *textsecureAPI) SendAttachment(to, message string, file string) error {
	session := sessionsModel.Get(to)
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()
	m := session.Add(message, "", r, true)
	r, err = os.Open(file)
	if err != nil {
		return err
	}
	go func() {
		ts := sendMessage(to, message, session.IsGroup, r)
		m.IsSent = true
		m.Timestamp = ts
		qml.Changed(m, &m.IsSent)
	}()
	return nil
}

var vcardPath string

func (api *textsecureAPI) ContactsImported(path string) {
	vcardPath = path
	refreshContacts()
}

// FIXME: receive members as splice, blocked by https://github.com/go-qml/qml/issues/137
func (api *textsecureAPI) NewGroup(name string, members string) error {
	m := strings.Split(members, ":")
	err := textsecure.NewGroup(name, m)
	if err != nil {
		return err
	}
	session := sessionsModel.Get(name)
	session.Add("Group created with "+members, "", nil, true)
	return nil
}

func runUI() error {
	engine = qml.NewEngine()

	engine.AddImageProvider("ts", func(id string, width, height int) image.Image {
		s := strings.Split(id, ":")
		tel := s[0]
		i, _ := strconv.Atoi(s[1])
		ses := sessionsModel.Get(tel)
		att := ses.messages[i].Att
		if att == nil {
			return image.NewAlpha(image.Rectangle{})
		}
		r := bytes.NewBuffer(att)
		img, _, err := image.Decode(r)
		if err != nil {
			return image.NewAlpha(image.Rectangle{})

		}
		return img
	})

	initModels()
	engine.Context().SetVar("textsecure", api)
	engine.Context().SetVar("appVersion", appVersion)

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
