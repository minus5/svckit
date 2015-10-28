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

func TestRound(t *testing.T) {
	assert.Equal(t, 123.45, Round(123.45123, 2))
	assert.Equal(t, 123.46, Round(123.455, 2))
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
