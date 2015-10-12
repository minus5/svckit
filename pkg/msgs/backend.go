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

func NewBackend(buf []byte) (*Backend, error) {
	return parseAsBackend(buf)
}

func MustNewBackend(buf []byte) *Backend {
	m, err := parseAsBackend(buf)
	if err != nil {
		log.Fatal(err)
	}
	return m
}

// Poruka za zatvaranja kanala ima samo tip i action, nema body
func CreateBackendDel(typ string) []byte {
	header := map[string]interface{}{
		"type":   typ,
		"action": "del",
	}
	buf, _ := json.Marshal(header)
	buf = append(buf, []byte{10}...)
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
	} /*else {
		header["ts"] = time.Now().UnixNano
	}*/
	if compress && len(body) > 1024 {
		body = util.Gzip(body)
		header["encoding"] = "gzip"
	}
	buf, _ := json.Marshal(header)
	buf = append(buf, []byte{10}...)
	buf = append(buf, body...)
	return buf
}

func CreateBackendTs(typ string, no int, ts int, body []byte) []byte {
	return createBackend(typ, no, ts, body, true)
}

func parseAsBackend(buf []byte) (*Backend, error) {
	parts := bytes.SplitN(buf, []byte{10}, 2)
	rawHeader := parts[0]

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
	}{
		No:      -1,
		IgracId: "*",
	}

	err := json.Unmarshal(rawHeader, &header)
	if err != nil {
		return nil, fmt.Errorf("error parsing json header: %s", rawHeader)
	}
	if header.Type == "" && header.DocType != "" {
		header.Type = header.DocType
	}
	if header.Id == "" && header.DocId != "" {
		header.Id = header.DocId
	}

	msg := &Backend{
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
	}
	if len(parts) > 1 {
		body := parts[1]
		msg.RawBody = body
		if msg.Gzip && util.IsGziped(body) {
			if msg.Body, err = util.Gunzip(body); err != nil {
				return nil, err
			}
		} else {
			msg.Body = body
		}
	} else {
		msg.Body, _ = util.GunzipIf(buf)
		msg.RawBody = buf
	}

	return msg, nil
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
		b.Write([]byte{10})
	}
	if len(body) > 0 {
		b.Write(body)
		if body[len(body)-1] != 10 {
			b.Write([]byte{10})
		}
	}
	b.Write([]byte{10})

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
