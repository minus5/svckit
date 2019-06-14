package util

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitTimeout(t *testing.T) {
	w := NewWaitTimeout()
	assert.NotNil(t, w)
	assert.False(t, w.Finished())

	var wg sync.WaitGroup
	wg.Add(4)
	var ok1, ok2, ok3, ok4 bool

	go func() {
		time.Sleep(time.Millisecond)
		w.Done()
		assert.True(t, w.Finished())
	}()
	go func() {
		ok1 = w.Wait(2 * time.Millisecond)
		wg.Done()
	}()
	go func() {
		ok2 = w.Wait(time.Nanosecond)
		wg.Done()
	}()
	go func() {
		ok3 = w.Wait(0)
		wg.Done()
	}()
	go func() {
		w.WaitInfinite()
		ok4 = true
		wg.Done()
	}()
	wg.Wait()

	assert.True(t, ok3)
	assert.True(t, ok1)
	assert.True(t, ok4)
	assert.False(t, ok2)

	ok := w.Wait(time.Hour)
	assert.True(t, ok)
}

func TestWorkerBugFixSecondCallToDone(t *testing.T) {
	w := NewWaitTimeout()
	w.Done()
	w.Done() //ovdje je bio panic close of a closed channel
}

func TestWorkerImpementation(t *testing.T) {
	w := newTestWorker()
	w.Start()

	var wg sync.WaitGroup
	wg.Add(2)
	var r1, r2 string
	go func() {
		r1 = w.GetResult(0)
		wg.Done()
	}()
	go func() {
		r2 = w.GetResult(time.Nanosecond)
		wg.Done()
	}()

	r3 := w.GetResult(2 * time.Millisecond)
	wg.Wait()

	assert.Equal(t, "done", r3)
	assert.Equal(t, "done", r1)
	assert.Equal(t, "timeout", r2)
}

type testWorker struct {
	worker *WaitTimeout
	result string
}

func newTestWorker() *testWorker {
	return &testWorker{worker: NewWaitTimeout()}
}

func (t *testWorker) Start() {
	go func() {
		//simuliram da nesto radi
		time.Sleep(time.Millisecond)
		t.result = "done"
		t.worker.Done()
	}()
}

func (t *testWorker) GetResult(waitDuration time.Duration) string {
	ok := t.worker.Wait(waitDuration)
	if !ok {
		return "timeout"
	}
	return t.result
}
