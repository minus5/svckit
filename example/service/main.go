package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/mnu5/svckit/dcy/sr"
	"github.com/mnu5/svckit/health"
	"github.com/mnu5/svckit/httpi"
	"github.com/mnu5/svckit/leader"
	"github.com/mnu5/svckit/log"
)

var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 0, "http port to listen on")
	flag.Parse()
}

func main() {
	health.Set(healthCheck)
	httpi.Route("/ping", httpPing)
	go httpi.Start(fmt.Sprintf(":%d", port))
	leader.New(worker)
}

func worker(stop <-chan struct{}) {
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

func httpPing(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("pong"))
}

/*
# how to build docker image in my minus5 environment

env GOOS=linux GOARCH=amd64 go build &&  mv service ./docker
dm use dev/build
cd -
cd docker
docker build -t registry.dev.minus5.hr/test_service:0.1 .
docker push registry.dev.minus5.hr/test_service:0.1
*/
