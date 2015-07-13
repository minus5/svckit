package qcheck

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	c := make(chan struct{}, 1000)
	qc := New(c, func() time.Duration {
		return time.Millisecond * 100
	})
	assert.NotNil(t, c)
	assert.NotNil(t, qc)
	assert.Equal(t, uint64(0), qc.Count())
	go func() {
		for range time.NewTicker(time.Millisecond * 10).C {
			qc.c <- struct{}{}
		}
		close(qc.c)
	}()
	time.Sleep(time.Second)
	assert.True(t, func() bool {
		c := qc.Count()
		return c > 8 && c < 12
	}())
	assert.True(t, func() bool {
		lastDelta := qc.Last().Sub(time.Now())
		return lastDelta > -time.Millisecond*20 && lastDelta < 0
	}())
}
