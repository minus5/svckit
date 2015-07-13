package qcheck

import (
	"sync/atomic"
	"time"
)

type timeFunc func() time.Duration

type QueueChecker struct {
	c            chan struct{}
	count        uint64
	last         time.Time
	ticker       *time.Ticker
	intervalFunc timeFunc
}

func New(c chan struct{}, intervalFunc timeFunc) *QueueChecker {
	qc := &QueueChecker{
		c:            c,
		intervalFunc: intervalFunc,
	}
	qc.check()
	return qc
}

func (t *QueueChecker) check() {
	go func() {
		for range t.c {
			atomic.AddUint64(&t.count, 1)
			t.last = time.Now()
		}
	}()
	go func() {
		for {
			time.Sleep(t.intervalFunc())
			atomic.StoreUint64(&t.count, 0)
		}
	}()
}

func (t *QueueChecker) Count() uint64 {
	return atomic.LoadUint64(&t.count)
}

func (t *QueueChecker) Last() time.Time {
	return t.last
}
