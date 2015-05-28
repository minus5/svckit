package util

import (
	"time"
)

// Sluzi kao guard oko taska koji se potencijalno dugo izvrsava.
// Pretplatnici koji zele znati kako je task zavrsio zovu Wait()
// U Wait posalju koliko su maksimalno spremni cekati.
// Kada task zavrsi Wait ce vratiti true, ili ako je timeout dostignut false.
// Done() oznacava task kao zavrsen (inicijalno je u working stanju).

type WaitTimeout struct {
	doneCh chan struct{}
}

func NewWaitTimeout() *WaitTimeout {
	return &WaitTimeout{doneCh: make(chan struct{})}
}

//Mark task as done, will release all goroutines blocked in Wait()
func (w *WaitTimeout) Done() {
	if !w.Finished() {
		close(w.doneCh)
	}
}

func (w *WaitTimeout) WaitInfinite() {
	select {
	case <-w.doneCh:
		return
	}
}

//Will wait here until Done is called or waitDuration is reached.
//Returns false if times out, otherwise true.
func (w *WaitTimeout) Wait(waitDuration time.Duration) bool {
	if waitDuration == 0 {
		w.WaitInfinite()
		return true
	}
	select {
	case <-w.doneCh: //block until Done is called
		return true
	case <-time.After(waitDuration):
		return false
	}
}

func (w *WaitTimeout) Finished() bool {
	select {
	case <-w.doneCh:
		return true
	default:
		return false
	}
}
