package v8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TestJsSource = `27 + 15`
	TestExpect    = `42`
)

var v *V8

func TestNew(t *testing.T) {
	var err error
	v, err = New(nil)
	assert.Nil(t, err)
	assert.NotNil(t, v)
}

func TestEval(t *testing.T) {
	r, err := v.Eval(TestJsSource)
	assert.Nil(t, err)
	assert.Equal(t, TestExpect, r)
}
