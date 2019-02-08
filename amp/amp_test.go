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
