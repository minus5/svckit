package msgs

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseMessageStr(str string) (*Backend, error) {
	return parseAsBackend([]byte(str))
}

func TestMsgListic(t *testing.T) {
	m := `{"doc_type":"listici","igrac_id":"235436360ef4e64add59b56894a488be0f89ff57","doc_id":"F405BEBC-A1CA-4AA4-90C8-331D52AF7904","doc_action":"upd","no":254,"encoding":"","content_type":"json"}
{"_id":"F405BEBC-A1CA-4AA4-90C8-331D52AF7904","listic_id":50000693,"igrac_id":"235436360ef4e64add59b56894a488be0f89ff57","tip":0,"vrijeme":"10.09.2013 16:32","broj":"999-00000306","kontrolni_broj":"HGIVG","ulog":4.76,"mt":0.24,"tecaj":27.6,"eventualni_dobitak":131.38,"status":0,"dobitak":0.0,"loto_brojevi":null,"loto_kombinacije":null,"status_updated_at":null,"ts":611316,"vrsta_uplate":"internet","poredak":"1","tecajevi":[{"tecaj_id":289,"poredak":1,"broj":"239","naziv":"Gabon-Tunis","tip":"X2","vrijeme":"10.09. 17:00","tecaj":1.2,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2473689},{"tecaj_id":296,"poredak":2,"broj":"240","naziv":"Kamerun-Zambija","tip":"X2","vrijeme":"10.09. 19:30","tecaj":2.5,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2473690},{"tecaj_id":53,"poredak":3,"broj":"571","naziv":"SSC Napoli-USC Palermo","tip":"X2","vrijeme":"10.09. 20:45","tecaj":1.6,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2287242},{"tecaj_id":141,"poredak":4,"broj":"1548","naziv":"AnBiella-Air Avellino","tip":"2","vrijeme":"10.09. 18:15","tecaj":2.3,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2313165},{"tecaj_id":146,"poredak":5,"broj":"1549","naziv":"Virtus Bologna-Virtus Roma","tip":"2","vrijeme":"10.09. 18:15","tecaj":2.5,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2313166}]}`
	msg, err := parseMessageStr(m)
	assert.NotNil(t, msg)
	assert.Nil(t, err)
	assert.Equal(t, "listici", msg.Type)
	assert.Equal(t, "235436360ef4e64add59b56894a488be0f89ff57", msg.IgracId)
	assert.Equal(t, "F405BEBC-A1CA-4AA4-90C8-331D52AF7904", msg.Id)
	assert.Equal(t, 254, msg.No)
	assert.True(t, strings.HasPrefix(msg.bodyStr(), `{"_id":"F405BEBC-A1CA-4AA4-90C8-331D52AF7904","listic_id":50000693,"igrac_id":`))
}

func TestMsgListicDel(t *testing.T) {
	msg, err := parseMessageStr(`{"doc_type":"listici","igrac_id":"235436360ef4e64add59b56894a488be0f89ff57","doc_id":"7F6C3FBA-0013-4387-8DED-218426E86ED0","doc_action":"del","msg_no":262,"encoding":""}`)
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.True(t, msg.IsDel)
}

func TestMsgDel(t *testing.T) {
	msg, err := parseMessageStr(`{"type":"listici","igrac_id":"235436360ef4e64add59b56894a488be0f89ff57","id":"7F6C3FBA-0013-4387-8DED-218426E86ED0","action":"del","msg_no":262,"encoding":""}`)
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.True(t, msg.IsDel)
	assert.Equal(t, "listici", msg.Type)
	assert.Equal(t, "7F6C3FBA-0013-4387-8DED-218426E86ED0", msg.Id)
}

func TestTecajnaManifest(t *testing.T) {
	m := `{"doc_type":"tecajna/manifest","doc_id":"","igrac_id":"*","encoding":"","source":"ponuda","action":"","content_type":"string"}
["152_1378994257","151_1378994233","150_1378994209","148_1378994195","147_1378994171"]`
	msg, err := parseMessageStr(m)
	assert.NotNil(t, msg)
	assert.Nil(t, err)
	assert.Equal(t, "tecajna/manifest", msg.Type)
	assert.Equal(t, "*", msg.IgracId)
	assert.Equal(t, "", msg.Id)
	assert.Equal(t, -1, msg.No)
	assert.True(t, strings.HasPrefix(msg.bodyStr(), `["152_1378994257",`))
}

func TestMsgGzip(t *testing.T) {
	content, err := ioutil.ReadFile("./fixtures/backend_gz")
	assert.Nil(t, err)
	msg, err := parseAsBackend(content)
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, "pero", msg.Type)
	assert.Equal(t, "iso medo u ducan", msg.bodyStr())
}

func TestMsgLiveDogadjaj(t *testing.T) {
	content, err := ioutil.ReadFile("./fixtures/live_dogadjaj_web2_6430239_full")
	assert.Nil(t, err)
	msg, err := parseAsBackend(content)
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, "live/dogadjaj_web2_6430239/full", msg.Type)
	assert.Equal(t, 311, msg.No)
	assert.Equal(t, false, msg.IsDel)
	assert.Equal(t, true, msg.Gzip)
}

func TestIsFullIsDiff(t *testing.T) {
	assert.True(t, (&Backend{Type: "pero/zdero/full"}).IsFull())
	assert.True(t, (&Backend{Type: "pero/zdero/nesto"}).IsFull())

	assert.True(t, (&Backend{Type: "pero/zdero/diff"}).IsDiff())
	assert.True(t, (&Backend{Type: "pero/zdero/nesto"}).IsDiff())

	assert.Equal(t, "pero.zdero", (&Backend{Type: "pero/zdero/diff"}).RootType())
	assert.Equal(t, "pero.zdero", (&Backend{Type: "pero/zdero/full"}).RootType())
	assert.Equal(t, "pero.zdero.nesto", (&Backend{Type: "pero/zdero/nesto"}).RootType())

	assert.Equal(t, "pero_zdero_nesto_0.json", (&Backend{Type: "pero/zdero/nesto"}).FileName())
	assert.Equal(t, "pero_zdero_nesto.json", (&Backend{Type: "pero/zdero/nesto", No: -1}).FileName())
}

func TestCreateBackend(t *testing.T) {
	buf := createBackend("pero", 12, 14, nil, true)
	assert.NotNil(t, buf)

	buf = CreateBackendNoGzip("pero", 12, nil)
	assert.NotNil(t, buf)

	buf = CreateBackend("pero", 12, nil)
	assert.NotNil(t, buf)

	buf = CreateBackendTs("pero", 12, 2, nil)
	assert.NotNil(t, buf)
}
