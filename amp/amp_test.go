package amp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURI(t *testing.T) {
	topic := "hr.mnu5"
	path := "resource/method"
	m := NewPublish(topic, path, 123, Full, nil)

	assert.Equal(t, path, m.Path())
	assert.Equal(t, topic, m.Topic())
	assert.Equal(t, topic+"/"+path, m.URI)
}

func TestURIWithoutPath(t *testing.T) {
	topic := "hr.mnu5"
	path := ""
	m := NewPublish(topic, path, 123, Full, nil)

	assert.Equal(t, path, m.Path())
	assert.Equal(t, topic, m.Topic())
	assert.Equal(t, topic, m.URI)
}

func TestPublish(t *testing.T) {
	o := struct {
		First string
		Last  string
	}{First: "jozo", Last: "bozo"}
	topic := "hr.mnu5"
	path := "resource/method"
	m := NewPublish(topic, path, 123, Full, o)

	buf := m.Marshal()
	expected := `{"u":"hr.mnu5/resource/method","s":123,"p":1}
{"First":"jozo","Last":"bozo"}`

	assert.Equal(t, expected, string(buf))
}

func TestParse(t *testing.T) {
	m := Parse(nil)
	assert.Nil(t, m)
}

func TestParseV1Subscribe(t *testing.T) {
	buf := `{"t":1,"u":[{"s":"m","n":93601933},{"s":"d_174626231","n":10},{"s":"s_2","n":11},{"s":"s_4","n":12}]}`
	m := ParseV1([]byte(buf))
	assert.NotNil(t, m)
	assert.Len(t, m.Subscriptions, 4)
	assert.Equal(t, m.Subscriptions["sportsbook/m"], int64(93601933))
	assert.Equal(t, m.Subscriptions["sportsbook/s_2"], int64(11))
	assert.Equal(t, m.Subscriptions["sportsbook/s_4"], int64(12))
	assert.Equal(t, m.Subscriptions["sportsbook/d_174626231"], int64(10))
}

func TestParseV1Ping(t *testing.T) {
	buf := `{"t":4}`
	m := ParseV1([]byte(buf))
	assert.NotNil(t, m)
	assert.Equal(t, m.Type, Ping)
}

func TestMarshalV1(t *testing.T) {
	m := &Msg{
		Type:       Publish,
		URI:        "sportsbook/m",
		Ts:         123,
		UpdateType: Full,
		body:       []byte(`{"First":"jozo","Last":"bozo"}`),
	}
	buf := m.MarshalV1()
	assert.Equal(t, string(buf), `{"s":"m","n":123,"f":1}
{"First":"jozo","Last":"bozo"}`)

	m.payloads = nil
	m.UpdateType = Diff
	buf = m.MarshalV1()
	assert.Equal(t, string(buf), `{"s":"m","n":123}
{"First":"jozo","Last":"bozo"}`)

	buf = m.Marshal()
	assert.Equal(t, string(buf), `{"u":"sportsbook/m","s":123}
{"First":"jozo","Last":"bozo"}`)

}

func TestParseV1Subscriptions(t *testing.T) {
	buf := []byte(`[{"s":"m","n":94067395},{"s":"s_4","n":1},{"s":"s_5","n":2}]`)
	m := ParseV1Subscriptions(buf)
	assert.NotNil(t, m)

	assert.Len(t, m.Subscriptions, 3)
	assert.Equal(t, m.Subscriptions["sportsbook/m"], int64(94067395))
	assert.Equal(t, m.Subscriptions["sportsbook/s_4"], int64(1))
	assert.Equal(t, m.Subscriptions["sportsbook/s_5"], int64(2))
}
