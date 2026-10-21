package dbus

import (
	"bufio"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// blockingConn is a net.Conn whose reads block until they are released.
type blockingConn struct {
	release chan struct{}
	closed  chan struct{}
	once    sync.Once
}

func newBlockingConn() *blockingConn {
	return &blockingConn{release: make(chan struct{}), closed: make(chan struct{})}
}

func (c *blockingConn) Read([]byte) (int, error) {
	select {
	case <-c.release:
		return 0, io.ErrUnexpectedEOF
	case <-c.closed:
		return 0, net.ErrClosed
	}
}

func (c *blockingConn) Write(p []byte) (int, error) { return len(p), nil }

func (c *blockingConn) Close() error {
	c.once.Do(func() { close(c.closed) })
	return nil
}

func (c *blockingConn) LocalAddr() net.Addr              { return nil }
func (c *blockingConn) RemoteAddr() net.Addr             { return nil }
func (c *blockingConn) SetDeadline(time.Time) error      { return nil }
func (c *blockingConn) SetReadDeadline(time.Time) error  { return nil }
func (c *blockingConn) SetWriteDeadline(time.Time) error { return nil }

// Close must not close the socket under an in-flight Call.
func TestCloseWaitsForCall(t *testing.T) {
	bc := newBlockingConn()
	c := &Conn{c: bc, r: bufio.NewReader(bc)}

	callDone := make(chan error, 1)
	go func() {
		_, err := c.Call("org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus", "GetId", "")
		callDone <- err
	}()

	// The call holds the lock while it waits for the reply.
	time.Sleep(50 * time.Millisecond)
	closeDone := make(chan error, 1)
	go func() { closeDone <- c.Close() }()

	select {
	case <-closeDone:
		t.Fatal("Close returned while a Call was in flight")
	case <-time.After(100 * time.Millisecond):
	}

	close(bc.release)
	if err := <-callDone; err == nil {
		t.Error("Call succeeded, want an error from the broken connection")
	}
	if err := <-closeDone; err != nil {
		t.Errorf("Close: %v", err)
	}
}
