package amp

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"io"
	"sync"

	"github.com/mnu5/svckit/log"
)

// Message types
const (
	Publish uint8 = iota
	Subscribe
	Request
	Response
	Ping
	Pong
	Alive
)

// Topic update types
const (
	Diff   uint8 = iota // merge into topic
	Full                // replace entire topic
	Append              // append to the end of the topic
	Update              // replace existing topic entry
	Close               // last message for the topic, topic is closed after this
)

// podrzani nacini kompresije poruke
const (
	CompressionNone uint8 = iota
	CompressionDeflate
)

var (
	// ne komprimiramo poruke manje od
	CompressionLenLimit = 8 * 1024
	separtor            = []byte{10}
)

// Subscriber is the interface for subscribing to the topics
type Subscriber interface {
	Send(m *Msg)
}

// BodyMarshaler nesto sto se zna zapakovati
type BodyMarshaler interface {
	ToJSON() ([]byte, error)
}

// Msg ...
type Msg struct {
	// message type
	Type uint8 `json:"t,omitempty"`
	// request reponse messages attributes
	Method        string `json:"m,omitempty"`
	ReplyTo       string `json:"r,omitempty"`
	CorrelationID string `json:"i,omitempty"` // TODO mozda uint64
	ExpiresAt     int64  `json:"x,omitempty"` // TODO unused so far
	Error         string `json:"e,omitempty"`
	ErrorCode     int    `json:"c,omitempty"`
	// publish message attributes
	Topic      string `json:"o,omitempty"`
	Ts         int64  `json:"s,omitempty"`
	UpdateType uint8  `json:"u,omitempty"`
	Replay     uint8  `json:"p,omitempty"`
	// sub
	Subscriptions map[string]int64 `json:"b,omitempty"`

	body          []byte
	noCompression bool
	payloads      map[uint8][]byte
	src           BodyMarshaler
	sync.Mutex
}

// Parse decodes Msg from []byte
func Parse(buf []byte) *Msg {
	parts := bytes.SplitN(buf, separtor, 2)
	m := &Msg{}
	if err := json.Unmarshal(parts[0], m); err != nil {
		log.S("header", string(parts[0])).Error(err)
		return nil
	}
	if len(parts) > 1 {
		m.body = parts[1]
	}
	return m
}

// Undeflate enodes ws deflated message
func Undeflate(data []byte) []byte {
	buf := bytes.NewBuffer(data)
	buf.Write([]byte{0x00, 0x00, 0xff, 0xff})
	r := flate.NewReader(buf)
	defer r.Close()
	out := bytes.NewBuffer(nil)
	io.Copy(out, r)
	return out.Bytes()
}

func (m *Msg) Marshal() []byte {
	buf, _ := m.marshal(CompressionNone)
	return buf
}

func (m *Msg) MarshalDeflate() ([]byte, bool) {
	return m.marshal(CompressionDeflate)
}

// Payload encodes message into []byte
func (m *Msg) marshal(supportedCompression uint8) ([]byte, bool) {
	m.Lock()
	defer m.Unlock()
	compression := supportedCompression
	if m.noCompression {
		compression = CompressionNone
	}
	// check if we already have payload
	key := payloadKey(compression)
	if payload, ok := m.payloads[key]; ok {
		return payload, compression != CompressionNone
	}

	payload := m.payload()
	// decide wather we need compression
	if len(payload) < CompressionLenLimit {
		m.noCompression = true
		compression = CompressionNone
	}
	// compress
	if compression == CompressionDeflate {
		payload = deflate(payload)
	}
	// store payload
	if m.payloads == nil {
		m.payloads = make(map[uint8][]byte)
	}
	m.payloads[key] = payload

	return payload, compression != CompressionNone
}

func (m *Msg) payload() []byte {
	header, _ := json.Marshal(m)
	buf := bytes.NewBuffer(header)
	buf.Write(separtor)
	if m.body != nil {
		buf.Write(m.body)
	}
	if m.src != nil {
		body, _ := m.src.ToJSON()
		buf.Write(body)
	}
	return buf.Bytes()
}

func payloadKey(compression uint8) uint8 {
	return compression
}

func deflate(src []byte) []byte {
	dest := bytes.NewBuffer(nil)
	c, _ := flate.NewWriter(dest, flate.DefaultCompression)
	c.Write(src)
	c.Close()
	buf := dest.Bytes()
	if len(buf) > 4 {
		return buf[0 : len(buf)-4]
	}
	return buf
}

func (m *Msg) BodyTo(v interface{}) error {
	return json.Unmarshal(m.body, v)
}

func (m *Msg) Response(b BodyMarshaler) *Msg {
	return &Msg{
		Type:          Response,
		CorrelationID: m.CorrelationID,
		Method:        m.Method,
		src:           b,
	}
}

func (m *Msg) ResponseTransportError() *Msg {
	return &Msg{
		Type:          Response,
		CorrelationID: m.CorrelationID,
		Method:        m.Method,
		Error:         "transport error", // TODO
		ErrorCode:     -128,
	}
}

func (m *Msg) Request() *Msg {
	return &Msg{
		Type:          Request,
		CorrelationID: m.CorrelationID,
		Method:        m.Method,
		src:           m.src,
		body:          m.body,
	}
}

func NewAlive() *Msg {
	return &Msg{Type: Alive}
}

func NewPong() *Msg {
	return &Msg{Type: Pong}
}

func (m *Msg) IsPing() bool {
	return m.Type == Ping
}

func (m *Msg) IsAlive() bool {
	return m.Type == Alive
}

func NewPublish(topic string, ts int64, updateType uint8, b BodyMarshaler) *Msg {
	return &Msg{
		Type:       Publish,
		Topic:      topic,
		Ts:         ts,
		UpdateType: updateType,
		src:        b,
	}
}

// type Subscriptions []*Subscription

// type Subscription struct {
// 	Topic string `json:"o,omitempty"`
// 	Ts    int64  `json:"n,omitempty"`
// }

func (m *Msg) IsTopicClose() bool {
	return m.UpdateType == Close
}

func (m *Msg) IsReplay() bool {
	return m.Replay == 1
}

func (m *Msg) IsFull() bool {
	return m.UpdateType == Full
}

// AsReplay marks message as replay
func (m *Msg) AsReplay() *Msg {
	return &Msg{
		Type:       m.Type,
		Topic:      m.Topic,
		UpdateType: m.UpdateType,
		Replay:     1,
		Ts:         m.Ts,
		body:       m.body,
		src:        m.src,
	}
}
