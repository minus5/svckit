package merger

import (
	"fmt"
	"pkg/jsonu"
	"strings"

	simplejson "github.com/minus5/go-simplejson"
)

type msg struct {
	typ      string
	channel  string
	isFull   bool
	isDiff   bool
	isDel    bool
	body     []byte
	no       int64
	jsonBody *simplejson.Json
}

func newMsg(typ string, no int64, body []byte, isDel bool) *msg {
	isDiff := strings.HasSuffix(typ, "/diff")
	channel := strings.Replace(strings.TrimSuffix(strings.TrimSuffix(typ, "/full"), "/diff"), "/", ".", -1)
	return &msg{
		typ:     typ,
		channel: channel,
		isFull:  !isDiff,
		isDiff:  isDiff,
		isDel:   isDel,
		body:    body,
		no:      no,
	}
}

func (m *msg) Merge(m2 *msg) {
	m.no = m2.no
	mj, err := m.Json()
	if err != nil {
		return
	}
	m2j, err := m2.Json()
	if err != nil {
		return
	}
	m.body = nil
	m.jsonBody = jsonu.Merge(mj, m2j)
}

func (m *msg) Json() (*simplejson.Json, error) {
	if m.jsonBody == nil {
		j, err := simplejson.NewJson(m.body)
		if err != nil {
			return nil, err
		}
		m.jsonBody = j
	}
	return m.jsonBody, nil
}

func (m *msg) JsonBody() *simplejson.Json {
	b, _ := m.Json()
	return b
}

type OutMsg struct {
	Type     string
	No       int64
	body     []byte
	jsonBody *simplejson.Json
}

func (o *OutMsg) Body() []byte {
	if o.body == nil {
		o.body, _ = o.jsonBody.Encode()
	}
	return o.body
}

func (o *OutMsg) Empty() bool {
	if o.body == nil {
		return len(o.jsonBody.MustMap()) == 0
	}
	return len(o.body) == 0
}

func (o *OutMsg) Json() *simplejson.Json {
	return o.jsonBody
}

func (o *OutMsg) Filename() string {
	t := strings.Replace(o.Type, "/", "_", -1)
	return fmt.Sprintf("%s_%d.json", t, o.No)
}
