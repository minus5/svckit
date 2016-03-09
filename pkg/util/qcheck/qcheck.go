package qcheck

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type timeFunc func() time.Duration

type QueueChecker struct {
	c            chan time.Time
	last         time.Time
	surplus      uint32
	ticker       *time.Ticker
	intervalFunc timeFunc
	sync.RWMutex
	first sync.Once
	any   bool
}

func Default() *QueueChecker {
	return New(1000., func() time.Duration { return time.Minute })
}

func New(maxSize int, intervalFunc timeFunc) *QueueChecker {
	qc := &QueueChecker{
		c:            make(chan time.Time, maxSize),
		intervalFunc: intervalFunc,
	}
	go qc.drain()
	return qc
}

func (t *QueueChecker) drain() {
	for tm := range t.c {
		atomic.StoreUint32(&t.surplus, 1)
		time.Sleep(tm.Sub(time.Now()))
		atomic.StoreUint32(&t.surplus, 0)
	}
}

func (t *QueueChecker) Push() error {
	t.Lock()
	t.last = time.Now()
	t.Unlock()
	t.first.Do(func() {
		t.any = true
	})
	select {
	case t.c <- time.Now().Add(t.intervalFunc()):
		return nil
	default:
		return fmt.Errorf("qheck: channel full")
	}
}

func (t *QueueChecker) Count() int {
	return len(t.c) + int(atomic.LoadUint32(&t.surplus))
}

func (t *QueueChecker) Last() time.Time {
	t.RLock()
	defer t.RUnlock()
	return t.last
}

func (t *QueueChecker) Any() bool {
	return t.any
}
