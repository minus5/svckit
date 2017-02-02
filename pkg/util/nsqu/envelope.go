package nsqu

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

var (
	headerSeparator = []byte{10} //new line
)

type Envelope struct {
	Type          string `json:"t,omitempty"`
	ReplyTo       string `json:"r,omitempty"`
	CorrelationId string `json:"c,omitempty"`
	ExpiresAt     int64  `json:"e,omitempty"`
	Body          []byte `json:"-"`
}

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

func (m *Envelope) Bytes() []byte {
	buf, _ := json.Marshal(m)
	buf = append(buf, headerSeparator...)
	buf = append(buf, m.Body...)
	return buf
}

func (m *Envelope) ParseBody(o interface{}) error {
	if err := json.Unmarshal(m.Body, o); err != nil {
		return err
	}
	return nil
}

func (m *Envelope) Reply(o interface{}) (*Envelope, error) {
	e := &Envelope{
		Type:          strings.Replace(m.Type, ".req", ".rsp", 1),
		CorrelationId: m.CorrelationId,
	}
	if o != nil {
		buf, err := json.Marshal(o)
		if err != nil {
			return nil, err
		}
		e.Body = buf
	}
	return e, nil
}

func (m *Envelope) Expired() bool {
	if m.ExpiresAt <= 0 {
		return false
	}
	return time.Now().Unix() > m.ExpiresAt
}
