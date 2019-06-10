package amp

import (
	"encoding/json"
	"strings"

	"github.com/minus5/svckit/log"
)

func ParseCompatibility(buf []byte, version uint8) *Msg {
	if version == CompatibilityVersion1 {
		return ParseV1(buf)
	}
	return Parse(buf)
}

func ParseV1(buf []byte) *Msg {
	if len(buf) == 0 {
		return nil
	}
	v1 := struct {
		Type          uint8 `json:"t,omitempty"`
		Subscriptions []struct {
			Stream string `json:"s,omitempty"`
			No     int64  `json:"n,omitempty"`
		} `json:"u,omitempty"`
	}{}
	if err := json.Unmarshal(buf, &v1); err != nil {
		log.S("header", string(buf)).Error(err)
		return nil
	}
	if v1.Type == Ping {
		return &Msg{Type: Ping}
	}
	if v1.Type != Subscribe {
		log.S("header", string(buf)).ErrorS("unknown message type")
		return nil
	}
	v2 := &Msg{
		Type:          Subscribe,
		Subscriptions: make(map[string]int64),
	}
	for _, s := range v1.Subscriptions {
		if s.Stream == "" {
			continue
		}
		v2.Subscriptions["sportsbook/"+s.Stream] = s.No
	}
	return v2
}

func ParseV1Subscriptions(buf []byte) *Msg {
	v1s := []struct {
		Stream string `json:"s,omitempty"`
		No     int64  `json:"n,omitempty"`
	}{}
	if err := json.Unmarshal(buf, &v1s); err != nil {
		log.S("header", string(buf)).Error(err)
		return nil
	}
	v2 := &Msg{
		Type:          Subscribe,
		Subscriptions: make(map[string]int64),
	}
	for _, s := range v1s {
		if s.Stream == "" || strings.Contains(s.Stream, "_NaN") {
			continue
		}
		v2.Subscriptions["sportsbook/"+s.Stream] = s.No
	}
	return v2
}

// Marshal packs message for sending on the wire
func (m *Msg) MarshalV1() []byte {
	buf, _ := m.marshal(CompressionNone, CompatibilityVersion1)
	return buf
}

// MarshalDeflate packs and compress message
func (m *Msg) MarshalV1Deflate() ([]byte, bool) {
	return m.marshal(CompressionDeflate, CompatibilityVersion1)
}

func (m *Msg) marshalV1header() []byte {
	v1 := struct {
		Type   uint8  `json:"t,omitempty"`
		Stream string `json:"s,omitempty"`
		No     int64  `json:"n,omitempty"`
		Full   uint8  `json:"f,omitempty"`
	}{
		Type:   m.Type,
		Stream: strings.TrimPrefix(m.URI, "sportsbook/"),
		No:     m.Ts,
	}
	if m.UpdateType == Full {
		v1.Full = 1
	}
	header, _ := json.Marshal(v1)
	return header
}

func (m *Msg) MarshalCompatiblity(version uint8) []byte {
	if version == CompatibilityVersion1 {
		return m.MarshalV1()
	}
	return m.Marshal()
}

func (m *Msg) MarshalDeflateCompatiblity(version uint8) ([]byte, bool) {
	if version == CompatibilityVersion1 {
		return m.MarshalV1Deflate()
	}
	return m.MarshalDeflate()
}
