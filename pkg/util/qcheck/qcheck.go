package qcheck

import (
	"fmt"
	"sync"
	"time"
)

const (
	CHANNEL_BUFFER = 1000000
)

type timeFunc func() time.Duration

type QueueChecker struct {
	c            chan time.Time
	last         time.Time
	ticker       *time.Ticker
	intervalFunc timeFunc
	sync.RWMutex
}

func New(intervalFunc timeFunc) *QueueChecker {
	qc := &QueueChecker{
		c:            make(chan time.Time, CHANNEL_BUFFER),
		intervalFunc: intervalFunc,
	}
	go qc.drain()
	return qc
}

func (t *QueueChecker) drain() {
	for tm := range t.c {
		time.Sleep(tm.Sub(time.Now()))
	}
}

func (t *QueueChecker) Push() error {
	t.Lock()
	t.last = time.Now()
	t.Unlock()
	select {
	case t.c <- time.Now().Add(t.intervalFunc()):
		return nil
	default:
		return fmt.Errorf("qheck: channel full")
	}
}

func (t *QueueChecker) Count() int {
	return len(t.c)
}

func (t *QueueChecker) Last() time.Time {
	t.RLock()
	defer t.RUnlock()
	return t.last
}
