package i3

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// If your computer takes more than 10s to restart i3, it must be seriously
// overloaded, in which case we are probably doing you a favor by erroring out.
const reconnectTimeout = 10 * time.Second

// remote is a singleton containing the socket path and auto-detected byte order
// which i3 is using. It is lazily initialized by getIPCSocket.
var remote struct {
	path  string
	order binary.ByteOrder
	mu    sync.Mutex
}

// SocketPathHook Provides a way to override the default socket path lookup mechanism. Overriding this is unsupported.
var SocketPathHook = func() (string, error) {
	out, err := exec.Command("i3", "--get-socketpath").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("getting i3 socketpath: %v (output: %s)", err, out)
	}
	return string(out), nil
}

func getIPCSocket(updateSocketPath bool) (*socket, net.Conn, error) {
	remote.mu.Lock()
	defer remote.mu.Unlock()
	path := remote.path
	if (!wasRestart && updateSocketPath) || remote.path == "" {
		out, err := SocketPathHook()
		if err != nil {
			return nil, nil, err
		}
		path = strings.TrimSpace(string(out))
	}
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, nil, err
	}
	remote.path = path
	if remote.order == nil {
		remote.order, err = detectByteOrder(conn)
		if err != nil {
			conn.Close()
			return nil, nil, err
		}
	}

	return &socket{conn: conn, order: remote.order}, conn, err
}

type messageType uint32

const (
	messageTypeRunCommand messageType = iota
	messageTypeGetWorkspaces
	messageTypeSubscribe
	messageTypeGetOutputs
	messageTypeGetTree
	messageTypeGetMarks
	messageTypeGetBarConfig
	messageTypeGetVersion
	messageTypeGetBindingModes
	messageTypeGetConfig
	messageTypeSendTick
	messageTypeSync
)

var messageAtLeast = map[messageType]majorMinor{
	messageTypeRunCommand:      {4, 0},
	messageTypeGetWorkspaces:   {4, 0},
	messageTypeSubscribe:       {4, 0},
	messageTypeGetOutputs:      {4, 0},
	messageTypeGetTree:         {4, 0},
	messageTypeGetMarks:        {4, 1},
	messageTypeGetBarConfig:    {4, 1},
	messageTypeGetVersion:      {4, 3},
	messageTypeGetBindingModes: {4, 13},
	messageTypeGetConfig:       {4, 14},
	messageTypeSendTick:        {4, 15},
	messageTypeSync:            {4, 16},
}

const (
	messageReplyTypeCommand messageType = iota
	messageReplyTypeWorkspaces
	messageReplyTypeSubscribe
)

var magic = [6]byte{'i', '3', '-', 'i', 'p', 'c'}

type header struct {
	Magic  [6]byte
	Length uint32
	Type   messageType
}

type message struct {
	Type    messageType
	Payload []byte
}

type socket struct {
	conn  io.ReadWriter
	order binary.ByteOrder
}

func (s *socket) recvMsg() (message, error) {
	if s == nil {
		return message{}, fmt.Errorf("not connected")
	}
	var h header
	if err := binary.Read(s.conn, s.order, &h); err != nil {
		return message{}, err
	}
	msg := message{
		Type:    h.Type,
		Payload: make([]byte, h.Length),
	}
	_, err := io.ReadFull(s.conn, msg.Payload)
	return msg, err
}

func (s *socket) roundTrip(t messageType, payload []byte) (message, error) {
	if s == nil {
		return message{}, fmt.Errorf("not connected")
	}

	if err := binary.Write(s.conn, s.order, &header{magic, uint32(len(payload)), t}); err != nil {
		return message{}, err
	}
	if len(payload) > 0 { // skip empty Write()s for net.Pipe
		_, err := s.conn.Write(payload)
		if err != nil {
			return message{}, err
		}
	}
	return s.recvMsg()
}

// defaultSock is a singleton, lazily initialized by roundTrip. All
// request/response messages are sent to i3 via this socket, whereas
// subscriptions use their own connection.
var defaultSock struct {
	sock *socket
	conn net.Conn
	mu   sync.Mutex
}

// roundTrip sends a message to i3 and returns the received result in a
// concurrency-safe fashion.
func roundTrip(t messageType, payload []byte) (message, error) {
	// Error out early in case the message type is not yet supported by the
	// running i3 version.
	if t != messageTypeGetVersion {
		if err := AtLeast(messageAtLeast[t].major, messageAtLeast[t].minor); err != nil {
			return message{}, err
		}
	}

	defaultSock.mu.Lock()
	defer defaultSock.mu.Unlock()

Outer:
	for {
		msg, err := defaultSock.sock.roundTrip(t, payload)
		if err == nil {
			return msg, nil // happy path: success
		}

		// reconnect
		start := time.Now()
		for time.Since(start) < reconnectTimeout && (defaultSock.sock == nil || i3Running()) {
			if defaultSock.sock != nil {
				defaultSock.conn.Close()
			}
			defaultSock.sock, defaultSock.conn, err = getIPCSocket(defaultSock.sock != nil)
			if err == nil {
				continue Outer
			}

			// Reconnect within [10, 20) ms to prevent CPU-starving i3.
			time.Sleep(time.Duration(10+rand.Int63n(10)) * time.Millisecond)
		}
		return msg, err
	}
}
