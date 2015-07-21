package qcheck

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	qc := New(1000, func() time.Duration {
		return time.Millisecond * 100
	})
	assert.NotNil(t, qc)
	assert.Equal(t, 0, qc.Count())
	go func() {
		for range time.NewTicker(time.Millisecond * 10).C {
			err := qc.Push()
			assert.Nil(t, err)
		}
		close(qc.c)
	}()
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		assert.True(t, func() bool {
			c := qc.Count()
			fmt.Printf("count: %d\n", c)
			return c > 8 && c < 12
		}())
		assert.True(t, func() bool {
			lastDelta := qc.Last().Sub(time.Now())
			fmt.Printf("last delta: %v\n", lastDelta)
			return lastDelta > -time.Millisecond*20 && lastDelta < 0
		}())
	}
}
