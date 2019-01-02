// +build svc_reg_example

package main

import (
	"math/rand"
	"time"

	"github.com/mnu5/svckit/dcy/sr"
	"github.com/mnu5/svckit/health"
	"github.com/mnu5/svckit/leader"
	"github.com/mnu5/svckit/log"
)

func main() {
	health.Set(healthCheck)
	leader.New(worker)
}

func worker(stop <-chan struct{}) {
	port := rand.Intn(100) // using random port to enable multiple services on same host
	reg, err := sr.New(port, sr.HealthCheck(health.Get))
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
