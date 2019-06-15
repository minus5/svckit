// Package qcheck sluzi za provjeru nagomilanih poruka u queue-u
//
// Princip rada:
// - Definira se koliko max poruka smije biti u queueu za neki period vremena
// - QueueChecker prazni poruke iz queue-a tako da
//	- za procitanu poruku ceka definirani vremenski period prije nego nastavi citati slijedecu
//	- tako prva poruka koja udje u queue definira vrijeme praznjenja queue-a
//	- poruke mogu ulaziti u queue sve dok se queue ne popuni
package qcheck

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type timeFunc func() time.Duration

// QueueChecker struktura za provjeru dubine poruka u nekom queue-u
type QueueChecker struct {
	c            chan time.Time // c			- kanal u koji ulaze vremena kad je poruka poslana
	quit         chan bool      // quit			- kanal koji sluzi za izlaz iz QueueChecker-a
	finished     chan bool      // finished		- kanal koji sluzi flag kad smo zavrsili
	maxSize      int            // maxSize		- max broj dovoljenih poruka u intervalu
	last         time.Time      // last 		- vrijeme zadnje poruke u queue-u
	surplus      uint32         // surplus		- broj poruka trenutno u drain-u, uobicajeno jedna
	intervalFunc timeFunc       // intervalFunc - funkcija koja trajanje za max poruka u queue-u
	any          bool           // any			- oznaka da je u queue usla barem 1 poruka
	sync.RWMutex
}

// Default QueueChecker sa max dubinom 10000 poruka u peroidu od 1 minute
func Default() *QueueChecker {
	return New(10000, func() time.Duration { return time.Minute })
}

// New2 kreira novi QueueChecker za
// - maxSize	- max broj broj poruka u queue-u
// - d			- period u kojem se smije poslati max broj poruka
func New2(maxSize int, d time.Duration) *QueueChecker {
	return New(maxSize, func() time.Duration { return d })
}

// New kreira novi QueueChecker
// - maxSize		- max broj broj poruka u queue-u
// - intervalFunc	- funkcija koja vraca interval za max broj poruka
func New(maxSize int, intervalFunc timeFunc) *QueueChecker {
	qc := &QueueChecker{
		c:            make(chan time.Time, maxSize),
		quit:         make(chan bool, 1),
		finished:     make(chan bool, 1),
		intervalFunc: intervalFunc,
		maxSize:      maxSize,
	}
	go qc.drain()
	return qc
}

// drain prazni poruke iz queue-a tako da
// - ceka dok poruka ne treba izaci iz queue-a za zadani period
// - na ovaj nacin se poruke nakupljuaju u queueu dok se
func (t *QueueChecker) drain() {
	defer close(t.finished)
	for {
		select {
		case tm := <-t.c:
			if t.sleepUntil(tm) {
				return
			}
		case <-t.quit:
			return
		}
	}
}

// sleepUntil ceka dok poruka u queueu ne istekne ili se prekida
// - vraca 	- true ako je prekinuto
func (t *QueueChecker) sleepUntil(tm time.Time) bool {
	wait := tm.Sub(time.Now())
	if wait.Nanoseconds() < 0 {
		return false
	}
	atomic.StoreUint32(&t.surplus, 1)
	defer atomic.StoreUint32(&t.surplus, 0)
	sleep := time.NewTimer(wait)
	select {
	case <-sleep.C:
		return false
	case <-t.quit:
		// Stopiraj i ocisti sleep, izlazimo
		if !sleep.Stop() {
			<-sleep.C
		}
		return true
	}
}

// Push poziva se kad nova poruka ulazi u queue
func (t *QueueChecker) Push() error {
	t.Lock()
	defer t.Unlock()
	now := time.Now()
	t.last = now
	t.any = true
	select {
	case t.c <- now.Add(t.intervalFunc()):
		return nil
	default:
		return fmt.Errorf("qheck: channel full")
	}
}

// Count broj poruka u queue-u
func (t *QueueChecker) Count() int {
	return len(t.c) + int(atomic.LoadUint32(&t.surplus))
}

// Full da li je queue trenutno zapunjen
func (t *QueueChecker) Full() bool {
	return t.Count() >= t.maxSize
}

// Last vraca vrijeme zadnje poruke u queue-u
func (t *QueueChecker) Last() time.Time {
	t.RLock()
	defer t.RUnlock()
	return t.last
}

// Any vraca da li je usla barem jedna poruka
func (t *QueueChecker) Any() bool {
	return t.any
}

// Close se poziva kad se QueueChecker vise ne koristi
// - zaustavlja obradu poruka iz queue-a
func (t *QueueChecker) Close() {
	close(t.quit)
	<-t.finished
}
