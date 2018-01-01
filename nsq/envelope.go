package nsq

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

var (
	headerSeparator = []byte{10} //new line
)

// Envelope arround message for request response communication over nsq.
type Envelope struct {
	// type of the message
	Type string `json:"t,omitempty"`
	// nsq topic to send reply to
	ReplyTo string `json:"r,omitempty"`
	// connection between request and response
	CorrelationId string `json:"c,omitempty"`
	// unix timestamp when message expires, after that should be dropped
	ExpiresAt int64  `json:"e,omitempty"`
	Error     string `json:"error,omitempty"`
	// message body
	Body []byte `json:"-"`
}

// NewEnvelope decodes envelope from bytes.
func NewEnvelope(buf []byte) (*Envelope, error) {
	parts := bytes.SplitN(buf, headerSeparator, 2)
	e := &Envelope{}
	if err := json.Unmarshal(parts[0], e); err != nil {
		return nil, err
	}
	if len(parts) > 1 {
		e.Body = parts[1]
	}
	return e, nil
}

// Bytes encodes envelope into bytes for putting on wire.
func (m *Envelope) Bytes() []byte {
	buf, _ := json.Marshal(m)
	buf = append(buf, headerSeparator...)
	buf = append(buf, m.Body...)
	return buf
}

// ParseBody deocdes Evelope body into o.
func (m *Envelope) ParseBody(o interface{}) error {
	if err := json.Unmarshal(m.Body, o); err != nil {
		return err
	}
	return nil
}

// Reply creates reply Envelope from request Envelope.
func (m *Envelope) Reply(o interface{}, err error) (*Envelope, error) {
	e := &Envelope{
		Type:          strings.Replace(m.Type, ".req", ".rsp", 1),
		CorrelationId: m.CorrelationId,
	}
	if err != nil {
		e.Error = err.Error()
	}
	if o != nil {
		if buf, ok := o.([]byte); ok {
			e.Body = buf
		} else {
			buf, err := json.Marshal(o)
			if err != nil {
				return nil, err
			}
			e.Body = buf
		}
	}
	return e, nil
}

// Expired returns true if message expired.
func (m *Envelope) Expired() bool {
	if m.ExpiresAt <= 0 {
		return false
	}
	return time.Now().Unix() > m.ExpiresAt
}
