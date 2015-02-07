package msgs

import (
	"encoding/json"
	"testing"

	"github.com/stretchrcom/testify/assert"
)

func TestPushNotSerializeListic(t *testing.T) {
	m := &PushNot{Tip: 3,
		Listic: &PushNotListic{Id: "guid", Status: 2, Dobitak: 123.45, Broj: "broj"},
	}
	d := m.Serialize()
	assert.Equal(t, d["tip"], 3)
	assert.NotNil(t, d["listic"])
	assert.Nil(t, d["tekst"])
	assert.Equal(t, len(d), 2)
	l := d["listic"].(map[string]interface{})
	assert.Equal(t, l["id"], "guid")
	assert.Equal(t, l["status"], 2)
	assert.Equal(t, l["dobitak"], 123.45)
	assert.Equal(t, len(l), 4)

	buf, _ := json.Marshal(d)
	assert.Equal(t, string(buf), `{"listic":{"broj":"broj","dobitak":123.45,"id":"guid","status":2},"tip":3}`)
}

func TestPushNotSerializeTekst(t *testing.T) {
	m := &PushNot{Tip: 1, Tekst: "iso medo u ducan"}
	d := m.Serialize()
	assert.Equal(t, d["tip"], 1)
	assert.Nil(t, d["listic"])
	assert.NotNil(t, d["tekst"])
	assert.Equal(t, d["tekst"], "iso medo u ducan")
	assert.Equal(t, len(d), 2)
}
