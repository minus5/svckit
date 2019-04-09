// +build leader_example

package main

import (
	"time"

	"github.com/minus5/svckit/leader"
	"github.com/minus5/svckit/log"
)

func main() {
	leader.New(worker)
}

func worker(stop <-chan struct{}) {
	i := 0
	for {
		select {
		case <-stop:
			log.Debug("not e leader any more")
			return
		default:
			for j := 1; j < 10; j++ {
				log.I("i", i).I("j", j).Debug("tick")
				time.Sleep(1e9)
			}
			i++
		}
	}
}
