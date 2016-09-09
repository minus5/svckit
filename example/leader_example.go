// +build leader_example

package main

import (
	"github.com/minus5/svckit/leader"
	"github.com/minus5/svckit/log"
	"time"
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
			log.I("i", i).Debug("tick")
			time.Sleep(1e9)
			i++
		}
	}
}
