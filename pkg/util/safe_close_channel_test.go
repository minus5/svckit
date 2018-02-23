package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSafeCloseChannel(t *testing.T) {
	c := NewSafeCloseChannel()
	assert.NotNil(t, c)

	go func() {
		x := <-c.C
		assert.Equal(t, x, "abc")
	}()
	c.C <- "abc"

	c.SafeClose()
	assert.Nil(t, <-c.C)
	c.SafeClose()
	c.SafeClose()
	assert.Nil(t, <-c.C)
}

