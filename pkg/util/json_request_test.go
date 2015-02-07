package util

import (
	"testing"
	"github.com/stretchrcom/testify/assert"
)

func TestCalcRetryInterval(t *testing.T) {
	r := NewJsonRequest("", []byte{0})
	r.RetrySleep = 1000
	assert.Equal(t, 1, r.calcRetryInterval(0))
	assert.Equal(t, 2, r.calcRetryInterval(1))
	assert.Equal(t, 7, r.calcRetryInterval(2))
	assert.Equal(t, 20, r.calcRetryInterval(3))
	assert.Equal(t, 54, r.calcRetryInterval(4))
	assert.Equal(t, 148, r.calcRetryInterval(5))
	assert.Equal(t, 403, r.calcRetryInterval(6))
	assert.Equal(t, 1000, r.calcRetryInterval(7))
}
