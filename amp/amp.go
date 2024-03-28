package amp

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/minus5/svckit/pkg/compress"

	"github.com/minus5/svckit/log"
)

var (
	jsonSerializer    = jsoniter.Config{TagKey: "json"}.Froze()
	backendSerializer = jsoniter.Config{TagKey: "backend"}.Froze()
)

// Message types
const (
	Publish   uint8 = iota // stream updated message, view update types below
	Subscribe              // subscribe on a topic or many topics
	Request                // request for some data
	Response               // response on request
	Ping                   // query weather other side is there
	Pong                   // replay to ping
	Alive                  // signal that server side is still alive
	Current                // request for current state of a stream
	Event                  // event stream, no cache
	Meta                   // set session metadata
)

// Topic update types
const (
	Diff       uint8 = iota // merge into topic
	Full                    // replace entire topic
	Append                  // append to the end of the topic
	Update                  // replace existing topic entry
	Close                   // last message for the topic, topic is closed after this
	BurstStart              // indicate that there will be burst of messages for the topic ...
	BurstEnd                // so we can stop updating UI until we get BurstEnd message
)

// Error sources
const (
	ApplicationError uint8 = iota
	TransportError
)

// Replay types
const (
	Original uint8 = iota // original message
	Replay                // replay of the previously sent message
)

// supported compression types
const (
	CompressionNone uint8 = iota
	CompressionDeflate
)

const (
	CompatibilityVersionDefault uint8 = iota
	CompatibilityVersion1
)

var (
	compressionLenLimit = 8 * 1024 // do not compress messages smaller than
	separator           = []byte{10}
)

// Subscriber is the interface for subscribing to the topics
type Subscriber interface {
	Send(m *Msg)
}

type Sender interface {
	Meta() map[string]string
	Send(m *Msg)
	SendMsgs(ms []*Msg)
	Headers() map[string]string
}

// BodyMarshaler nesto sto se zna zapakovati
type BodyMarshaler interface {
	MarshalJSON() ([]byte, error)
}

// Msg basic application message structure
type Msg struct {
	Type           uint8             `json:"t,omitempty" backend:"t,omitempty"` // message type
	ReplyTo        string            `json:"r,omitempty" backend:"r,omitempty"` // topic to send replay to
	CorrelationID  uint64            `json:"i,omitempty" backend:"i,omitempty"` // correlationID between request and response
	Error          *Error            `json:"e,omitempty" backend:"e,omitempty"` // error description in response message
	URI            string            `json:"u,omitempty" backend:"u,omitempty"` // has structure: topic/path
	Ts             int64             `json:"s,omitempty" backend:"s,omitempty"` // timestamp unix milli
	UpdateType     uint8             `json:"p,omitempty" backend:"p,omitempty"` // explains how to handle publish message
	Replay         uint8             `json:"l,omitempty" backend:"l,omitempty"` // is this a re-play message (repeated)
	Subscriptions  map[string]int64  `json:"b,omitempty" backend:"b,omitempty"` // topics to subscribe to
	CacheDepth     int               `json:"d,omitempty" backend:"d,omitempty"` // cache depth for append update type messages
	Meta           map[string]string `json:"m,omitempty" backend:"m,omitempty"` // client session metadata
	BackendHeaders map[string]string `json:"-" backend:"h,omitempty"`           // exclusive for communication between backend services

	body          []byte
	noCompression bool
	payloads      map[uint8][]byte
	src           BodyMarshaler
	topic         string
	path          string

	sync.Mutex
}

// Error related attributes in the message
type Error struct {
	Source  uint8  `json:"s,omitempty"`
	Message string `json:"m,omitempty"`
	Code    int    `json:"c,omitempty"`
}

// Parse decodes Msg received from client.
func Parse(buf []byte) *Msg {
	return parse(buf, jsonSerializer)
}

// ParseFromBackend parses the message received from another backend service.
func ParseFromBackend(buf []byte) *Msg {
	return parse(buf, backendSerializer)
}

func parse(buf []byte, serializer jsoniter.API) *Msg {
	if buf == nil {
		return nil
	}
	parts := bytes.SplitN(buf, separator, 2)

	m := &Msg{}
	if err := serializer.Unmarshal(parts[0], m); err != nil {
		log.S("header", string(parts[0])).Error(err)
		return nil
	}

	if len(parts) > 1 {
		m.body = parts[1]
		body, err := compress.GunzipIf(m.body)
		if err != nil {
			log.Error(err)
			return nil
		} else {
			m.body = body
		}
	}

	return m
}

func ParseWithMeta(buf []byte, query url.Values) *Msg {
	m := Parse(buf)
	if m == nil {
		return nil
	}
	if len(query) > 0 {
		meta := make(map[string]string)
		for k, v := range query {
			meta[k] = strings.Join(v, ",")
		}
		if len(meta) > 0 {
			m.Meta = meta
		}
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

// Marshal the message that will be sent to the client.
func (m *Msg) Marshal() []byte {
	buf, _ := m.marshal(CompressionNone, CompatibilityVersionDefault, jsonSerializer)
	return buf
}

// MarshalForBackend is used when marshaling Msg for backend to backend communication.
func (m *Msg) MarshalForBackend() []byte {
	buf, _ := m.marshal(CompressionNone, CompatibilityVersionDefault, backendSerializer)
	return buf
}

// MarshalDeflate packs and compress message that will be sent to the client.
func (m *Msg) MarshalDeflate() ([]byte, bool) {
	return m.marshal(CompressionDeflate, CompatibilityVersionDefault, jsonSerializer)
}

// marshal encodes message into []byte
func (m *Msg) marshal(supportedCompression, version uint8, serializer jsoniter.API) ([]byte, bool) {
	if version == CompatibilityVersion1 {
		if m.UpdateType == BurstStart || m.UpdateType == BurstEnd {
			// unsuported mesage types in this version
			return nil, false
		}
	}
	m.Lock()
	defer m.Unlock()

	compression := supportedCompression
	if m.noCompression {
		compression = CompressionNone
	}
	// check if we already have payload
	key := payloadKey(compression, version)
	if payload, ok := m.payloads[key]; ok {
		return payload, compression != CompressionNone
	}

	payload := m.payload(version, serializer)
	// decide wather we need compression
	if len(payload) < compressionLenLimit {
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

func (m *Msg) payload(version uint8, serializer jsoniter.API) []byte {
	var header []byte
	if version == CompatibilityVersion1 {
		header = m.marshalV1header()
	} else {
		header, _ = serializer.Marshal(m)
	}
	buf := bytes.NewBuffer(header)
	buf.Write(separator)
	if m.body != nil {
		buf.Write(m.body)
	}
	if m.src != nil {
		body, _ := m.src.MarshalJSON()
		buf.Write(body)
	}
	return buf.Bytes()
}

func payloadKey(compression, version uint8) uint8 {
	return version*4 + compression
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

// BodyTo unmarshals message body to the v
func (m *Msg) BodyTo(v interface{}) error {
	return json.Unmarshal(m.body, v)
}

// Unmarshal unmarshals message body to the v
func (m *Msg) Unmarshal(v interface{}) error {
	return json.Unmarshal(m.body, v)
}

// Response creates response message from original request
func (m *Msg) Response(o interface{}) *Msg {
	return &Msg{
		Type:           Response,
		CorrelationID:  m.CorrelationID,
		src:            toBodyMarshaler(o),
		BackendHeaders: m.BackendHeaders,
	}
}

// BurstStart creates burst start message for the uri from the original message.
func (m *Msg) BurstStart() *Msg {
	return &Msg{
		Type:       Publish,
		URI:        m.URI,
		UpdateType: BurstStart,
		Ts:         m.Ts,
	}
}

// BurstEnd creates burst end message for the uri from the original message.
func (m *Msg) BurstEnd() *Msg {
	return &Msg{
		Type:       Publish,
		URI:        m.URI,
		UpdateType: BurstEnd,
		Ts:         m.Ts,
	}
}

// ResponseTransportError creates response message with error set to transport error
func (m *Msg) ResponseTransportError(err error) *Msg {
	return &Msg{
		Type:          Response,
		CorrelationID: m.CorrelationID,
		Error: &Error{
			Source:  TransportError,
			Message: err.Error(),
		},
	}
}

func (m *Msg) ResponseError(err error) *Msg {
	return &Msg{
		Type:          Response,
		CorrelationID: m.CorrelationID,
		Error: &Error{
			Source:  ApplicationError,
			Message: err.Error(),
		},
	}
}

// Request creates request type message from original message
func (m *Msg) Request() *Msg {
	return &Msg{
		Type:           Request,
		CorrelationID:  m.CorrelationID,
		URI:            m.URI,
		Meta:           m.Meta,
		BackendHeaders: m.BackendHeaders,
		src:            m.src,
		body:           m.body,
	}
}

// Pong creates Pong for corresponding Ping
func (m *Msg) Pong() *Msg {
	return &Msg{
		Type:          Pong,
		Ts:            m.Ts,
		CorrelationID: m.CorrelationID,
	}
}

// NewAlive creates new alive type message
func NewAlive() *Msg {
	return &Msg{Type: Alive}
}

// NewPong creates new pong type message
func NewPong() *Msg {
	return &Msg{Type: Pong}
}

// NewCurrent message for the uri
func NewCurrent(uri string) *Msg {
	return &Msg{
		Type: Current,
		URI:  uri,
	}
}

// IsPing returns true is message is Ping type
func (m *Msg) IsPing() bool {
	return m.Type == Ping
}

// IsAlive returns true is message is Alive type
func (m *Msg) IsAlive() bool {
	return m.Type == Alive
}

// NewPublish creates new publish type message
// Topic and path are combined in URI: topic/path
func NewPublish(topic, path string, ts int64, updateType uint8, o interface{}) *Msg {
	uri := topic
	if path != "" {
		uri = topic + "/" + path
	}

	return &Msg{
		Type:       Publish,
		URI:        uri,
		Ts:         ts,
		UpdateType: updateType,
		topic:      topic,
		path:       path,
		src:        toBodyMarshaler(o),
	}
}

func toBodyMarshaler(o interface{}) BodyMarshaler {
	if t, ok := o.(BodyMarshaler); ok {
		return t
	}
	return JSONMarshaler(o)
}

// IsTopicClose ...
func (m *Msg) IsTopicClose() bool {
	return m.UpdateType == Close
}

// IsReplay ...
func (m *Msg) IsReplay() bool {
	return m.Replay == Replay
}

func (m *Msg) IsCurrent() bool {
	return m.Type == Current
}

func (m *Msg) IsRequest() bool {
	return m.Type == Request
}

// IsFull ...
func (m *Msg) IsFull() bool {
	return m.UpdateType == Full
}

// Topic returns topic part of the URI
func (m *Msg) Topic() string {
	if m.topic == "" {
		m.topic = m.URI
		if strings.Contains(m.URI, "/") {
			m.topic = strings.Split(m.URI, "/")[0]
		}
	}
	return m.topic
}

// Path returns path part of the URI
func (m *Msg) Path() string {
	if strings.Contains(m.URI, "/") {
		p := strings.SplitN(m.URI, "/", 2)
		if len(p) > 1 {
			return p[1]
		}
	}
	return ""
}

// AsReplay marks message as replay
func (m *Msg) AsReplay() *Msg {
	return &Msg{
		Type:       m.Type,
		URI:        m.URI,
		UpdateType: m.UpdateType,
		Replay:     Replay,
		Ts:         m.Ts,
		body:       m.body,
		src:        m.src,
	}
}

func (m *Msg) MetaResponse(newMeta map[string]string) *Msg {
	return &Msg{
		Type:          Meta,
		Meta:          newMeta,
		Ts:            m.Ts,
		CorrelationID: m.CorrelationID,
	}
}

type jsonMarshaler struct {
	o interface{}
}

func (j jsonMarshaler) MarshalJSON() ([]byte, error) {
	if t, ok := j.o.(BodyMarshaler); ok {
		return t.MarshalJSON()
	}
	return json.Marshal(j.o)
}

// JSONMarshaler converst o to something which has MarshalJSON method
func JSONMarshaler(o interface{}) *jsonMarshaler {
	return &jsonMarshaler{o: o}
}

// TS return timestamp in unix milliseconds
func TS() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
