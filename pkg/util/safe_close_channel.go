package util

import "sync"

type SafeCloseChannel struct {
	C    chan interface{}
	once sync.Once
}

func NewSafeCloseChannel() *SafeCloseChannel {
	return &SafeCloseChannel{C: make(chan interface{})}
}

func (sc *SafeCloseChannel) SafeClose() {
	sc.once.Do(func() {
		close(sc.C)
	})
}