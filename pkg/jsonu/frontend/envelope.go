package frontend

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Envelope struct {
	Type      string
	Id        string
	Status    string
	BodyPath  string
	Timestamp string
	No        int64
	Header    string
	Body      []byte
}

func NewEnvelope(buf []byte) (*Envelope, error) {
	p := bytes.SplitN(buf, []byte{10}, 2)
	if len(p) < 2 {
		return nil, fmt.Errorf("can't find header in message: %s", buf)
	}
	m := &Envelope{
		Header: string(p[0]),
		Body:   p[1],
	}
	h := strings.Split(m.Header, ":")
	if len(h) < 6 {
		return nil, fmt.Errorf("could not parse header: %s", m.Header)
	}
	m.Type = h[0]
	m.Id = h[1]
	m.Status = h[2]
	m.BodyPath = h[3]
	m.Timestamp = h[4]
	if h[5] != "" {
		if i, err := strconv.Atoi(h[5]); err == nil {
			m.No = int64(i)
		}
	}
	return m, nil
}

func (e *Envelope) Filename() string {
	t := strings.Replace(e.Type, "/", "_", -1)
	if e.No == 0 {
		return fmt.Sprintf("%s_%s.json", t, e.No)
	}
	return fmt.Sprintf("%s.json", t)
}

func (e *Envelope) IsFullDiff() bool {
	if e.Type == "tecajna/diff" {
		return false
	}
	return strings.Contains(e.Type, "/diff") ||
		strings.Contains(e.Type, "/full")
}
