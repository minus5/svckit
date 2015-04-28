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
	queue   chan bool
	working bool
}

func NewWaitTimeout() *WaitTimeout {
	return &WaitTimeout{queue: make(chan bool), working: true}
}

//Mark task as done, will release all goroutines blocked in Wait()
func (w *WaitTimeout) Done() {
	if !w.working {
		return
	}
	w.working = false
	w.notifyWaiters()
	close(w.queue)
}

//reading from queue will release all writers
func (w *WaitTimeout) notifyWaiters() {
	for {
		select {
		case <-w.queue:
		default:
			return
		}
	}
}

func (w *WaitTimeout) WaitInfinite() bool {
	return w.Wait(0)
}

//Will wait here until Done is called or waitDuration is reached.
//Returns false if times out, otherwise true.
func (w *WaitTimeout) Wait(waitDuration time.Duration) bool {
	if w.Finished() {
		return true
	}
	if waitDuration == 0 {
		waitDuration = time.Hour
	}
	select {
	//block until someone starts reading from queue
	case w.queue <- true:
		return true
	case <-time.After(waitDuration):
		return false
	}
}

func (w *WaitTimeout) Finished() bool {
	return !w.working
}
