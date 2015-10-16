package main

import (
	"crypto/rand"
	"encoding/hex"
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

	"github.com/dustin/go-humanize"
	"github.com/janimo/textsecure"
	"gopkg.in/qml.v1"
)

var appName = "textsecure.jani"

var appVersion = "0.3.1"

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
	attachDir    string
)

func init() {
	flag.BoolVar(&phone, "phone", false, "Indicate the app runs on the Ubuntu phone")
	flag.StringVar(&mainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
}

func saveAttachment(r io.Reader) (string, error) {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		return "", err
	}

	fn := filepath.Join(attachDir, hex.EncodeToString(id))
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return "", err

	}

	return fn, nil
}

func groupUpdateMsg(tels []string, title string) string {
	s := ""
	for _, t := range tels {
		s += telToName(t) + ", "
	}
	return s[:len(s)-2] + " joined the group. Title is now '" + title + "'."
}

func messageHandler(msg *textsecure.Message) {
	var r io.Reader
	var err error

	f := ""
	if len(msg.Attachments()) > 0 {
		r = msg.Attachments()[0]
		f, err = saveAttachment(r)
		if err != nil {
			log.Printf("Error saving %s\n", err.Error())
		}
	}

	text := msg.Message()
	if msg.Flags() == textsecure.EndSessionFlag {
		text = sessionReset
	}

	gr := msg.Group()

	if gr != nil && gr.Flags == textsecure.GroupLeaveFlag {
		text = telToName(msg.Source()) + " has left the group."
	}
	if gr != nil && gr.Flags == textsecure.GroupUpdateFlag {
		text = groupUpdateMsg(gr.Members, gr.Name)
	}

	if gr != nil && gr.Flags != 0 {
		_, ok := groups[gr.Hexid]
		groups[gr.Hexid] = &GroupRecord{
			GroupID: gr.Hexid,
			Members: strings.Join(gr.Members, ","),
			Name:    gr.Name,
		}
		if ok {
			updateGroup(groups[gr.Hexid])
		} else {
			saveGroup(groups[gr.Hexid])
		}
	}

	s := msg.Source()
	if gr != nil {
		s = gr.Hexid
	}
	session := sessionsModel.Get(s)
	m := session.Add(text, msg.Source(), f, false)
	m.ReceivedAt = uint64(time.Now().UnixNano() / 1000000)
	m.SentAt = msg.Timestamp()
	m.HTime = humanizeTimestamp(m.SentAt)
	qml.Changed(m, &m.HTime)
	session.Timestamp = m.SentAt
	session.When = m.HTime
	qml.Changed(session, &session.When)
	if gr != nil && gr.Flags == textsecure.GroupUpdateFlag {
		session.Name = gr.Name
		qml.Changed(session, &session.Name)
	}
	saveMessage(m)
	updateSession(session)
}

func receiptHandler(source string, devID uint32, timestamp uint64) {
	s := sessionsModel.Get(source)
	for i := len(s.messages) - 1; i >= 0; i-- {
		m := s.messages[i]
		if m.SentAt == timestamp {
			m.IsRead = true
			qml.Changed(m, &m.IsRead)
			updateMessageRead(m)
			return
		}
	}
	log.Printf("Message with timestamp %d not found\n", timestamp)
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
	attachDir = filepath.Join(dataDir, "attachments")
	os.MkdirAll(attachDir, 0700)
	storageDir = filepath.Join(dataDir, ".storage")
	if err := setupDB(); err != nil {
		showError(err)
	}
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

func sendMessage(to, message string, group bool, att io.Reader, end bool) uint64 {
	var err error
	var ts uint64
	for {
		err = nil
		if end {
			ts, err = textsecure.EndSession(to, "TERMINATE")
		} else if att == nil {
			if group {
				ts, err = textsecure.SendGroupMessage(to, message)
			} else {
				ts, err = textsecure.SendMessage(to, message)
			}
		} else {
			if group {
				ts, err = textsecure.SendGroupAttachment(to, message, att)
			} else {
				ts, err = textsecure.SendAttachment(to, message, att)
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

func humanizeTimestamp(ts uint64) string {
	return humanize.Time(time.Unix(0, int64(1000000*ts)))
}

func (api *textsecureAPI) SendMessage(to, message string) error {
	session := sessionsModel.Get(to)
	m := session.Add(message, "", "", true)
	saveMessage(m)
	go func() {
		ts := sendMessage(to, message, session.IsGroup, nil, false)
		m.SentAt = ts
		session.Timestamp = m.SentAt
		m.IsSent = true
		qml.Changed(m, &m.IsSent)
		m.HTime = humanizeTimestamp(m.SentAt)
		qml.Changed(m, &m.HTime)
		session.When = m.HTime
		qml.Changed(session, &session.When)
		updateMessageSent(m)
		updateSession(session)
	}()
	return nil
}

func (api *textsecureAPI) SendContactAttachment(to, message string, file string) error {
	phone, err := phoneFromVCardFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	return api.SendMessage(to, phone)
}

// Do not allow sending attachments larger than 100M for now
var maxAttachmentSize int64 = 100 * 1024 * 1024

func (api *textsecureAPI) SendAttachment(to, message string, file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if fi.Size() > maxAttachmentSize {
		showError(errors.New("Attachment too large, not sending"))
		return nil
	}

	session := sessionsModel.Get(to)
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	m := session.Add(message, "", file, true)
	saveMessage(m)
	r, err = os.Open(file)
	if err != nil {
		return err
	}
	go func() {
		ts := sendMessage(to, message, session.IsGroup, r, false)
		m.IsSent = true
		m.SentAt = ts
		qml.Changed(m, &m.IsSent)
		updateMessageSent(m)
		updateSession(session)
	}()
	return nil
}

var sessionReset = "Secure session reset."

func (api *textsecureAPI) EndSession(tel string) error {
	session := sessionsModel.Get(tel)
	m := session.Add(sessionReset, "", "", true)
	saveMessage(m)
	go func() {
		ts := sendMessage(tel, "", false, nil, true)
		m.IsSent = true
		m.SentAt = ts
		session.Timestamp = m.SentAt
		qml.Changed(m, &m.IsSent)
		updateMessageSent(m)
		updateSession(session)
	}()
	return nil
}

var vcardPath string

func (api *textsecureAPI) ContactsImported(path string) {
	vcardPath = path
	refreshContacts()
}

var groups = map[string]*GroupRecord{}

// FIXME: receive members as splice, blocked by https://github.com/go-qml/qml/issues/137
func (api *textsecureAPI) NewGroup(name string, members string) error {
	m := strings.Split(members, ",")
	group, err := textsecure.NewGroup(name, m)
	if err != nil {
		showError(err)
		return err
	}

	members = members + "," + config.Tel
	groups[group.Hexid] = &GroupRecord{
		GroupID: group.Hexid,
		Name:    name,
		Members: members,
	}
	saveGroup(groups[group.Hexid])
	session := sessionsModel.Get(group.Hexid)
	msg := session.Add(groupUpdateMsg(append(m, config.Tel), name), "", "", true)
	saveMessage(msg)

	return nil

}

// membersUnion perfoms a set union of two contact sets represented as
// comma separated strings.
func membersUnion(aa, bb string) string {

	if bb == "" {
		return aa
	}

	as := strings.Split(aa, ",")
	bs := strings.Split(bb, ",")

	rs := as
	for _, b := range bs {
		found := false
		for _, a := range as {
			if a == b {
				found = true
				break
			}
		}
		if !found {
			rs = append(rs, b)
		}
	}
	return strings.Join(rs, ",")
}

func (api *textsecureAPI) UpdateGroup(hexid, name string, members string) error {
	g, ok := groups[hexid]
	if !ok {
		return fmt.Errorf("Unknown group id %s\n", hexid)
	}
	members = membersUnion(g.Members, members)
	m := strings.Split(members, ",")
	group, err := textsecure.UpdateGroup(hexid, name, m)
	if err != nil {
		showError(err)
		return err
	}

	groups[group.Hexid] = &GroupRecord{
		GroupID: group.Hexid,
		Name:    name,
		Members: members,
	}
	updateGroup(groups[group.Hexid])
	session := sessionsModel.Get(group.Hexid)
	msg := session.Add(groupUpdateMsg(m, name), "", "", true)
	saveMessage(msg)
	session.Name = name
	qml.Changed(session, &session.Name)
	updateSession(session)
	return nil
}

func (api *textsecureAPI) LeaveGroup(hexid string) error {
	err := textsecure.LeaveGroup(hexid)
	if err != nil {
		return err
	}
	return deleteGroup(hexid)
}

func (api *textsecureAPI) GroupInfo(hexid string) string {
	s := ""
	if g, ok := groups[hexid]; ok {
		for _, t := range strings.Split(g.Members, ",") {
			s += telToName(t) + "\n\n"
		}
	}
	return s
}

func avatarImageProvider(id string, width, height int) image.Image {
	c := getContactForTel(id)
	if c == nil || c.Photo == "" {
		return image.NewAlpha(image.Rectangle{})
	}
	r := strings.NewReader(c.Photo)
	img, _, err := image.Decode(r)
	if err != nil {
		return image.NewAlpha(image.Rectangle{})

	}
	return img
}

func runUI() error {
	engine = qml.NewEngine()
	engine.AddImageProvider("avatar", avatarImageProvider)
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
