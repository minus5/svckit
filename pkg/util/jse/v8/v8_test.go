package v8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TEST_JS_SOURCE = `27 + 15`
	TEST_EXPECT    = `42`
)

var v *V8

func TestNew(t *testing.T) {
	var err error
	v, err = New(nil)
	assert.Nil(t, err)
	assert.NotNil(t, v)
}

func TestEval(t *testing.T) {
	r, err := v.Eval(TEST_JS_SOURCE)
	assert.Nil(t, err)
	assert.Equal(t, TEST_EXPECT, r)
}
