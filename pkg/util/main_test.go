package util

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUuid(t *testing.T) {
	u := Uuid()
	assert.Equal(t, 36, len(u))
}

func TestCoverage(t *testing.T) {
	Hostname()
	InitLogger()
	InitLoggerNoFile()
	UnixMilli()

	fn := fmt.Sprintf("%s/%s", os.TempDir(), "util/test/file")
	err := WriteFile(fn, []byte{1, 2, 3})
	assert.Nil(t, err)
}

func TestStopwatch(t *testing.T) {
	sw := NewStopwatch()
	time.Sleep(1e7)
	elapsed := sw.GetMs()
	assert.InDelta(t, 10, elapsed, 10)
}

func TestSanitize(t *testing.T) {
	s := "Smeće: -/* Sí, Señor!"
	assert.Equal(t, "Smece-SiSenor", Sanitize(s))
}

func TestTimeUnixMilli(t *testing.T) {
	tt, _ := time.Parse("Jan 2, 2006 at 3:04 (MST)", "Oct 1, 2015 at 9:30 (CET)")
	assert.Equal(t, int64(1443688200000), TimeUnixMilli(tt))
}

func TestEqualFloat64(t *testing.T) {
	var src, dst, maxDelta float64
	src = 1.111
	dst = 1.111
	maxDelta = 0
	assert.True(t, EqualFloat64(src, dst, maxDelta))
	dst += 0.001
	assert.False(t, EqualFloat64(src, dst, maxDelta))
	maxDelta = 0.001
	assert.True(t, EqualFloat64(src, dst, maxDelta))
	assert.True(t, EqualFloat64(dst, src, maxDelta))
}

func TestEqualTime(t *testing.T) {
	var src, dst time.Time
	maxDeltaTime := 2 * time.Second
	src = time.Date(2010, 10, 24, 11, 1, 20, 0, time.Local)
	dst = time.Date(2010, 10, 24, 11, 1, 20, 0, time.Local)
	assert.True(t, EqualTime(src, dst, 0))
	assert.True(t, EqualTime(src, dst, maxDeltaTime))
	src = src.Add(maxDeltaTime)
	assert.False(t, EqualTime(src, dst, 0))
	assert.True(t, EqualTime(src, dst, maxDeltaTime))
	assert.True(t, EqualTime(dst, src, maxDeltaTime))
}

func TestStringArrayContains(t *testing.T) {
	var slice StringArray
	slice.Set("iso")
	slice.Set("medo")
	assert.True(t, slice.Contains("medo"))
	assert.False(t, slice.Contains("ducan"))
}

