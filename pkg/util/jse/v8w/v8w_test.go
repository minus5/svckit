package v8w

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TestJsSource = `27 + 15`
	TestExpect   = `42`
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
	r, err := v.Eval(TestJsSource)
	assert.Nil(t, err)
	assert.Equal(t, TestExpect, r)
}

func TestEvalConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			fmt.Println("start", i)
			defer fmt.Println("done", i)
			r, err := v.Eval(TestJsSource)
			assert.Nil(t, err)
			assert.Equal(t, TestExpect, r)
			wg.Done()
		}(i)
	}
	wg.Add(1)
	go func() {
		fmt.Println("first start")
		defer fmt.Println("first end")
		r, err := v.Eval(TestJsSource)
		assert.Nil(t, err)
		assert.Equal(t, TestExpect, r)
		wg.Done()
	}()
	func() {
		fmt.Println("second start")
		defer fmt.Println("second end")
		r, err := v.Eval(TestJsSource)
		assert.Nil(t, err)
		assert.Equal(t, TestExpect, r)
	}()
	wg.Wait()
}
