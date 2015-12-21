package msgs

import (
	"io/ioutil"
	"pkg/common"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseMessageStr(str string) *Backend {
	return parseAsBackend([]byte(str))
}

func TestMsgListic(t *testing.T) {
	m := `{"doc_type":"listici","igrac_id":"235436360ef4e64add59b56894a488be0f89ff57","doc_id":"F405BEBC-A1CA-4AA4-90C8-331D52AF7904","doc_action":"upd","no":254,"encoding":"","content_type":"json"}
{"_id":"F405BEBC-A1CA-4AA4-90C8-331D52AF7904","listic_id":50000693,"igrac_id":"235436360ef4e64add59b56894a488be0f89ff57","tip":0,"vrijeme":"10.09.2013 16:32","broj":"999-00000306","kontrolni_broj":"HGIVG","ulog":4.76,"mt":0.24,"tecaj":27.6,"eventualni_dobitak":131.38,"status":0,"dobitak":0.0,"loto_brojevi":null,"loto_kombinacije":null,"status_updated_at":null,"ts":611316,"vrsta_uplate":"internet","poredak":"1","tecajevi":[{"tecaj_id":289,"poredak":1,"broj":"239","naziv":"Gabon-Tunis","tip":"X2","vrijeme":"10.09. 17:00","tecaj":1.2,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2473689},{"tecaj_id":296,"poredak":2,"broj":"240","naziv":"Kamerun-Zambija","tip":"X2","vrijeme":"10.09. 19:30","tecaj":2.5,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2473690},{"tecaj_id":53,"poredak":3,"broj":"571","naziv":"SSC Napoli-USC Palermo","tip":"X2","vrijeme":"10.09. 20:45","tecaj":1.6,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2287242},{"tecaj_id":141,"poredak":4,"broj":"1548","naziv":"AnBiella-Air Avellino","tip":"2","vrijeme":"10.09. 18:15","tecaj":2.3,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2313165},{"tecaj_id":146,"poredak":5,"broj":"1549","naziv":"Virtus Bologna-Virtus Roma","tip":"2","vrijeme":"10.09. 18:15","tecaj":2.5,"status":0,"rezultat":null,"dobitni_tip":null,"is_fix":false,"ponuda_id":2313166}]}`
	msg := parseMessageStr(m)
	assert.NotNil(t, msg)
	assert.Equal(t, "listici", msg.Type)
	assert.Equal(t, "235436360ef4e64add59b56894a488be0f89ff57", msg.IgracId)
	assert.Equal(t, "F405BEBC-A1CA-4AA4-90C8-331D52AF7904", msg.Id)
	assert.Equal(t, 254, msg.No)
	assert.True(t, strings.HasPrefix(msg.bodyStr(), `{"_id":"F405BEBC-A1CA-4AA4-90C8-331D52AF7904","listic_id":50000693,"igrac_id":`))
}

func TestMsgListicDel(t *testing.T) {
	msg := parseMessageStr(`{"doc_type":"listici","igrac_id":"235436360ef4e64add59b56894a488be0f89ff57","doc_id":"7F6C3FBA-0013-4387-8DED-218426E86ED0","doc_action":"del","msg_no":262,"encoding":""}`)
	assert.NotNil(t, msg)
	assert.True(t, msg.IsDel)
}

func TestMsgDel(t *testing.T) {
	msg := parseMessageStr(`{"type":"listici","igrac_id":"235436360ef4e64add59b56894a488be0f89ff57","id":"7F6C3FBA-0013-4387-8DED-218426E86ED0","action":"del","msg_no":262,"encoding":""}`)
	assert.NotNil(t, msg)
	assert.True(t, msg.IsDel)
	assert.Equal(t, "listici", msg.Type)
	assert.Equal(t, "7F6C3FBA-0013-4387-8DED-218426E86ED0", msg.Id)
}

func TestTecajnaManifest(t *testing.T) {
	m := `{"doc_type":"tecajna/manifest","doc_id":"","igrac_id":"*","encoding":"","source":"ponuda","action":"","content_type":"string"}
["152_1378994257","151_1378994233","150_1378994209","148_1378994195","147_1378994171"]`
	msg := parseMessageStr(m)
	assert.NotNil(t, msg)
	assert.Equal(t, "tecajna/manifest", msg.Type)
	assert.Equal(t, "*", msg.IgracId)
	assert.Equal(t, "", msg.Id)
	assert.Equal(t, -1, msg.No)
	assert.True(t, strings.HasPrefix(msg.bodyStr(), `["152_1378994257",`))
}

func TestMsgGzip(t *testing.T) {
	content, err := ioutil.ReadFile("./fixtures/backend_gz")
	assert.Nil(t, err)
	msg := parseAsBackend(content)
	assert.NotNil(t, msg)
	assert.Equal(t, "pero", msg.Type)
	assert.Equal(t, "iso medo u ducan", msg.bodyStr())
}

func TestMsgLiveDogadjaj(t *testing.T) {
	content, err := ioutil.ReadFile("./fixtures/live_dogadjaj_web2_6430239_full")
	assert.Nil(t, err)
	msg := parseAsBackend(content)
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

func TestListiciDel(t *testing.T) {
	m := `{"_deleted_id":"B481B078-0169-427C-8038-A65DC195E08F","igrac_id":"781f220e1dbfc1e1493d57caded83559b4ac6b57"}`
	msg := parseMessageStr(m)
	assert.NotNil(t, msg)
	assert.Equal(t, "", msg.Type)
	assert.Equal(t, "781f220e1dbfc1e1493d57caded83559b4ac6b57", msg.IgracId)
	assert.Equal(t, "B481B078-0169-427C-8038-A65DC195E08F", msg.Id)
	assert.True(t, msg.IsDel)

	msg = NewBackendFromTopic([]byte(m), "listici.novi")
	assert.Equal(t, "listici", msg.Type)
	assert.Equal(t, "781f220e1dbfc1e1493d57caded83559b4ac6b57", msg.IgracId)
	assert.Equal(t, "B481B078-0169-427C-8038-A65DC195E08F", msg.Id)
	assert.True(t, msg.IsDel)
}

var igraciMessage string = `{"_id":"69fe88b8105f62b8622dec7b4cab34c6cf673e2e","adresa":"kresimirova 54","beta_programi":["konji2web"],"datum_rodjenja":"1953-01-15T00:00:00+01:00","dokumenata":{"internet":1,"listica":1,"poruka":0,"poslovnica":0,"transakcija":1609,"www":0},"drzava":"Hrvatska","email":"brunokapor@gmail.com","filename":null,"grad":"Rijeka","identitet_potvrdjen":false,"igrac_id":207879,"ime":"bruno","nadimak":"setter","neodigrani_iznos":0,"neodigrano":0,"original_filename":null,"pending_email":null,"postanski_broj":"51000","poziv_na_broj":"0839101776","prezime":"kapor","racuni":[],"raspolozivo":58,"state":"internet_active","stranac_potvrdjen":false,"telefon":"214655","tip":null,"ts":6304258350,"zadnja_isplata_broj_racuna":null,"zadnja_isplata_poslovnica_id":null,"zadnja_isplata_tip":null,"zadnja_procitana_poruka":16122928}`

// func TestIgraciParse(t *testing.T) {
// 	msg := parseMessageStr(igraciMessage)
// 	assert.Equal(t, "", msg.Type)
// 	assert.Equal(t, "", msg.IgracId)
// 	assert.Equal(t, "", msg.Id)
// }

func TestIgraciNonBackend(t *testing.T) {
	msg := NewBackendFromTopic([]byte(igraciMessage), IgraciTopic)
	assert.Equal(t, IgraciTopic, msg.Type)
	assert.Equal(t, "69fe88b8105f62b8622dec7b4cab34c6cf673e2e", msg.IgracId)
	assert.Equal(t, "69fe88b8105f62b8622dec7b4cab34c6cf673e2e", msg.Id)
	assert.False(t, msg.IsDel)
	assert.Equal(t, string(msg.Body), igraciMessage)
}

func TestIgraciBackend(t *testing.T) {
	header := `{"type":"igraci","id":"neki","igrac_id":"drugi"}`
	m := []byte(header + "\n" + igraciMessage)
	assert.True(t, hasHeader(m))
	msg := NewBackendFromTopic(m, IgraciTopic)
	assert.Equal(t, IgraciTopic, msg.Type)
	assert.Equal(t, "drugi", msg.IgracId)
	assert.Equal(t, "neki", msg.Id)
	assert.False(t, msg.IsDel)
	assert.Equal(t, string(msg.Body), igraciMessage)
}

func TestAppVersion(t *testing.T) {
	m := `{"_id":"plazma","app":"plazma","version":"3.0.10"}`
	msg := parseMessageStr(m)
	assert.NotNil(t, msg)
	assert.Equal(t, "", msg.Type)
	assert.Equal(t, "*", msg.IgracId)
	assert.Equal(t, "plazma", msg.Id)
	assert.False(t, msg.IsDel)

	msg = NewBackendFromTopic([]byte(m), "app_version")
	assert.Equal(t, "app_version", msg.Type)
	assert.Equal(t, "*", msg.IgracId)
	assert.Equal(t, "plazma", msg.Id)
	assert.False(t, msg.IsDel)
}

var listiciMessage string = `{"_id":"00395309-4D54-4A4F-9975-4D7E6EC35C1E","broj":"999-01743696","dobitak":0,"eventualni_dobitak":2308.69,"igrac_id":"81cf2cc561fff9099192f01ece5be4d149772eab","kontrolni_broj":"RV1ZZ","listic_id":865670203,"loto":{"broj":"G2344","broj_izvlacenih":20,"broj_kuglica":80,"dobitni_brojevi":[],"kolo":532344,"loto_tip_id":100,"naziv":"GRČKI KINO LOTO 20/80","tecajevi":[3.75,15,65,275,1350,6500,25000,125000],"vrijeme":"16.12.2015 11:35"},"loto_brojevi":[16,47,57,61,79],"loto_kombinacije":[5],"mt":0.1,"novi":1,"poredak":"1","porez":256.31,"porez_tip":2,"status":0,"status_updated_at":null,"tecaj":1350,"tip":2,"ts":6304298028,"ulog":1.9,"vrijeme":"16.12.2015 11:07","vrsta_uplate":"internet"}`

func TestListiciNonBackend(t *testing.T) {
	msg := NewBackendFromTopic([]byte(listiciMessage), "listici.dopuna")
	assert.Equal(t, "listici", msg.Type)
	assert.Equal(t, "81cf2cc561fff9099192f01ece5be4d149772eab", msg.IgracId)
	assert.Equal(t, "00395309-4D54-4A4F-9975-4D7E6EC35C1E", msg.Id)
	assert.False(t, msg.IsDel)
	assert.Equal(t, string(msg.Body), listiciMessage)
}

func TestListiciBackend(t *testing.T) {
	header := `{"type":"listici","id":"neki","igrac_id":"drugi"}`
	msg := NewBackendFromTopic([]byte(header+"\n"+listiciMessage), "listici.dopuna")
	assert.Equal(t, "listici", msg.Type)
	assert.Equal(t, "drugi", msg.IgracId)
	assert.Equal(t, "neki", msg.Id)
	assert.False(t, msg.IsDel)
	assert.Equal(t, string(msg.Body), listiciMessage)

	//msg.SetDc("ec2")
	//t.Logf("packed: %s", msg.Pack())
}

func TestSetDc(t *testing.T) {
	header := []byte(`{"dc":"ec2"}`)
	msg := NewBackendFromTopic(header, "topic")
	assert.Equal(t, "topic", msg.Type)
	assert.Equal(t, "ec2", msg.Dc)
	assert.True(t, msg.SameDc("ec2"))
	assert.False(t, msg.SetDc("pero"))
}

func TestParsePackHeaders(t *testing.T) {
	before := []string{
		`{"encoding":"gzip","no":3720164,"ts":1450259323479380960,"type":"dogadjaj"}`,
		`{"doc_type":"live/simple/diff","action":"del","from":"15938004","to":"15938005","no":15938005,"ts":1450259324462,"encoding":"gzip"}`,
		`{"igrac_id":"*","no":281,"created_at":1450259340,"doc_type":"konji20/kraj_oklada","encoding":"gzip"}`,
		`{"version":"344705_1450259522","message_type":"insert/delete","encoding":"gzip"}`,
		`{"igrac_id":"*","msg_no":1466,"created_at":1450259536,"from":"344705_1450259091","to":"344705_1450259522","doc_type":"tecajna/diff"}`,
	}

	after := []string{
		`{"type":"dogadjaj","no":3720164,"ts":1450259323479380960,"dc":"ec2","encoding":"gzip"}`,
		`{"type":"live/simple/diff","no":15938005,"from":"15938004","to":"15938005","is_del":true,"ts":1450259324462,"dc":"ec2","encoding":"gzip"}`,
		`{"type":"konji20/kraj_oklada","no":281,"dc":"ec2","encoding":"gzip"}`,
		`{"dc":"ec2","version":"344705_1450259522","encoding":"gzip","message_type":"insert/delete"}`,
		`{"type":"tecajna/diff","no":1466,"from":"344705_1450259091","to":"344705_1450259522","dc":"ec2"}`,
	}

	for i, header := range before {
		m, err := parseHeader([]byte(header))
		assert.Nil(t, err)
		assert.True(t, m.SetDc("ec2"))
		assert.Equal(t, string(m.Pack()), after[i]+"\n")
		//t.Logf("%s", m.Pack())
	}
}

func TestParseTransakcije(t *testing.T) {
	buf := []byte(`{"_id":"27302A97-8B21-407E-A68F-22D89220D83B","broj_listica":"999-02518899","created_at":"2015-12-21T11:14:06.08+01:00","id":248019016,"igrac_id":"1a3ee64c0003c7bdedc192231236ddc1763c68a3","iznos":-3,"raspolozivo":31,"tip":"uplata listića","ts":6324292907}`)
	m := NewBackendFromTopic(buf, TransakcijeTopic)
	assert.NotNil(t, m)
	assert.Equal(t, "27302A97-8B21-407E-A68F-22D89220D83B", m.Id)
	assert.Equal(t, 6324292907, m.Ts)
	assert.Equal(t, "1a3ee64c0003c7bdedc192231236ddc1763c68a3", m.IgracId)
	assert.Equal(t, buf, m.Body)
	assert.Equal(t, buf, m.RawBody)
	//t.Logf("%s", m.Pack())
}

func TestParsePoruke(t *testing.T) {
	buf := []byte(`{"_id":16667654,"created_at":"2015-12-21T11:04:03.886666666+01:00","igrac_id":"4e1f30b96807887d9734d103396c2bbbedc99e56","text":"        \u003cp\u003e \n          Izvršena je uplata na Vaš račun u iznosu od: 98,37 Kn.\n          \u003cbr/\u003e\n          Trenutno stanje Vašeg računa je 101,93 Kn.\n        \u003c/p\u003e\n","ts":6324281979}`)
	m := NewBackendFromTopic(buf, PorukeTopic)
	assert.NotNil(t, m)
	assert.Equal(t, "16667654", m.Id)
	assert.Equal(t, 6324281979, m.Ts)
	assert.Equal(t, "4e1f30b96807887d9734d103396c2bbbedc99e56", m.IgracId)
	assert.Equal(t, buf, m.Body)
	assert.Equal(t, buf, m.RawBody)
	//t.Logf("%s", m.Pack())
}

func TestListiciBrisi(t *testing.T) {
	buf := []byte(`{"igrac_id":"0ceb682759f558654ae308a3fd0d8d307ac14f6a","listici":["42D5506D-1751-4563-A996-7AFEDD26DC67"]}`)
	m := NewBackendFromTopic(buf, "listici.brisi")
	assert.NotNil(t, m)
	assert.Equal(t, "0ceb682759f558654ae308a3fd0d8d307ac14f6a", m.IgracId)
	assert.Equal(t, buf, m.Body)
	assert.Equal(t, buf, m.RawBody)
	//t.Logf("%s", m.Pack())

	var lb ListiciBrisiMessage
	err := m.UnmarshalBody(&lb)
	assert.Nil(t, err)
	assert.Equal(t, "0ceb682759f558654ae308a3fd0d8d307ac14f6a", lb.IgracId)
	assert.Equal(t, 1, len(lb.Listici))
}

func TestPushNotSubscribe(t *testing.T) {
	buf := []byte(`{"igrac_id":"707e46ff7fdf7c3815210a91cc9dce35f077acab","uredjaj":{"gcm_id":"APA91bGrkydnlbNUNm15S5pGcfFNmvIa5OfuVeJW-iQvXAewqzWEuxXqNRsG0Ue3Yk7v9nl4m80_Af7i45x0JqIbk_QXFxrwe_CTdkjaSdeaG5VCqNljgbA","apple_id":"","aktivan":true},"pretplate":{"privatne_poruke":true,"novosti":true,"listic_vrednovan":true}}`)
	m := NewBackendFromTopic(buf, "push_not.subscribe")
	assert.NotNil(t, m)
	assert.Equal(t, "707e46ff7fdf7c3815210a91cc9dce35f077acab", m.IgracId)
	assert.Equal(t, buf, m.Body)
	assert.Equal(t, buf, m.RawBody)
	//t.Logf("%s", m.Pack())

	var p PushNotSubscribe
	err := m.UnmarshalBody(&p)
	assert.Nil(t, err)
	assert.Equal(t, "707e46ff7fdf7c3815210a91cc9dce35f077acab", p.IgracId)
	assert.Regexp(t, "^APA91bGrkydnlbNUNm15S5pGcfFNmvIa5OfuVeJW", p.Uredjaj.GcmId)
	assert.True(t, p.Pretplate.Novosti)
}

func TestLogWebAppApi(t *testing.T) {
	buf := []byte(`{"session_id":"3C95C664-6163-4B74-7586-A3162A79CF7D-2bf2","igrac_id":"","source":"kladomat_plazma","level":"plazma","message":"{\"tag\": \"kladomat_plazma\", \"version\": \"2.0.3\", \"plazmaId\": \"a8ee1ad4\", \"sinceStart\": 59, \"sinceLast\": 59, \"lastData\": 0, \"ua\": \"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/536.8 (KHTML, like Gecko) Chrome/20.0.1109.0 Safari/536.8\", \"href\": \"https://kladomat-www.supersport.hr/ponuda/uzivo_kladomat?v=1450716027935\"}","value":0,"created":"2015-12-21T16:41:46.136Z"}`)
	m := NewBackendFromTopic(buf, "log.web_app_api")
	assert.NotNil(t, m)
	assert.Equal(t, "", m.IgracId)
	assert.Equal(t, buf, m.Body)
	assert.Equal(t, buf, m.RawBody)
	//t.Logf("%s", m.Pack())

	var l common.LogMessage
	m.UnmarshalBody(&l)
	assert.Equal(t, "kladomat_plazma", l.Source)

}
