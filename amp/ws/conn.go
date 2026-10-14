package ws

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/pkg/errors"
)

// Conn handles sending and reciving on websocket connection
type Conn struct {
	tcpConn net.Conn
	cap     connCap
	no      uint64

	// backendHeaders can only be set and read on the backend.
	backendHeaders map[string]string
}

// connCap connection capabilities and usefull atributes
type connCap struct {
	deflateSupported bool
	deflateDisabled  bool
	userAgent        string
	forwardedFor     string
	host             string
	meta             map[string]string
	headers          map[string]string
	cookie           string
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
		no:      no(),
	}
	return c
}

// Headers usefull http headers
func (c *Conn) Headers() map[string]string {
	return c.cap.headers
}

func (c *Conn) SetBackendHeaders(headers map[string]string) {
	if c.backendHeaders == nil {
		c.backendHeaders = make(map[string]string)
	}

	for k, v := range headers {
		c.backendHeaders[k] = v
	}
}

func (c *Conn) GetBackendHeaders() map[string]string {
	return c.backendHeaders
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
	header, err := ws.ReadHeader(c.tcpConn)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if header.Length < 0 || header.Length > 1000000 {
		return nil, fmt.Errorf("malformed: %d -- %t -- %s -- %s", header.Length, c.cap.deflateSupported, c.cap.forwardedFor, c.cap.userAgent)
	}
	payload := make([]byte, header.Length)
	_, err = io.ReadFull(c.tcpConn, payload)
	if err != nil {
		c.Close()
		return nil, errors.WithStack(err)
	}

	if header.OpCode == ws.OpClose {
		c.Close()
		return nil, errors.WithStack(io.EOF)
	}
	if header.OpCode == ws.OpContinuation {
		return nil, errors.WithStack(io.ErrUnexpectedEOF)
	}
	if header.OpCode == ws.OpPing {
		header.OpCode = ws.OpPong
		header.Masked = false
		_ = ws.WriteHeader(c.tcpConn, header)
		return nil, nil
	}
	if header.OpCode == ws.OpPong {
		return nil, nil
	}
	if header.Masked {
		ws.Cipher(payload, header.Mask, 0)
	}
	if header.Rsv1() {
		payload = undeflate(payload)
	}

	setDeadline(c.tcpConn)
	return payload, nil
}

// No returns connection identificator.
func (c *Conn) No() uint64 {
	return c.no
}

// DeflateSupported whether websocket connection supports per message deflate.
func (c *Conn) DeflateSupported() bool {
	return c.cap.deflateSupported
}

// Close starts clearing connection.
// Closes tcpConn, that will raise error on reading and break receiveLoop.
func (c *Conn) Close() error {
	return c.tcpConn.Close()
}

// Cookies from the requests which started connection
func (c *Conn) Meta() map[string]string {
	return c.cap.meta
}

func (c *Conn) SetMeta(m map[string]string) {
	for k, v := range m {
		c.cap.meta[k] = v
	}
}

func (c *Conn) GetRemoteIp() string {
	return c.cap.forwardedFor
}

func (c *Conn) GetCookie() string {
	return c.cap.cookie
}

// syncFlushMarker is stripped by the sender and appended by the receiver (RFC 7692 7.2.1).
var syncFlushMarker = []byte{0x00, 0x00, 0xff, 0xff}

// flateReader is what flate.NewReader returns.
type flateReader interface {
	io.ReadCloser
	flate.Resetter
}

// flateReaderPool pools decompressor state only; the output escapes to the caller.
var flateReaderPool = sync.Pool{New: func() interface{} {
	return flate.NewReader(bytes.NewReader(nil)).(flateReader)
}}

// emptyReader releases the payload on reset. Not an io.ByteReader, so flate keeps
// its internal bufio.Reader.
type emptyReader struct{}

func (emptyReader) Read([]byte) (int, error) { return 0, io.EOF }

// maxUndeflateSizeHint caps the up-front allocation, not the message size.
const maxUndeflateSizeHint = 8 << 20 // 8 MiB

// undeflatedSizeHint assumes amp JSON compresses about 8x.
// Clamp before multiplying, so compressed*8 cannot overflow.
func undeflatedSizeHint(compressed int) int {
	if compressed > maxUndeflateSizeHint/8 {
		return maxUndeflateSizeHint
	}
	return compressed * 8
}

// undeflate decompresses a websocket per-message-deflate payload
func undeflate(data []byte) []byte {
	r := flateReaderPool.Get().(flateReader)
	defer func() {
		_ = r.Reset(emptyReader{}, nil)
		flateReaderPool.Put(r)
	}()
	// MultiReader: appending to bytes.NewBuffer(data) would write into the read buffer.
	_ = r.Reset(io.MultiReader(bytes.NewReader(data), bytes.NewReader(syncFlushMarker)), nil)

	out := bytes.NewBuffer(make([]byte, 0, undeflatedSizeHint(len(data))))
	_, _ = io.Copy(out, r)
	return out.Bytes()
}
