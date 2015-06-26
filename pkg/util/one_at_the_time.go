package util

import "sync/atomic"

type OneAtTheTime struct {
	mu uint32
}

// Do Executes f one at a time.
// Other calls are not blocked and returns false.
func (t *OneAtTheTime) Do(f func()) bool {
	if t.tryLock() {
		defer t.unlock()
		f()
		return true
	}
	return false
}

// tryMutex Returns true if lock if acquired, otherwise false.
func (t *OneAtTheTime) tryLock() bool {
	return atomic.CompareAndSwapUint32(&t.mu, 0, 1)
}

// unlock Call this when finished, after successful tryMutex
func (t *OneAtTheTime) unlock() {
	atomic.StoreUint32(&t.mu, 0)
}
