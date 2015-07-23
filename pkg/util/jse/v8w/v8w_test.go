package v8w

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TEST_JS_SOURCE = `27 + 15`
	TEST_EXPECT    = `42`
)

var v *V8W

func TestNew(t *testing.T) {
	var err error
	v, err = New(nil)
	assert.Nil(t, err)
	assert.NotNil(t, v)
	fmt.Printf("V8 version: %s\n", V8Version())
}

func TestEval(t *testing.T) {
	r, err := v.Eval(TEST_JS_SOURCE)
	assert.Nil(t, err)
	assert.Equal(t, TEST_EXPECT, r)
}

func TestEvalConcurrent(t *testing.T) {
	c := make(chan string)
	for i := 0; i < 10; i++ {
		go func() {
			fmt.Println("first start")
			defer fmt.Println("first end")
			r, err := v.Eval(TEST_JS_SOURCE)
			assert.Nil(t, err)
			assert.Equal(t, TEST_EXPECT, r)
		}()
	}
	go func() {
		fmt.Println("first start")
		defer fmt.Println("first end")
		r, err := v.Eval(TEST_JS_SOURCE)
		assert.Nil(t, err)
		assert.Equal(t, TEST_EXPECT, r)
	}()
	func() {
		fmt.Println("second start")
		defer fmt.Println("second end")
		r, err := v.Eval(TEST_JS_SOURCE)
		assert.Nil(t, err)
		assert.Equal(t, TEST_EXPECT, r)
	}()

}
