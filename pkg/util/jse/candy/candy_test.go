package candy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TEST_JS_SOURCE = `27 + 15`
	TEST_EXPECT    = `42`
)

var c *Candy

func TestNew(t *testing.T) {
	var err error
	c, err = New(nil)
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestEval(t *testing.T) {
	r, err := c.Eval(TEST_JS_SOURCE)
	assert.Nil(t, err)
	assert.Equal(t, TEST_EXPECT, r)
}
