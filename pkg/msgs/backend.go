package msgs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"pkg/jsonu"
	"pkg/util"
	"strings"
	"time"
)

var (
	HeaderSeparator []byte = []byte{10} //new line
)

//Backend - poruka koja dolazi iz backend servisa
type Backend struct {
	Type      string
	Id        string
	IgracId   string
	No        int64
	From      string
	To        string
	IsDel     bool
	Gzip      bool //da li je body inicijalno bio gzip-an
	Ts        int
	Body      []byte //raspakovan body
	RawBody   []byte
	RawHeader []byte
}

func NewBackendFromTopic(buf []byte, topic string) *Backend {
	if topic == "igraci" {
		//mogu doci i kao multipart i bez headera
		if hasHeader(buf) {
			parseAsBackend(buf)
		}
		//igraci su specificni jer u igrac_id imaju int, a ne string kao svi drugi
		//pa ih ovdje tretiram posebno
		return newIgraciBackend(buf)
	}
	m := parseAsBackend(buf)
	if m.Type == "" {
		m.Type = topic
	}
	switch topic {
	case "listici.novi", "listici.promjene", "listici.dopuna", "listici":
		m.Type = "listici"
	}
	return m
}

func NewBackend(buf []byte) (*Backend, error) {
	return parseAsBackend(buf), nil
}

func MustNewBackend(buf []byte) *Backend {
	m := parseAsBackend(buf)
	return m
}

// Poruka za zatvaranja kanala ima samo tip i action, nema body
func CreateBackendDel(typ string) []byte {
	header := map[string]interface{}{
		"type":   typ,
		"action": "del",
	}
	buf, _ := json.Marshal(header)
	buf = append(buf, HeaderSeparator...)
	return buf
}

func CreateBackend(typ string, no int, body []byte) []byte {
	return createBackend(typ, no, 0, body, true)
}

func CreateBackendNoGzip(typ string, no int, body []byte) []byte {
	return createBackend(typ, no, 0, body, false)
}

func createBackend(typ string, no int, ts int, body []byte, compress bool) []byte {
	header := map[string]interface{}{
		"type": typ,
	}
	if no != -1 {
		header["no"] = no
	}
	if ts != 0 {
		header["ts"] = ts
	} else {
		header["ts"] = time.Now().UnixNano()
	}
	if compress && len(body) > 1024 {
		body = util.Gzip(body)
		header["encoding"] = "gzip"
	}
	buf, _ := json.Marshal(header)
	buf = append(buf, HeaderSeparator...)
	buf = append(buf, body...)
	return buf
}

func CreateBackendTs(typ string, no int, ts int, body []byte) []byte {
	return createBackend(typ, no, ts, body, true)
}

func parseAsBackend(buf []byte) *Backend {
	parts := bytes.SplitN(buf, HeaderSeparator, 2)
	rawHeader := parts[0]
	msg, err := parseHeader(rawHeader)
	if len(parts) == 1 || err != nil {
		msg.Body, _ = util.GunzipIf(buf)
		msg.RawBody = buf
		return msg
	}
	body := parts[1]
	msg.RawBody = body
	msg.Body, _ = util.GunzipIf(body)
	return msg
}

func parseHeader(rawHeader []byte) (*Backend, error) {
	header := struct {
		DocType   string `json:"doc_type"`
		Type      string `json:"type"`
		DocId     string `json:"doc_id"`
		Id        string `json:"id"`
		DocAction string `json:"doc_action"`
		Action    string `json:"action"`
		IgracId   string `json:"igrac_id"`
		From      string `json:"from"`
		To        string `json:"to"`
		Ts        int    `json:"ts"`
		No        int64  `json:"no"`
		Encoding  string `json:"encoding"`
		DeletedId string `json:"_deleted_id"`
		Id2       string `json:"_id"`
	}{
		No:      -1,
		IgracId: "*",
	}

	err := json.Unmarshal(rawHeader, &header)
	if err != nil {
		log.Printf("[ERROR]: %s, rawHeader: %s", err, rawHeader)
		return &Backend{}, fmt.Errorf("error parsing json header: %s, %v", rawHeader, err)
	}
	if header.Id == "" && header.Id2 != "" {
		header.Id = header.Id2
	}
	if header.Type == "" && header.DocType != "" {
		header.Type = header.DocType
	}
	if header.Id == "" && header.DocId != "" {
		header.Id = header.DocId
	}
	if header.Id == "" && header.DeletedId != "" {
		header.Id = header.DeletedId
		header.Action = "del"
	}

	return &Backend{
		Type:      header.Type,
		IgracId:   header.IgracId,
		Id:        header.Id,
		No:        header.No,
		IsDel:     header.DocAction == "del" || header.Action == "del",
		From:      header.From,
		To:        header.To,
		Gzip:      header.Encoding == "gzip",
		Ts:        header.Ts,
		RawHeader: rawHeader,
	}, nil

}

func (b *Backend) bodyStr() string {
	return string(b.Body)
}

func (b *Backend) IsDiff() bool {
	return !strings.HasSuffix(b.Type, "/full")
}

func (b *Backend) IsFull() bool {
	return !strings.HasSuffix(b.Type, "/diff")
}

func (b *Backend) RootType() string {
	return strings.Replace(strings.TrimSuffix(strings.TrimSuffix(b.Type, "/full"), "/diff"), "/", ".", -1)
}

//todo - test za ovo
func (m *Backend) FileName() string {
	fn := strings.Replace(m.Type, "/", "_", -1)
	if m.No != -1 {
		fn = fmt.Sprintf("%s_%d", fn, m.No)
	} else {
		if m.From != "" || m.To != "" {
			fn = fmt.Sprintf("%s_%s-%s", fn, m.From, m.To)
		}
	}
	return fmt.Sprintf("%s.json", fn)
}

func (m *Backend) format(bufferMarshalFunc func(buf []byte) ([]byte, error), noHeader bool) io.Reader {
	var b bytes.Buffer

	body := m.Body
	if bufferMarshalFunc != nil {
		body, _ = bufferMarshalFunc([]byte(body))
	}

	if !noHeader {
		b.Write(m.RawHeader)
		b.Write(HeaderSeparator)
	}
	if len(body) > 0 {
		b.Write(body)
		if body[len(body)-1] != 10 {
			b.Write(HeaderSeparator)
		}
	}
	b.Write(HeaderSeparator)

	return &b
}

func (m *Backend) Format(prettyJson, noHeader bool) io.Reader {
	if prettyJson {
		return m.format(jsonu.MarshalPrettyBuf, noHeader)
	} else {
		return m.format(nil, noHeader)
	}
}

func (m *Backend) FormatWith(bufferMarshalFunc func(buf []byte) ([]byte, error), noHeader bool) io.Reader {
	return m.format(bufferMarshalFunc, noHeader)
}

func hasHeader(buf []byte) bool {
	return bytes.Contains(buf, HeaderSeparator)
}

func newIgraciBackend(buf []byte) *Backend {
	var msg struct {
		Id        string `json:"_id"`
		DeletedId string `json:"_deleted_id"`
		IgracId   int    `json:"igrac_id"`
	}
	if err := json.Unmarshal(buf, &msg); err != nil {
		log.Printf("[ERROR] unmarshal error %s %s", err, buf)
		return nil
	}
	id := msg.Id
	if msg.DeletedId != "" && id == "" {
		id = msg.DeletedId
	}
	if id == "" {
		log.Printf("[ERROR] ovo nije igraci poruka %s", buf)
		return nil
	}
	return &Backend{
		Type:    "igraci",
		Id:      id,
		IgracId: id,
		IsDel:   msg.DeletedId != "",
		Body:    buf,
		RawBody: buf,
	}
}
