package msgs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"pkg/jsonu"
	"pkg/util"
	"strconv"
	"strings"
	"time"
)

var (
	HeaderSeparator   []byte = []byte{10} //new line
	IgraciTopic              = "igraci"
	PorukeTopic              = "poruke"
	TransakcijeTopic         = "transakcije"
	VideoStreamsTopic        = "video_streams"
	StatTopic                = "stat"
)

//Backend - poruka koja dolazi iz backend servisa
type Backend struct {
	Type        string                 `json:"type,omitempty"`
	Id          string                 `json:"id,omitempty"`
	IgracId     string                 `json:"igrac_id,omitempty"`
	No          int                    `json:"no,omitempty"`
	From        string                 `json:"from,omitempty"`
	To          string                 `json:"to,omitempty"`
	IsDel       bool                   `json:"is_del,omitempty"`
	Gzip        bool                   `json:"-"` //da li je body inicijalno bio gzip-an
	Ts          int                    `json:"ts,omitempty"`
	Dc          string                 `json:"dc,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Encoding    string                 `json:"encoding,omitempty"`
	MessageType string                 `json:"message_type,omitempty"`
	Body        []byte                 `json:"-"` //raspakovan body
	RawBody     []byte                 `json:"-"`
	Header      map[string]interface{} `json:"-"`
	RawHeader   []byte                 `json:"-"`
}

func NewBackendFromTopic(buf []byte, topic string) *Backend {
	if !hasHeader(buf) {
		switch topic {
		case IgraciTopic:
			//igraci su specificni jer u igrac_id imaju int, a ne string kao svi drugi
			//pa ih ovdje tretiram posebno
			return newIgraciBackend(buf)
		case PorukeTopic:
			//poruke imaju _id int
			return newPorukeBackend(buf)
		case TransakcijeTopic:
			//transakcije imaju id int
			return newTransakcijeBackend(buf)
		case VideoStreamsTopic:
			return newVideoStreams(buf)
		}
	}
	if topic == StatTopic {
		return newNonJson(buf, topic)
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
		msg.RawHeader = nil
		return msg
	}
	body := parts[1]
	msg.RawBody = body
	msg.Body, _ = util.GunzipIf(body)
	return msg
}

func parseHeader(rawHeader []byte) (*Backend, error) {
	header := struct {
		DocType     string `json:"doc_type"`
		Type        string `json:"type"`
		DocId       string `json:"doc_id"`
		Id          string `json:"id"`
		DocAction   string `json:"doc_action"`
		Action      string `json:"action"`
		IgracId     string `json:"igrac_id"`
		From        string `json:"from"`
		To          string `json:"to"`
		Ts          int    `json:"ts"`
		No          int    `json:"no"`
		MsgNo       int    `json:"msg_no"`
		Encoding    string `json:"encoding"`
		DeletedId   string `json:"_deleted_id"`
		IsDel       bool   `json:"is_del"`
		Id2         string `json:"_id"`
		Dc          string `json:"dc"`
		Version     string `json:"version"`
		MessageType string `json:"message_type,omitempty"`
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
	if header.MsgNo != 0 && header.No == -1 {
		header.No = header.MsgNo
	}
	if header.To != "" && header.No == -1 { //probaj convertati to u no
		if no, err := strconv.ParseInt(header.To, 10, 32); err == nil {
			header.No = int(no)
		}
	}

	return &Backend{
		Type:        header.Type,
		IgracId:     header.IgracId,
		Id:          header.Id,
		No:          header.No,
		IsDel:       header.DocAction == "del" || header.Action == "del" || header.IsDel,
		From:        header.From,
		To:          header.To,
		Gzip:        header.Encoding == "gzip",
		Ts:          header.Ts,
		RawHeader:   rawHeader,
		Dc:          header.Dc,
		Version:     header.Version,
		Encoding:    header.Encoding,
		MessageType: header.MessageType,
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

func (m *Backend) Pack() []byte {
	//igracid i no imaju defaulte koji se ne serijaliziraju lijepo uz ommitempty, pa malo kemijam oko toga
	//volio bi neko inteligentnije rjesenje
	igracId := m.IgracId
	no := m.No
	if m.IgracId == "*" {
		m.IgracId = ""
	}
	if m.No == -1 {
		m.No = 0
	}
	var err error
	m.RawHeader, err = json.Marshal(m)
	if err != nil {
		log.Printf("[ERROR] %s", err)
	}
	m.IgracId = igracId
	m.No = no
	buf := append([]byte{}, m.RawHeader...)
	buf = append(buf, HeaderSeparator...)
	return append(buf, m.RawBody...)
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
		Type:    IgraciTopic,
		Id:      id,
		IgracId: id,
		IsDel:   msg.DeletedId != "",
		Body:    buf,
		RawBody: buf,
	}
}

func newTransakcijeBackend(buf []byte) *Backend {
	var msg struct {
		Id      string `json:"_id"`
		IgracId string `json:"igrac_id"`
		DbId    int    `json:"id"`
		Ts      int    `json:"ts"`
	}
	if err := json.Unmarshal(buf, &msg); err != nil {
		log.Printf("[ERROR] unmarshal error %s %s", err, buf)
		return nil
	}
	return &Backend{
		Type:    TransakcijeTopic,
		Id:      msg.Id,
		Ts:      msg.Ts,
		IgracId: msg.IgracId,
		Body:    buf,
		RawBody: buf,
	}
}

func newPorukeBackend(buf []byte) *Backend {
	var msg struct {
		Id      int    `json:"_id"`
		IgracId string `json:"igrac_id"`
		Ts      int    `json:"ts"`
	}
	if err := json.Unmarshal(buf, &msg); err != nil {
		log.Printf("[ERROR] unmarshal error %s %s", err, buf)
		return nil
	}
	return &Backend{
		Type:    PorukeTopic,
		Id:      strconv.Itoa(msg.Id),
		Ts:      msg.Ts,
		IgracId: msg.IgracId,
		Body:    buf,
		RawBody: buf,
	}
}

func newVideoStreams(buf []byte) *Backend {
	var msg struct {
		Id string `json:"_id"`
		Ts int    `json:"ts"`
	}
	if err := json.Unmarshal(buf, &msg); err != nil {
		log.Printf("[ERROR] unmarshal error %s %s", err, buf)
		return nil
	}
	return &Backend{
		Type:    VideoStreamsTopic,
		Id:      msg.Id,
		Ts:      msg.Ts,
		Body:    buf,
		RawBody: buf,
	}
}

func newNonJson(buf []byte, typ string) *Backend {
	return &Backend{
		Type:    typ,
		Body:    buf,
		RawBody: buf,
	}
}

func (m *Backend) SetDc(dc string) bool {
	if m.Dc == "" {
		m.Dc = dc
		return true
	}
	return false
}

func (m *Backend) SameDc(dc string) bool {
	return m.Dc == dc
}

func (m *Backend) AddToHeader(key string, value interface{}) bool {
	if m.Header == nil {
		m.Header = make(map[string]interface{})
		err := json.Unmarshal(m.RawHeader, &m.Header)
		if err != nil {
			log.Printf("[ERROR] %s", err)
			return false
		}
		//kad raspakiram u map svi brojevi odu u float64
		//pa onda kada zapakujem u exp notaciji poslije pukne na slijedecem raspakiravanju u int
		toInt(m.Header, "ts")
		toInt(m.Header, "no")
	}
	if _, ok := m.Header[key]; !ok {
		m.Header[key] = value
		m.RawHeader, _ = json.Marshal(m.Header)
		return true
	}
	return false
}

func toInt(m map[string]interface{}, key string) {
	if v, ok := m[key]; ok {
		f, ok := v.(float64)
		if ok {
			m[key] = int(f)
		}
	}
}

//UnmarshalBody - json unmarshal body-ja
func (m *Backend) UnmarshalBody(v interface{}) error {
	if err := json.Unmarshal(m.Body, v); err != nil {
		log.Printf("[ERROR] %s while parsing %s", err, m.Body)
		return err
	}
	return nil
}
