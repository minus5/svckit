package sequence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequence(t *testing.T) {
	Set("test", 42)
	assert.Equal(t, 42, sequences["test"].last)
	assert.Equal(t, 43, Next("test"))
	assert.Equal(t, 43, sequences["test"].last)
	s, ok := Get("test")
	assert.True(t, ok)
	assert.NotNil(t, s)
	assert.Equal(t, 43, s.last)
	s, ok = Get("gulas")
	assert.False(t, ok)
	assert.Nil(t, s)
}
