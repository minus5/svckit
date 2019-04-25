package ws

import (
	"bytes"
	"compress/flate"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/pkg/errors"
)

// Conn handles sending and reciving on websocket connection
type Conn struct {
	tcpConn    net.Conn
	cap        connCap
	receive    chan []byte
	receiveErr error
	no         uint64
}

// connCap connection capabilities and usefull atributes
type connCap struct {
	deflateSupported bool
	userAgent        string
	forwardedFor     string
}

var (
	maxQueueLen        = 1024
	aliveInterval      = 32 * time.Second
	tcpDeadline        = 48 * time.Second // 1.5 * aliveInterval
	connectionsCounter uint64
)

func setDeadline(c net.Conn) {
	_ = c.SetDeadline(time.Now().Add(tcpDeadline))
}

func no() uint64 {
	return atomic.AddUint64(&connectionsCounter, 1)
}

func newConn(tc net.Conn, cap connCap) *Conn {
	c := &Conn{
		tcpConn: tc,
		cap:     cap,
		receive: make(chan []byte),
		no:      no(),
	}
	return c
}

// Headers usefull http headers
func (c *Conn) Headers() map[string]string {
	return nil
}

// Write writes payload to the websocket connection.
func (c *Conn) Write(payload []byte, deflated bool) error {
	var header ws.Header
	header.OpCode = ws.OpText
	header.Length = int64(len(payload))
	header.Fin = true
	if deflated {
		header.Rsv = ws.Rsv(true, false, false)
	}
	if err := ws.WriteHeader(c.tcpConn, header); err != nil {
		_ = c.Close()
		return errors.WithStack(err)
	}
	_, err := c.tcpConn.Write(payload)
	if err == nil {
		setDeadline(c.tcpConn)
	} else {
		_ = c.Close()
	}
	return errors.WithStack(err)
}

// Read reads message from the connection.
func (c *Conn) Read() ([]byte, error) {
	buf, ok := <-c.receive
	if !ok {
		return nil, c.receiveErr
	}
	return buf, nil
}

// No returns connection identificator.
func (c *Conn) No() uint64 {
	return c.no
}

// DeflateSupported whether websocket connection supports per message deflate.
func (c *Conn) DeflateSupported() bool {
	return c.cap.deflateSupported
}

// wait blocks until connection is closed
func (c *Conn) wait() {
	c.receiveLoop()
}

// Close starts clearing connection.
// Closes tcpConn, that will raise error on reading and break receiveLoop.
func (c *Conn) Close() error {
	return c.tcpConn.Close()
}

func (c *Conn) receiveLoop() {
	defer close(c.receive)
	for {
		header, err := ws.ReadHeader(c.tcpConn)
		if err != nil {
			c.receiveErr = errors.WithStack(err)
			break
		}

		payload := make([]byte, header.Length)
		_, err = io.ReadFull(c.tcpConn, payload)
		if err != nil {
			c.receiveErr = errors.WithStack(err)
			c.Close()
			break
		}

		if header.OpCode == ws.OpClose {
			c.receiveErr = errors.WithStack(io.EOF)
			c.Close()
			break
		}
		if header.OpCode == ws.OpContinuation {
			// TODO not implemented
			c.receiveErr = errors.WithStack(io.ErrUnexpectedEOF)
			break
		}
		if header.OpCode == ws.OpPing {
			header.OpCode = ws.OpPong
			header.Masked = false
			_ = ws.WriteHeader(c.tcpConn, header)
			continue
		}
		if header.OpCode == ws.OpPong {
			continue
		}
		if header.Masked {
			ws.Cipher(payload, header.Mask, 0)
		}
		if header.Rsv1() {
			payload = undeflate(payload)
		}

		c.receive <- payload
		setDeadline(c.tcpConn)
	}
}

// undeflate uncomresses websocket payload
func undeflate(data []byte) []byte {
	buf := bytes.NewBuffer(data)
	buf.Write([]byte{0x00, 0x00, 0xff, 0xff})
	r := flate.NewReader(buf)
	defer r.Close()
	out := bytes.NewBuffer(nil)
	_, _ = io.Copy(out, r)
	return out.Bytes()
}
