package msgs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"pkg/jsonu"
	"pkg/util"
	"strconv"
	"strings"
	"time"

	"github.com/minus5/go-simplejson"
	"github.com/minus5/svckit/log"
)

var (
	HeaderSeparator = []byte{10} //new line
)

const (
	// GzipMsgSizeLimit poruke manje od ove ne gzipamo
	GzipMsgSizeLimit  = 32768
	IgraciTopic       = "igraci"
	PorukeTopic       = "poruke"
	TransakcijeTopic  = "transakcije"
	VideoStreamsTopic = "video_streams"
	StatsTopic        = "stats"
)

const (
	KeyId      = "id"
	KeyIgracId = "igrac_id"
	KeyFrom    = "from"
	KeyTo      = "to"
)

//Backend - poruka koja dolazi iz backend servisa
type Backend struct {
	Type        string `json:"type,omitempty"`
	Id          string `json:"id,omitempty"`
	IgracId     string `json:"igrac_id,omitempty"`
	No          int    `json:"no,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	IsDel       bool   `json:"is_del,omitempty"`
	Ts          int    `json:"ts,omitempty"`
	Dc          string `json:"dc,omitempty"`
	Version     string `json:"version,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	MessageType string `json:"message_type,omitempty"`
	Body        []byte `json:"-"` //raspakovan body
	RawBody     []byte `json:"-"`
	RawHeader   []byte `json:"-"`
	rawMsg      []byte
	jsonBody    *simplejson.Json
}

func NewBackendOrSimple(buf []byte, topic string) *Backend {
	if bytes.Contains(buf, HeaderSeparator) {
		return parseAsBackend(buf)
	}
	return &Backend{
		Type:    topic,
		Body:    buf,
		RawBody: buf,
	}
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
	if topic == StatsTopic {
		if buf[0] != '{' {
			return newNonJson(buf, topic)
		}
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
	if compress && len(body) > GzipMsgSizeLimit {
		body = util.Gzip(body)
		header["encoding"] = "gzip"
	}
	buf, _ := json.Marshal(header)
	buf = append(buf, HeaderSeparator...)
	buf = append(buf, body...)
	return buf
}

func Header(key string, value interface{}) func(map[string]interface{}) {
	return func(h map[string]interface{}) {
		h[key] = value
	}
}

var gzipKey = "__gzipKey__"

func NoGzip() func(map[string]interface{}) {
	return func(h map[string]interface{}) {
		h[gzipKey] = false
	}
}

func BackendFactory(typ string, body []byte, opts ...func(map[string]interface{})) []byte {
	header := map[string]interface{}{
		"type": typ,
	}
	for _, o := range opts {
		o(header)
	}
	compress := true
	if v, ok := header[gzipKey]; ok {
		compress = v.(bool)
		delete(header, gzipKey)
	}
	if compress && len(body) > GzipMsgSizeLimit {
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
	msg.rawMsg = buf
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

// IsDiff da li je rijec o diff poruci
func (b *Backend) IsDiff() bool {
	return !strings.HasSuffix(b.Type, "/full")
}

// IsFull da li je rijec o full poruci.
func (b *Backend) IsFull() bool {
	return !strings.HasSuffix(b.Type, "/diff")
}

func (b *Backend) RootType() string {
	return strings.Replace(strings.TrimSuffix(strings.TrimSuffix(b.Type, "/full"), "/diff"), "/", ".", -1)
}

//todo - test za ovo
func (b *Backend) FileName() string {
	fn := strings.Replace(b.Type, "/", "_", -1)
	if b.No != -1 {
		fn = fmt.Sprintf("%s_%d", fn, b.No)
	} else {
		if b.From != "" || b.To != "" {
			fn = fmt.Sprintf("%s_%s-%s", fn, b.From, b.To)
		}
	}
	return fmt.Sprintf("%s.json", fn)
}

func (b *Backend) format(bufferMarshalFunc func(buf []byte) ([]byte, error), noHeader bool) io.Reader {
	var buf bytes.Buffer

	body := b.Body
	if bufferMarshalFunc != nil {
		body, _ = bufferMarshalFunc([]byte(body))
	}

	if !noHeader {
		buf.Write(b.RawHeader)
		buf.Write(HeaderSeparator)
	}
	if len(body) > 0 {
		buf.Write(body)
		if body[len(body)-1] != 10 {
			buf.Write(HeaderSeparator)
		}
	}
	buf.Write(HeaderSeparator)

	return &buf
}

func (b *Backend) pack() []byte {
	if b.rawMsg != nil {
		// u medjuvremenu nista nije promjenjeno
		return b.rawMsg
	}
	if b.jsonBody != nil {
		b.Body, _ = b.jsonBody.Encode()
		b.RawBody = b.Body
		b.Encoding = ""
	}
	//igracid i no imaju defaulte koji se ne serijaliziraju lijepo uz ommitempty, pa malo kemijam oko toga
	//volio bi neko inteligentnije rjesenje
	igracId := b.IgracId
	no := b.No
	if b.IgracId == "*" {
		b.IgracId = ""
	}
	if b.No == -1 {
		b.No = 0
	}
	var err error
	b.RawHeader, err = json.Marshal(b)
	if err != nil {
		log.Printf("[ERROR] %s", err)
	}
	b.IgracId = igracId
	b.No = no
	buf := append([]byte{}, b.RawHeader...)
	buf = append(buf, HeaderSeparator...)
	b.rawMsg = append(buf, b.RawBody...)
	return b.rawMsg
}

func (b *Backend) RawMessage() []byte {
	return b.pack()
}

func (b *Backend) Format(prettyJson, noHeader bool) io.Reader {
	if prettyJson {
		return b.format(jsonu.MarshalPrettyBuf, noHeader)
	} else {
		return b.format(nil, noHeader)
	}
}

func (b *Backend) FormatWith(bufferMarshalFunc func(buf []byte) ([]byte, error), noHeader bool) io.Reader {
	return b.format(bufferMarshalFunc, noHeader)
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

// SetDc postavi dc ako vec nije postavljen, true ako je uspio
func (b *Backend) SetDc(dc string) bool {
	if b.Dc == "" {
		b.Dc = dc
		b.rawMsg = nil
		return true
	}
	return false
}

// SameDc jesmo li u istom dc
func (b *Backend) SameDc(dc string) bool {
	return b.Dc == dc
}

//UnmarshalBody - json unmarshal body-ja
func (b *Backend) UnmarshalBody(v interface{}) error {
	if err := json.Unmarshal(b.Body, v); err != nil {
		log.Printf("[ERROR] %s while parsing %s", err, b.Body)
		return err
	}
	return nil
}

// Json vrati body u simplejson formatu
func (b *Backend) Json() *simplejson.Json {
	if b.jsonBody == nil {
		j, err := simplejson.NewJson(b.Body)
		if err != nil {
			log.Fatal(err)
		}
		b.jsonBody = j
	}
	return b.jsonBody
}

// Merge spaja diff proruke na postojeci full.
// I tako nadogradjuje u novi full.
func (b *Backend) Merge(diff *Backend) {
	if !b.IsFullDiff() || !diff.IsFullDiff() {
		// ovo se ne bi smijelo dogoditi, greska je u logici
		log.Notice("poruka nije full/diff tipa")
		return
	}
	jsonu.Merge(b.Json(), diff.Json())
	b.Ts = diff.Ts
	b.No = diff.No
	// ovi podaci vise nemaju smisla pa ih brisem da ih ne bi greskom koristio
	b.RawBody = nil
	b.Encoding = ""
	b.RawHeader = nil
	b.Body = nil
	// rawMsg ce se ponov izgraditi u pack
	b.rawMsg = nil
	b.pack()
}

// IsFullDiff radi li se o full/diff tipu poruke
func (b *Backend) IsFullDiff() bool {
	return isFullDiff(b.Type)
}

func isFullDiff(typ string) bool {
	if typ == "tecajna/diff" {
		return false
	}
	return strings.Contains(typ, "/diff") ||
		strings.Contains(typ, "/full")
}

// IsHeartbeat radi li se o heartbeat tipu poruke
func (b *Backend) IsHeartbeat() bool {
	return b.Type == "heartbeat"
}

// MessageId is kod snimanja u message store (mongo)
func (b *Backend) MessageId() string {
	if b.Type == "tecajna/diff" {
		return fmt.Sprintf("%s-%s-%s", b.Type, b.From, b.To)
	}
	if b.Type == "tecajna/full" {
		return fmt.Sprintf("%s-%s", b.Type, b.From)
	}
	return b.Type
}

// MessageExpiresAt vrijeme kada message vise ne vazi, moze se brisati iz message store (mongo)
func (b *Backend) MessageExpiresAt() *time.Time {
	if b.Type == "tecajna/diff" || b.Type == "tecajna/full" {
		aDay := time.Now().Add(24 * time.Hour)
		return &aDay
	}
	return nil
}
