package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gosexy/gettext"
	"github.com/janimo/textsecure"
	"github.com/ubuntu-core/snappy/helpers"
	"gopkg.in/qml.v1"
)

var appName = "textsecure.jani"

var appVersion = "0.3.8"

var (
	isPhone      bool
	isPushHelper bool
	mainQml      string

	homeDir      string
	configDir    string
	cacheDir     string
	configFile   string
	contactsFile string
	settingsFile string
	dataDir      string
	storageDir   string
	attachDir    string
	tsDeviceURL  string
)

func init() {
	flag.StringVar(&mainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
}

func saveAttachment(a *textsecure.Attachment) (string, error) {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		return "", err
	}

	ext := ""
	if strings.HasPrefix(a.MimeType, "video/") {
		ext = strings.Replace(a.MimeType, "video/", ".", 1)
	}

	fn := filepath.Join(attachDir, hex.EncodeToString(id)+ext)
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

func groupUpdateMsg(tels []string, title string) string {
	s := ""
	if len(tels) > 0 {
		for _, t := range tels {
			s += telToName(t) + ", "
		}
		s = s[:len(s)-2] + " joined the group. "
	}

	return s + "Title is now '" + title + "'."
}

func messageHandler(msg *textsecure.Message) {
	var err error

	f := ""
	mt := ""
	if len(msg.Attachments()) > 0 {
		mt = msg.Attachments()[0].MimeType
		f, err = saveAttachment(msg.Attachments()[0])
		if err != nil {
			log.Printf("Error saving %s\n", err.Error())
		}
	}

	msgFlags := 0

	text := msg.Message()
	if msg.Flags() == textsecure.EndSessionFlag {
		text = sessionReset
		msgFlags = msgFlagResetSession
	}

	gr := msg.Group()

	if gr != nil && gr.Flags != 0 {
		_, ok := groups[gr.Hexid]
		members := ""
		if ok {
			members = groups[gr.Hexid].Members
		}
		av := []byte{}

		if gr.Avatar != nil {
			av, err = ioutil.ReadAll(gr.Avatar)
			if err != nil {
				log.Println(err)
				return
			}
		}
		groups[gr.Hexid] = &GroupRecord{
			GroupID: gr.Hexid,
			Members: strings.Join(gr.Members, ","),
			Name:    gr.Name,
			Avatar:  av,
			Active:  true,
		}
		if ok {
			updateGroup(groups[gr.Hexid])
		} else {
			saveGroup(groups[gr.Hexid])
		}

		if gr.Flags == textsecure.GroupUpdateFlag {
			dm, _ := membersDiffAndUnion(members, strings.Join(gr.Members, ","))
			text = groupUpdateMsg(dm, gr.Name)
			msgFlags = msgFlagGroupUpdate
		}
		if gr.Flags == textsecure.GroupLeaveFlag {
			text = telToName(msg.Source()) + " has left the group."
			msgFlags = msgFlagGroupLeave
		}
	}

	s := msg.Source()
	if gr != nil {
		s = gr.Hexid
	}
	session := sessionsModel.Get(s)
	m := session.Add(text, msg.Source(), f, mt, false)
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

	if msgFlags != 0 {
		m.Flags = msgFlags
		qml.Changed(m, &m.Flags)
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
	if isPhone {
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
	config.LogLevel = "debug"
	config.AlwaysTrustPeerID = true
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

func setup() {

	setupTranslations()

	isPhone = exists("/home/phablet")
	isPushHelper = filepath.Base(os.Args[0]) == "pushHelper"

	flag.Parse()
	if len(flag.Args()) == 1 {
		tsDeviceURL = flag.Arg(0)
	}

	if isPushHelper {
		homeDir = "/home/phablet"
	} else {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		homeDir = user.HomeDir
	}
	cacheDir = filepath.Join(homeDir, ".cache/", appName)
	configDir = filepath.Join(homeDir, ".config/", appName)
	contactsFile = filepath.Join(configDir, "contacts.yml")
	settingsFile = filepath.Join(configDir, "settings.yml")
	os.MkdirAll(configDir, 0700)
	dataDir = filepath.Join(homeDir, ".local", "share", appName)
	attachDir = filepath.Join(dataDir, "attachments")
	os.MkdirAll(attachDir, 0700)
	storageDir = filepath.Join(dataDir, ".storage")
	if err := setupDB(); err != nil {
		log.Fatal(err)
	}
}

func runBackend() {
	client := &textsecure.Client{
		GetConfig:           getConfig,
		GetPhoneNumber:      getPhoneNumber,
		GetVerificationCode: getVerificationCode,
		GetStoragePassword:  getStoragePassword,
		MessageHandler:      messageHandler,
		ReceiptHandler:      receiptHandler,
		RegistrationDone:    registrationDone,
	}

	if isPhone {
		client.GetLocalContacts = getAddressBookContactsFromContentHub
	} else {
		client.GetLocalContacts = getDesktopContacts
	}

	err := textsecure.Setup(client)
	if _, ok := err.(*strconv.NumError); ok {
		showError(fmt.Errorf("Switching to unencrypted session store, removing %s\nThis will reset your sessions and reregister your phone.\n", storageDir))
		os.RemoveAll(storageDir)
		os.Exit(1)
	}
	if err != nil {
		showError(err)
		return
	}

	if exists(contactsFile) {
		api.HasContacts = true
		refreshContacts()
	}

	sendUnsentMessages()

	for {
		if err := textsecure.StartListening(); err != nil {
			log.Println(err)
			time.Sleep(3 * time.Second)
		}
	}
}

func sendUnsentMessages() {
	for _, s := range sessionsModel.sessions {
		for _, m := range s.messages {
			if m.Outgoing && !m.IsSent {
				go sendMessage(s, m)
			}
		}
	}
}

func main() {
	setup()
	if isPushHelper {
		pushHelperProcess()
	}

	err := qml.Run(runUI)
	if err != nil {
		log.Fatal(err)
	}
}

var engine *qml.Engine
var win *qml.Window

type textsecureAPI struct {
	HasContacts     bool
	ActiveSessionID string
}

var api = &textsecureAPI{}

func sendMessage(s *Session, m *Message) {
	var att io.Reader
	var err error

	if m.Attachment != "" {
		att, err = os.Open(m.Attachment)
		if err != nil {
			return
		}
	}

	ts := sendMessageLoop(s.Tel, m.Message, s.IsGroup, att, m.Flags)

	m.SentAt = ts
	s.Timestamp = m.SentAt
	m.IsSent = true
	qml.Changed(m, &m.IsSent)
	m.HTime = humanizeTimestamp(m.SentAt)
	qml.Changed(m, &m.HTime)
	s.When = m.HTime
	qml.Changed(s, &s.When)
	updateMessageSent(m)
	updateSession(s)
}

func sendMessageLoop(to, message string, group bool, att io.Reader, flags int) uint64 {
	var err error
	var ts uint64
	for {
		err = nil
		if flags == msgFlagResetSession {
			ts, err = textsecure.EndSession(to, "TERMINATE")
		} else if flags == msgFlagGroupLeave {
			err = textsecure.LeaveGroup(to)
		} else if flags == msgFlagGroupUpdate {
			_, err = textsecure.UpdateGroup(to, groups[to].Name, strings.Split(groups[to].Members, ","))
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
	nowms := uint64(time.Now().UnixNano() / 1000000)
	if ts > nowms {
		ts = nowms
	}
	return humanize.Time(time.Unix(0, int64(1000000*ts)))
}

func (api *textsecureAPI) SendMessage(to, message string) error {
	return sendMessageHelper(to, message, "")
}

// copyAttachment makes a copy of a file that is in the volatile content hub cache
func copyAttachment(src string) (string, error) {
	_, b := filepath.Split(src)
	dest := filepath.Join(attachDir, b)
	err := helpers.CopyFile(src, dest, helpers.CopyFlagOverwrite)
	if err != nil {
		return "", err
	}
	return dest, nil
}

func sendMessageHelper(to, message, file string) error {
	var err error
	if file != "" {
		file, err = copyAttachment(file)
		if err != nil {
			return err
		}
	}
	session := sessionsModel.Get(to)
	m := session.Add(message, "", file, "", true)
	saveMessage(m)
	go sendMessage(session, m)
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

	go sendMessageHelper(to, message, file)
	return nil
}

func (api *textsecureAPI) EndSession(tel string) error {
	session := sessionsModel.Get(tel)
	m := session.Add(sessionReset, "", "", "", true)
	m.Flags = msgFlagResetSession
	saveMessage(m)
	go sendMessage(session, m)
	return nil
}

// MarkSessionsRead marks one or all sessions as read
func (api *textsecureAPI) MarkSessionsRead(tel string) {
	if tel != "" {
		s := sessionsModel.Get(tel)
		s.MarkRead()
		return
	}
	for _, s := range sessionsModel.sessions {
		s.MarkRead()
	}
}

func (api *textsecureAPI) DeleteSession(tel string) {
	err := deleteSession(tel)
	if err != nil {
		showError(err)
	}
}

func (api *textsecureAPI) DeleteMessage(msg *Message, tel string) {
	deleteMessage(msg.ID)
	s := sessionsModel.Get(tel)
	for i, m := range s.messages {
		if m.ID == msg.ID {
			s.messages = append(s.messages[:i], s.messages[i+1:]...)
			s.Len--
			qml.Changed(s, &s.Len)
			return
		}
	}
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
	msg := session.Add(groupUpdateMsg(append(m, config.Tel), name), "", "", "", true)
	msg.Flags = msgFlagGroupNew
	qml.Changed(msg, &msg.Flags)
	saveMessage(msg)

	return nil

}

// membersDiffAndUnion returns a set diff and union of two contact sets represented as
// comma separated strings.
func membersDiffAndUnion(aa, bb string) ([]string, string) {

	if bb == "" {
		return nil, aa
	}

	as := strings.Split(aa, ",")
	bs := strings.Split(bb, ",")

	rs := []string{}

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
	return rs, strings.Join(append(as, rs...), ",")
}

func (api *textsecureAPI) UpdateGroup(hexid, name string, members string) error {
	g, ok := groups[hexid]
	if !ok {
		return fmt.Errorf("Unknown group id %s\n", hexid)
	}
	dm, members := membersDiffAndUnion(g.Members, members)
	groups[hexid] = &GroupRecord{
		GroupID: hexid,
		Name:    name,
		Members: members,
		Active:  g.Active,
		Avatar:  g.Avatar,
	}
	updateGroup(groups[hexid])
	session := sessionsModel.Get(hexid)
	msg := session.Add(groupUpdateMsg(dm, name), "", "", "", true)
	msg.Flags = msgFlagGroupUpdate
	qml.Changed(msg, &msg.Flags)
	saveMessage(msg)
	session.Name = name
	qml.Changed(session, &session.Name)
	updateSession(session)
	go sendMessage(session, msg)
	return nil
}

func (api *textsecureAPI) LeaveGroup(hexid string) error {
	session := sessionsModel.Get(hexid)
	msg := session.Add(youLeftGroup, "", "", "", true)
	msg.Flags = msgFlagGroupLeave
	qml.Changed(msg, &msg.Flags)
	saveMessage(msg)
	session.Active = false
	qml.Changed(session, &session.Active)
	groups[hexid].Active = false
	err := updateGroup(groups[hexid])
	go sendMessage(session, msg)
	return err
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

func (api *textsecureAPI) AvatarImage(id string) string {
	url := ""

	if c := getContactForTel(id); c != nil {
		if c.Photo != "" {
			url = "image://avatar/" + id
		}
	}
	if g, ok := groups[id]; ok {
		if len(g.Avatar) > 0 {
			url = "image://avatar/" + id
		}
	}
	return url
}

func (api *textsecureAPI) IdentityInfo(id string) string {
	myID := textsecure.MyIdentityKey()
	theirID, err := textsecure.ContactIdentityKey(id)
	if err != nil {
		log.Println(err)
	}
	return gettext.Gettext("Their identity (they read):") + "<br>" + fmt.Sprintf("% 0X", theirID) + "<br><br>" +
		gettext.Gettext("Your identity (you read):") + "<br><br>" + fmt.Sprintf("% 0X", myID)
}

func (api *textsecureAPI) RegisterUPS(token string) error {
	err := textsecure.RegisterWithUPS(token)
	if err != nil {
		log.Println(err)
	}
	return err
}

func avatarImageProvider(id string, width, height int) image.Image {
	var r io.Reader

	if c := getContactForTel(id); c != nil {
		r = strings.NewReader(c.Photo)
	}

	if g, ok := groups[id]; ok {
		r = bytes.NewReader(g.Avatar)
	}

	if r == nil {
		return image.NewAlpha(image.Rectangle{})
	}
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
