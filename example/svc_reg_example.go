// +build svc_reg_example

package main

import (
	"math/rand"
	"time"

	"github.com/minus5/svckit/dcy/sr"
	"github.com/minus5/svckit/health"
	"github.com/minus5/svckit/leader"
	"github.com/minus5/svckit/log"
)

func main() {
	health.Set(healthCheck)
	leader.New(worker)
}

func worker(stop <-chan struct{}) {
	reg, err := sr.New(rand.Intn(100), sr.HealthCheck(health.Get))
	if err != nil {
		log.Fatal(err)
	}
	work(stop)
	reg.Deregister()
}

func work(stop <-chan struct{}) {
	i := 0
	for {
		select {
		case <-stop:
			log.Debug("not e leader any more")
			return
		default:
			log.I("i", i).Debug("working")
			time.Sleep(1e9)
			i++
		}
	}
}

func healthCheck() (health.Status, []byte) {
	return health.Passing, []byte("I'm really ok!")
}
