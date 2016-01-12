package qcheck

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	qc := New(1000, func() time.Duration {
		return time.Millisecond * 10
	})
	assert.NotNil(t, qc)
	assert.Equal(t, 0, qc.Count())
	assert.Equal(t, false, qc.Any())
	ticker := time.NewTicker(time.Millisecond * 1)
	go func() {
		for range ticker.C {
			err := qc.Push()
			assert.Nil(t, err)
		}
	}()
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 100)
		assert.True(t, func() bool {
			c := qc.Count()
			fmt.Printf("count: %d\n", c)
			return c > 5 && c < 15
		}())
		assert.True(t, func() bool {
			lastDelta := qc.Last().Sub(time.Now())
			fmt.Printf("last delta: %v\n", lastDelta)
			return lastDelta > -time.Millisecond*20 && lastDelta < 0
		}())
	}
	assert.Equal(t, true, qc.Any())
	ticker.Stop()
}

func TestFull(t *testing.T) {
	qc := New(5, func() time.Duration {
		return time.Second
	})
	assert.NotNil(t, qc)
	assert.Equal(t, 0, qc.Count())
	ticker := time.NewTicker(time.Millisecond * 100)
	i := 0
	for range ticker.C {
		err := qc.Push()
		if i == 6 {
			assert.NotNil(t, err)
			break
		} else {
			assert.Nil(t, err)
			i++
		}
	}
	ticker.Stop()
}
