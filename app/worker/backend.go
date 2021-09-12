package worker

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/signal-golang/textsecure/rootCa"
	log "github.com/sirupsen/logrus"
)

func RunRustBackend() {
	var cmd *exec.Cmd
	log.Infoln("[axolotl] Starting crayfish-backend")
	if _, err := os.Stat("./crayfish"); err == nil {
		cmd = exec.Command("./crayfish")
	} else {
		cmd = exec.Command("./backend/target/debug/crayfish")
	}
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		log.Fatalf("[axolotl] Starting crayfish-backend cmd.Start() failed with '%s'\n", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = cmd.Wait()
	if errStdout != nil || errStderr != nil {
		log.Fatal("[axolotl] failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdout), string(stderr)
	log.Infof("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	log.Infof("[axolotl] Crayfish-backend finished with error: %v", err)

}
func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

type CryfishWebSocketMessageType int32

const (
	CryfishWebSocketMessage_UNKNOWN  CryfishWebSocketMessageType = 0
	CryfishWebSocketMessage_REQUEST  CryfishWebSocketMessageType = 1
	CryfishWebSocketMessage_RESPONSE CryfishWebSocketMessageType = 2
)

type CryfishWebSocketMessage struct {
	Type     *CryfishWebSocketMessageType     `json:"type,omitempty"`
	Request  *CryfishWebSocketRequestMessage  `json:"request,omitempty"`
	Response *CryfishWebSocketResponseMessage `json:"response,omitempty"`
}
type CryfishWebSocketRequestMessageType int32

const (
	CryfishWebSocketRequestMessageTyp_UNKNOWN              CryfishWebSocketRequestMessageType = 0
	CryfishWebSocketRequestMessageTyp_START_REGISTRATION   CryfishWebSocketRequestMessageType = 1
	CryfishWebSocketRequestMessageTyp_CONFIRM_REGISTRATION CryfishWebSocketRequestMessageType = 2
)

type CryfishWebSocketRequestMessage struct {
	Type    *CryfishWebSocketRequestMessageType `json:"type,omitempty"`
	Message *string                             `json:"message,omitempty"`
}

type CryfishWebSocketResponseMessageType int32

const (
	CryfishWebSocketResponseMessageTyp_UNKNOWN              CryfishWebSocketResponseMessageType = 0
	CryfishWebSocketResponseMessageTyp_ACK                  CryfishWebSocketResponseMessageType = 1
	CryfishWebSocketResponseMessageTyp_CONFIRM_REGISTRATION CryfishWebSocketResponseMessageType = 2
)

type CryfishWebSocketResponseMessage struct {
	Type    *CryfishWebSocketResponseMessageType `json:"type,omitempty"`
	Message *string                              `json:"message,omitempty"`
}

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

// Conn is a wrapper for the websocket connection
type Conn struct {
	// The websocket connection
	ws *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte
}

var wsconn *Conn

// Connect to Signal websocket API at originURL with user and pass credentials
func (c *Conn) connect(originURL string) error {
	u, _ := url.Parse(originURL)

	log.Debugf("[axolotl] cryfish websocket Connecting to signal-server")

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

	log.Debugf("[axolotl] cryfish websocket Connected successfully")

	return nil
}

// Send ack response message
func (c *Conn) sendAck(id uint64) error {
	typ := CryfishWebSocketMessage_RESPONSE
	message := "OK"

	csm := &CryfishWebSocketMessage{
		Type: &typ,
		Response: &CryfishWebSocketResponseMessage{
			Message: &message,
		},
	}

	b, err := json.Marshal(csm)
	if err != nil {
		return err
	}

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
		log.Debugf("[axolotl] cryfish closing writeWorker")
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				log.Errorf("[axolotl] cryfish failed to read message from channel")
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			log.Debugf("[axolotl] cryfish websocket sending message")
			if err := c.write(websocket.BinaryMessage, message); err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl] cryfish Failed to send websocket message")
				return
			}
		case <-ticker.C:
			log.Debugf("[axolotl] cryfish Sending websocket ping message")
			if err := c.write(websocket.PingMessage, nil); err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl] cryfish Failed to send websocket ping message")
				return
			}
		}
	}
}

// StartListening connects to the server and handles incoming websocket messages.
func BackendStartListening() error {
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
		log.Debugf("[axolotl] cryfish Received websocket pong message")
		wsconn.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, bmsg, err := wsconn.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debugf("[axolotl] cryfish Websocket UnexpectedCloseError: %s", err)
			}
			return err
		}

		csm := &CryfishWebSocketMessage{}
		err = json.Unmarshal(bmsg, csm)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("[axolotl] cryfish Failed to unmarshal websocket message")
			return err
		}
		if *csm.Type == CryfishWebSocketMessage_REQUEST {
			err = handleCryfishRequestMessage(csm.Request)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl] cryfish Failed to handle received request message")
			}
		} else if *csm.Type == CryfishWebSocketMessage_RESPONSE {
			err = handleCryfishResponseMessage(csm.Response)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("[axolotl] cryfish Failed to handle received request message")
			}

		} else {
			log.Errorln("[axolotl] cryfish failed to handle incoming websocket message")
		}
		err = wsconn.sendAck(200)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("[axolotl] cryfish Failed to send ack")
			return err
		}

	}

}

// ErrNotListening is returned when trying to stop listening when there's no
// valid listening connection set up
var ErrNotListening = errors.New("[axolotl] cryfish there is no listening connection to stop")

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

func handleCryfishRequestMessage(request *CryfishWebSocketRequestMessage) error {
	return nil

}

func handleCryfishResponseMessage(response *CryfishWebSocketResponseMessage) error {
	return nil
}
