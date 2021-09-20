package crayfish

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/axolotl/app/settings"
	"github.com/nanu-c/axolotl/app/ui"
	uuid "github.com/satori/go.uuid"
	"github.com/signal-golang/textsecure"
	"github.com/signal-golang/textsecure/rootCa"
	log "github.com/sirupsen/logrus"
)

var (
	wsconn         *Conn
	cmd            *exec.Cmd
	stopping       = false
	receiveChannel chan *CrayfishWebSocketResponseMessage
	// ErrNotListening is returned when trying to stop listening when there's no
	// valid listening connection set up
	ErrNotListening = errors.New("[axolotl-crayfish-ws] there is no listening connection to stop")
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 25 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Signal websocket endpoint
	websocketPath = "/libsignal"
	serverAdress  = "ws://localhost:9081"
)

type CrayfishWebSocketMessageType int32

const (
	CrayfishWebSocketMessage_UNKNOWN  CrayfishWebSocketMessageType = 0
	CrayfishWebSocketMessage_REQUEST  CrayfishWebSocketMessageType = 1
	CrayfishWebSocketMessage_RESPONSE CrayfishWebSocketMessageType = 2
)

type CrayfishWebSocketMessage struct {
	Type     *CrayfishWebSocketMessageType     `json:"type,omitempty"`
	Request  *CrayfishWebSocketRequestMessage  `json:"request,omitempty"`
	Response *CrayfishWebSocketResponseMessage `json:"response,omitempty"`
}
type CrayfishWebSocketRequestMessageType int32

const (
	CrayfishWebSocketRequestMessageTyp_UNKNOWN             CrayfishWebSocketRequestMessageType = 0
	CrayfishWebSocketRequestMessageTyp_START_REGISTRATION  CrayfishWebSocketRequestMessageType = 1
	CrayfishWebSocketRequestMessageTyp_VERIFY_REGISTRATION CrayfishWebSocketRequestMessageType = 2
)

type CrayfishWebSocketRequestMessage struct {
	Type    *CrayfishWebSocketRequestMessageType `json:"type,omitempty"`
	Message interface{}                          `json:"message,omitempty"`
}

type CrayfishWebSocketResponseMessageType int32

const (
	CrayfishWebSocketResponseMessageTyp_UNKNOWN             CrayfishWebSocketResponseMessageType = 0
	CrayfishWebSocketResponseMessageTyp_ACK                 CrayfishWebSocketResponseMessageType = 1
	CrayfishWebSocketResponseMessageTyp_VERIFY_REGISTRATION CrayfishWebSocketResponseMessageType = 2
)

type CrayfishWebSocketResponseMessage struct {
	Type    *CrayfishWebSocketResponseMessageType `json:"type,omitempty"`
	Message interface{}                           `json:"message,omitempty"`
}

type ACKMessage struct {
	Status string `json:"status"`
}

// Conn is a wrapper for the websocket connection
type Conn struct {
	// The websocket connection
	ws *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte
}

type CrayfishWebSocketRequest_REGISTER_MESSAGE struct {
	Number   string `json:"number"`
	Password string `json:"password"`
	Captcha  string `json:"captcha"`
	UseVoice bool   `json:"use_voice"`
}

type CrayfishWebSocketRequest_VERIFY_REGISTER_MESSAGE struct {
	Code         uint64   `json:"confirm_code"`
	Number       string   `json:"number"`
	Password     string   `json:"password"`
	SignalingKey [52]byte `json:"signaling_key"`
}
type CrayfishWebSocketResponse_VERIFY_REGISTER_MESSAGE struct {
	UUID           [16]byte `json:"uuid"`
	StorageCapable bool     `json:"storage_capable"`
}

func Run() {

	log.Infoln("[axolotl-crayfish] Starting crayfish-backend")
	path, err := exec.LookPath("crayfish")
	if err != nil {
		if _, err := os.OpenFile("./crayfish", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666); err == nil {
			cmd = exec.Command("./crayfish")
		} else if _, err := os.Stat("./crayfish/target/debug/crayfish"); err == nil {
			cmd = exec.Command("./crayfish/target/debug/crayfish")
		} else {
			log.Errorln("[axolotl-crayfish] crayfish not found")
			cmd = exec.Command("pwd")
		}
	} else {
		cmd = exec.Command(path)
	}
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		log.Fatalf("[axolotl-crayfish] Starting crayfish-backend cmd.Start() failed with '%s'\n", err)
	}

	go StartListening()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()
	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Errorf("[axolotl-crayfish] Starting crayfish-backend cmd.Wait() failed with '%s'\n", err)
		return
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("[axolotl-crayfish] failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdout), string(stderr)
	log.Infof("[axolotl-crayfish-ws] out: %s\n", outStr)
	log.Infof("[axolotl-crayfish-ws] err: %s\n", errStr)
	log.Infof("[axolotl-crayfish] Crayfish-backend finished with error: %v", err)
	cmd.Process.Kill()

}
func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			log.Println("[crayfish]", string(d))
			// _, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			log.Debugln("copy and capture", out)

			return out, err
		}
	}
}

// Connect to Signal websocket API at originURL with user and pass credentials
func (c *Conn) connect(originURL string) error {
	u, _ := url.Parse(originURL)

	log.Debugf("[axolotl-crayfish-ws] websocket connecting to crayfish-server")

	var err error
	d := &websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	d.NetDial = func(network, addr string) (net.Conn, error) { return net.Dial(network, u.Host) }
	d.TLSClientConfig = &tls.Config{
		RootCAs: rootCa.RootCA,
	}

	c.ws, _, err = d.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	log.Debugf("[axolotl-crayfish-ws] websocket Connected successfully")

	return nil
}

// Send ack response message
func (c *Conn) sendAck(id uint64) error {
	typ := CrayfishWebSocketMessage_RESPONSE
	message := ACKMessage{
		Status: "ok",
	}
	responseType := CrayfishWebSocketResponseMessageTyp_ACK
	csm := &CrayfishWebSocketMessage{
		Type: &typ,
		Response: &CrayfishWebSocketResponseMessage{
			Type:    &responseType,
			Message: &message,
		},
	}

	b, err := json.Marshal(csm)
	if err != nil {
		return err
	}
	log.Debugln("[axolotl-crayfish-ws] websocket sending ack response ")

	c.send <- b
	return nil
}

// write writes a message with the given message type and payload.
func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writeWorker writes messages to websocket connection
func (c *Conn) writeWorker() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Debugf("[axolotl-crayfish-ws] closing writeWorker")
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			log.Debugln("[axolotl-crayfish-ws] incoming ws send message", string(message))
			if !ok {
				log.Errorf("[axolotl-crayfish-ws] failed to read message from channel")
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			log.Debugf("[axolotl-crayfish-ws] websocket sending message")
			if err := c.write(websocket.TextMessage, message); err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl-crayfish-ws] Failed to send websocket message")
				return
			}
		case <-ticker.C:
			log.Debugf("[axolotl-crayfish-ws] Sending websocket ping message")
			if err := c.write(websocket.PingMessage, nil); err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl-crayfish-ws] Failed to send websocket ping message")
				return
			}
		}
	}
}
func StartListening() error {
	defer func() {
		log.Debugf("[axolotl-crayfish-ws] BackendStartListening")
		for {
			if !stopping {
				err := StartWebsocket()
				if err != nil && !stopping {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("[axolotl-crayfish-ws] Failed to start listening")
					time.Sleep(time.Second * 5)
				}
			} else {
				break
			}
		}
	}()
	return nil

}

// BackendStartWebsocket connects to the server and handles incoming websocket messages.
func StartWebsocket() error {
	var err error

	wsconn = &Conn{send: make(chan []byte, 256)}
	err = wsconn.connect(serverAdress + websocketPath)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	defer wsconn.ws.Close()

	// Can only have a single goroutine call write methods
	go wsconn.writeWorker()

	wsconn.ws.SetReadDeadline(time.Now().Add(pongWait))
	wsconn.ws.SetPongHandler(func(string) error {
		log.Debugf("[axolotl-crayfish-ws] Received websocket pong message")
		wsconn.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, bmsg, err := wsconn.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debugf("[axolotl-crayfish-ws] Websocket UnexpectedCloseError: %s", err)
			}
			return err
		}
		log.Debugln("[axolotl-crayfish-ws] incoming msg", string(bmsg))
		csm := &CrayfishWebSocketMessage{}
		err = json.Unmarshal(bmsg, csm)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("[axolotl-crayfish-ws] Failed to unmarshal websocket message")
			return err
		}
		if csm.Type == nil {
			log.Errorln("[axolotl-crayfish-ws] Websocket message type is nil", string(bmsg))
		} else if *csm.Type == CrayfishWebSocketMessage_REQUEST {
			err = handleCrayfishRequestMessage(csm.Request)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl-crayfish-ws] Failed to handle received request message")
			}
		} else if *csm.Type == CrayfishWebSocketMessage_RESPONSE {
			err = handleCrayfishResponseMessage(csm.Response)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl-crayfish-ws] Failed to handle received request message")
			}

		} else {
			log.Errorln("[axolotl-crayfish-ws] failed to handle incoming websocket message")
		}
		if csm.Type != nil {
			err = wsconn.sendAck(200)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl-crayfish-ws] Failed to send ack")
				return err
			}
		}
	}
}

// StopListening disables the receiving of messages.
func StopListening() error {
	if wsconn == nil {
		return ErrNotListening
	}
	if wsconn.ws != nil {
		wsconn.ws.Close()
	}
	return nil
}

func handleCrayfishRequestMessage(request *CrayfishWebSocketRequestMessage) error {
	log.Debugln("[axolotl-crayfish-ws] Received websocket request message", *request.Type)
	return nil
}

func handleCrayfishResponseMessage(response *CrayfishWebSocketResponseMessage) error {
	log.Debugln("[axolotl-crayfish-ws] Received websocket response message", *response.Type)
	if receiveChannel != nil && *response.Type > 1 {
		receiveChannel <- response
	}
	return nil
}

func CrayfishRegister(registrationInfo *textsecure.RegistrationInfo) (*textsecure.CrayfishRegistration, error) {
	log.Debugf("[axolotl-crayfish-ws] Registering via crayfish ask for phone")
	var phoneNumber string
	if !settings.SettingsModel.Registered {
		phoneNumber = ui.GetPhoneNumber()
	}
	log.Debugf("[axolotl-crayfish-ws] Registering via crayfish ask for captcha")
	captcha := ui.GetCaptchaToken()
	log.Debugf("[axolotl-crayfish-ws] Registering via crayfish build message")
	registerMessage := &CrayfishWebSocketRequest_REGISTER_MESSAGE{
		Number:   phoneNumber,
		Password: registrationInfo.Password(),
		Captcha:  captcha,
		UseVoice: false,
	}
	messageType := CrayfishWebSocketMessage_REQUEST
	requestType := CrayfishWebSocketRequestMessageTyp_START_REGISTRATION
	request := &CrayfishWebSocketRequestMessage{
		Type:    &requestType,
		Message: registerMessage,
	}
	registerRequestMessage := &CrayfishWebSocketMessage{
		Type:    &messageType,
		Request: request,
	}
	m, err := json.Marshal(registerRequestMessage)
	if err != nil {
		return nil, err
	}
	log.Debugf("[axolotl-crayfish-ws] Registering via crayfish send")
	wsconn.send <- m

	code := ui.GetVerificationCode()
	codeInt, err := strconv.ParseUint(code, 10, 32)
	if err != nil {
		return nil, err
	}
	requestType = CrayfishWebSocketRequestMessageTyp_VERIFY_REGISTRATION
	var signalingKey [52]byte
	copy(signalingKey[:], registrationInfo.SignalingKey())
	verificationMessage := &CrayfishWebSocketRequest_VERIFY_REGISTER_MESSAGE{
		Number:       phoneNumber,
		Code:         codeInt,
		SignalingKey: signalingKey,
		Password:     registrationInfo.Password(),
	}
	requestVerifyType := CrayfishWebSocketRequestMessageTyp_VERIFY_REGISTRATION
	verificationRequest := &CrayfishWebSocketRequestMessage{
		Type:    &requestVerifyType,
		Message: verificationMessage,
	}
	verificationRequestMessage := &CrayfishWebSocketMessage{
		Type:    &messageType,
		Request: verificationRequest,
	}
	mv, err := json.Marshal(verificationRequestMessage)
	if err != nil {
		return nil, err
	}
	log.Debugf("[axolotl-crayfish-ws] Registering via crayfish send verification")
	wsconn.send <- mv
	receiveChannel = make(chan *CrayfishWebSocketResponseMessage, 1)
	response := <-receiveChannel
	rm, err := json.Marshal(response.Message)
	if err != nil {
		return nil, err
	}
	var data CrayfishWebSocketResponse_VERIFY_REGISTER_MESSAGE
	json.Unmarshal(rm, &data)
	uuidString, err := uuid.FromBytes(data.UUID[:])
	if err != nil {
		return nil, err
	}
	log.Debugf("[axolotl-crayfish-ws] Registering via crayfish uuid %s", uuidString.String())
	return &textsecure.CrayfishRegistration{
		UUID: uuidString.String(),
		Tel:  phoneNumber,
	}, nil
}

func Stop() error {
	stopping = true
	err := StopListening()
	if err != nil {
		return err
	}
	if cmd != nil {
		err = cmd.Process.Signal(os.Interrupt)
		if err != nil {
			return err
		}
	}
	return nil
}
