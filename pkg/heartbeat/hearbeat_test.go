package heartbeat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.Len(t, heartbeats, 0)
	New(42, time.Second)
	assert.Len(t, heartbeats, 1)
	New(42, time.Second)
	New(43, time.Second)
	New(44, time.Second)
	assert.Len(t, heartbeats, 3)
}

func TestHearbeat(t *testing.T) {
	New(1, time.Millisecond*100)
	assert.True(t, OK(1))
	time.Sleep(time.Millisecond * 200)
	assert.False(t, OK(1))
	Heartbeat(1, time.Now())
	assert.True(t, OK(1))
	time.Sleep(time.Millisecond * 200)
	assert.False(t, OK(1))
}
