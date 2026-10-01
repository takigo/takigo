package dbus

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrNoBus is returned by SessionBus when there is no session bus to
// connect to.
var ErrNoBus = errors.New("dbus: no session bus")

// Error is an error reply from the bus or the called service.
type Error struct {
	Name    string
	Message string
}

func (e *Error) Error() string { return "dbus: " + e.Name + ": " + e.Message }

// Conn is a connection to a message bus. Calls are serialized.
type Conn struct {
	mu     sync.Mutex
	c      net.Conn
	r      *bufio.Reader
	serial uint32
}

const (
	typeCall   = 1
	typeReturn = 2
	typeError  = 3

	fieldPath        = 1
	fieldInterface   = 2
	fieldMember      = 3
	fieldErrorName   = 4
	fieldReplySerial = 5
	fieldDestination = 6
	fieldSignature   = 8
)

func sortedKeys(m map[string]Variant) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// sessionAddress returns the unix socket to dial for the session bus.
func sessionAddress() (string, error) {
	addr := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	if addr == "" {
		if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
			addr = "unix:path=" + dir + "/bus"
		}
	}
	for part := range strings.SplitSeq(addr, ";") {
		rest, ok := strings.CutPrefix(part, "unix:")
		if !ok {
			continue
		}
		for kv := range strings.SplitSeq(rest, ",") {
			k, v, _ := strings.Cut(kv, "=")
			switch k {
			case "path":
				return v, nil
			case "abstract":
				return "@" + v, nil
			}
		}
	}
	return "", ErrNoBus
}

// SessionBus connects to the user's session bus.
func SessionBus() (*Conn, error) {
	path, err := sessionAddress()
	if err != nil {
		return nil, err
	}
	c, err := net.DialTimeout("unix", path, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoBus, err)
	}
	conn := &Conn{c: c, r: bufio.NewReader(c)}
	if err := conn.auth(); err != nil {
		c.Close()
		return nil, err
	}
	if _, err := conn.Call("org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus", "Hello", ""); err != nil {
		c.Close()
		return nil, err
	}
	return conn, nil
}

// Close closes the connection.
func (c *Conn) Close() error { return c.c.Close() }

// auth does the SASL EXTERNAL handshake: the bus checks the socket's
// credentials against the uid we name.
func (c *Conn) auth() error {
	_ = c.c.SetDeadline(time.Now().Add(2 * time.Second))
	defer func() { _ = c.c.SetDeadline(time.Time{}) }()
	uid := hex.EncodeToString([]byte(strconv.Itoa(os.Getuid())))
	if _, err := c.c.Write([]byte("\x00AUTH EXTERNAL " + uid + "\r\n")); err != nil {
		return err
	}
	line, err := c.r.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, "OK ") {
		return fmt.Errorf("dbus: authentication refused: %s", strings.TrimSpace(line))
	}
	_, err = c.c.Write([]byte("BEGIN\r\n"))
	return err
}

// Call invokes a method and returns the values of its reply. sig is the
// signature of args. It waits up to five seconds for the reply.
func (c *Conn) Call(dest string, path ObjectPath, iface, member, sig string, args ...any) ([]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var body encoder
	if err := body.values(sig, args); err != nil {
		return nil, err
	}
	c.serial++
	serial := c.serial

	fields := []any{
		[]any{byte(fieldPath), Variant{"o", path}},
		[]any{byte(fieldInterface), Variant{"s", iface}},
		[]any{byte(fieldMember), Variant{"s", member}},
		[]any{byte(fieldDestination), Variant{"s", dest}},
	}
	if sig != "" {
		fields = append(fields, []any{byte(fieldSignature), Variant{"g", sig}})
	}
	var msg encoder
	msg.buf = append(msg.buf, 'l', typeCall, 0, 1)
	msg.u32(uint32(len(body.buf)))
	msg.u32(serial)
	if err := msg.value("a(yv)", fields); err != nil {
		return nil, err
	}
	msg.align(8)
	msg.buf = append(msg.buf, body.buf...)

	_ = c.c.SetDeadline(time.Now().Add(5 * time.Second))
	defer func() { _ = c.c.SetDeadline(time.Time{}) }()
	if _, err := c.c.Write(msg.buf); err != nil {
		return nil, err
	}
	for {
		m, err := c.read()
		if err != nil {
			return nil, err
		}
		if m.replySerial != serial {
			continue // a signal, or a reply to nothing we are waiting for
		}
		if m.typ == typeError {
			e := &Error{Name: m.errorName}
			if len(m.body) > 0 {
				e.Message, _ = m.body[0].(string)
			}
			return nil, e
		}
		return m.body, nil
	}
}

type message struct {
	typ         byte
	replySerial uint32
	errorName   string
	body        []any
}

func (c *Conn) read() (*message, error) {
	head := make([]byte, 16)
	if _, err := io.ReadFull(c.r, head); err != nil {
		return nil, err
	}
	var order binary.ByteOrder = binary.LittleEndian
	if head[0] == 'B' {
		order = binary.BigEndian
	}
	bodyLen := order.Uint32(head[4:])
	fieldsLen := order.Uint32(head[12:])
	const maxMessage = 1 << 27 // the specification's limit
	if bodyLen > maxMessage || fieldsLen > maxMessage {
		return nil, fmt.Errorf("%w: oversized", ErrBadMessage)
	}
	padded := (int(fieldsLen) + 7) &^ 7
	rest := make([]byte, padded+int(bodyLen))
	if _, err := io.ReadFull(c.r, rest); err != nil {
		return nil, err
	}

	// The header fields array starts at offset 12 of the message, which
	// keeps the alignment of its structs: decode it in place.
	d := decoder{buf: append(head, rest[:fieldsLen]...), pos: 12, order: order}
	fields, err := d.value("a(yv)")
	if err != nil {
		return nil, err
	}
	m := &message{typ: head[1]}
	signature := ""
	for _, f := range fields.([]any) {
		pair := f.([]any)
		code, v := pair[0].(byte), pair[1].(Variant).Value
		switch code {
		case fieldReplySerial:
			m.replySerial, _ = v.(uint32)
		case fieldErrorName:
			m.errorName, _ = v.(string)
		case fieldSignature:
			signature, _ = v.(string)
		}
	}
	body := decoder{buf: rest[padded:], order: order}
	m.body, err = body.values(signature)
	return m, err
}
