package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEntry(t *testing.T) {
	e, err := NewEntry([]byte(`
		{
			"time": "2009-11-10T23:05:06.000000+00:00",
			"file": "main.go:123",
			"level": "notice",
			"host": "localhost",
			"keyi": 123,
			"keyf": 3.1415926535,
			"key": "val",
			"msg": "{\"inside\":2,\"inside2\":4}"
			}`))
	assert.NoError(t, err)
	if !assert.NotNil(t, e) {
		t.FailNow()
	}
	assert.Equal(t, 2009, e.Time.Year())
	assert.Equal(t, 6, e.Time.Second())
	assert.Equal(t, "main.go:123", e.File)
	assert.Equal(t, "notice", e.Level)
	i, ok := e.I("keyi")
	if !assert.True(t, ok) {
		t.FailNow()
	}
	assert.Equal(t, 123, i)
	f, ok := e.F("keyf")
	if !assert.True(t, ok) {
		t.FailNow()
	}
	assert.Equal(t, 3.1415926535, f)
	s, ok := e.S("key")
	if !assert.True(t, ok) {
		t.FailNow()
	}
	assert.Equal(t, "val", s)
	assert.Equal(t, "{\"inside\":2,\"inside2\":4}", e.Msg)
}
