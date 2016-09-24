// +build svc_reg_example

package main

import (
	"math/rand"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/leader"
	"github.com/minus5/svckit/log"
)

func main() {
	leader.New(worker)
}

func worker(stop <-chan struct{}) {
	i := 0
	reg, err := dcy.Register(rand.Intn(100))
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-stop:
			log.Debug("not e leader any more")
			reg.Deregister()
			return
		default:
			log.I("i", i).Debug("tick")
			time.Sleep(1e9)
			i++
		}
	}
}
