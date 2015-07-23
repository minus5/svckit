package otto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TEST_JS_SOURCE = `27 + 15`
	TEST_EXPECT    = `42`
)

var o *Otto

func TestNew(t *testing.T) {
	var err error
	o, err = New(nil)
	assert.Nil(t, err)
	assert.NotNil(t, o)
}

func TestEval(t *testing.T) {
	r, err := o.Eval(TEST_JS_SOURCE)
	assert.Nil(t, err)
	assert.Equal(t, TEST_EXPECT, r)
}
