package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testTime = time.Date(2009, time.November, 10, 23, 5, 6, 7, time.UTC)

func newTestAgregator() *Agregator {
	prefix = []byte{}
	a := newAgregator(3)
	a.t = testTime
	a.file = "main.go"
	a.line = 123
	return a
}

func currentBuffer(a *Agregator) string {
	buf := *a.buf
	return string(buf[0 : len(buf)-1])
}

func TestAgregator(t *testing.T) {
	a := newTestAgregator()
	a.Info("msg")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"info", "msg":"msg"}`, currentBuffer(a))

	a.I("keyi", 123).F("keyf", 3.1415926535, -1).S("key", "val").Notice("msg")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"notice", "keyi":123, "keyf":3.1415926535, "key":"val", "msg":"msg"}`, currentBuffer(a))
}

func TestLimitStrLen(t *testing.T) {
	mali := strings.Repeat("0", MaxStrLen)
	veliki := strings.Repeat("0", MaxStrLen+1)
	assert.Equal(t, mali, limitStrLen(mali))
	l := limitStrLen(veliki)
	assert.NotEqual(t, veliki, l)
	assert.Equal(t, MaxStrLen, len(l))
	assert.Equal(t, "0...\"", l[len(l)-5:])

	quoted := strings.Repeat("0", MaxStrLen-10) + strings.Repeat("\\", 11)
	l = limitStrLen(quoted)
	assert.Equal(t, MaxStrLen-10+4, len(l))
	assert.Equal(t, "0...\"", l[len(l)-5:])
}

// itoa ne radi za negativne
func TestItoa(t *testing.T) {
	buf := make([]byte, 0)
	itoa(&buf, -2, -1)
	//fmt.Printf("buf: %s", buf)
}

func TestReservedKeys(t *testing.T) {
	a := newTestAgregator()
	a.S("msg", "reserved msg").Info("test")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"info", "_msg":"reserved msg", "msg":"test"}`, currentBuffer(a))
	a = newTestAgregator()
	a.S("app", "reserved app").Info("test")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"info", "_app":"reserved app", "msg":"test"}`, currentBuffer(a))
	a = newTestAgregator()
	a.S("level", "reserved level").Info("test")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"info", "_level":"reserved level", "msg":"test"}`, currentBuffer(a))
	a = newTestAgregator()
	a.S("host", "reserved host").Info("test")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"info", "_host":"reserved host", "msg":"test"}`, currentBuffer(a))
	a = newTestAgregator()
	a.S("file", "reserved file").Info("test")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"info", "_file":"reserved file", "msg":"test"}`, currentBuffer(a))
	a = newTestAgregator()
	a.S("time", "reserved time").Info("test")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"info", "_time":"reserved time", "msg":"test"}`, currentBuffer(a))
}

func TestJ(t *testing.T) {
	a := newTestAgregator()
	a.J("json", []byte(`{"some":"valid","json":2.1}`)).Debug("msg")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"debug", "json":{"some":"valid","json":2.1}, "msg":"msg"}`, currentBuffer(a))
}

func TestJc(t *testing.T) {
	a := newTestAgregator()
	s := struct {
		Some string  `json:"some"`
		Json float64 `json:"json"`
	}{"valid", 2.1}
	buf, _ := json.Marshal(s)
	a.Jc("json", buf).Debug("msg")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"debug", "json":{"some":"valid","json":2.1}, "msg":"msg"}`, currentBuffer(a))
}

func TestEmptyJson(t *testing.T) {
	a := newTestAgregator()
	var buf []byte
	a.J("key", buf).Debug("empty json")
	assert.Equal(t, `{"time":"2009-11-10T23:05:06.000000+00:00", "file":"main.go:123", "level":"debug", "key":null, "msg":"empty json"}`, currentBuffer(a))
	fmt.Println(currentBuffer(a))
}
