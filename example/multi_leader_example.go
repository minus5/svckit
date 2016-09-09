// +build multi_leader_example

package main

import (
	"github.com/minus5/svckit/leader"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"
	"sync"
	"time"
)

func main() {
	// Primjer kako u aplikaciji koristimo dva leadera.
	// Svakom defirniramo key koji ce imati u consulu.
	// Naziv key-a u consulu ce biti npr: first-leadership_key
	//
	// WaitGroup postavljam da bi bio siguran da je svatko odradio cleanup.
	// Leaderi kao i aplikacija (u WaitForInterupt) reagiraju na TERM signal i pokrecu cleanup.
	// Wg osigurava da svi stignu obaviti cleanup inace se moze dogoditi da nam ostane lockan key y consulu.
	var wg sync.WaitGroup
	go leader.New(workerFactory("first"), leader.KeyPrefix("first"), leader.WaitGroup(&wg))
	go leader.New(workerFactory("second"), leader.KeyPrefix("second"), leader.WaitGroup(&wg))
	signal.WaitForInterupt()
	wg.Wait()
}

// workerFactory proizvodi novi worker koji ce u logove dodati key.
// Da bi mogao razlikovati dva workera.
func workerFactory(key string) func(<-chan struct{}) {
	return func(stop <-chan struct{}) {
		i := 0
		for {
			select {
			case <-stop:
				log.S("key", key).Debug("not e leader any more")
				return
			default:
				log.I("i", i).S("key", key).Debug("tick")
				time.Sleep(1e9)
				i++
			}
		}
	}
}
