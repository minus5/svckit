package main

import (
	"expvar"
	"net/http"
	"strconv"

	"github.com/mnu5/svckit/amp/broker"
	"github.com/mnu5/svckit/amp/nsq"
	"github.com/mnu5/svckit/amp/ws"
	"github.com/mnu5/svckit/env"

	"github.com/mnu5/svckit/amp/session"
	"github.com/mnu5/svckit/health"
	"github.com/mnu5/svckit/httpi"
	"github.com/mnu5/svckit/log"
	"github.com/mnu5/svckit/signal"

	_ "github.com/mnu5/svckit" // adding svckit.stats to expvar
)

var (
	inputTopics   = []string{"math.v1"}
	wsPortLabel   = "ws"
	demoPortLabel = "demo"
)

func main() {
	log.Debug("starting")
	defer log.Debug("stopped")

	tcpListener := ws.MustOpen(strconv.Itoa(env.Port(wsPortLabel))) // try to open ws port
	interupt := signal.InteruptContext()

	requester := nsq.MustRequester(interupt)
	defer requester.Wait()
	broker := broker.New()
	defer broker.Wait()
	broker.Consume(nsq.Subscribe(interupt, inputTopics))
	sessions := session.Factory(broker, requester)
	defer sessions.Wait()

	go debugHTTP()
	go demoServer()
	expvar.Publish("svckit.amp.broker", expvar.Func(broker.Expvar)) // TODO this is experimental

	ws.Listen(interupt, tcpListener, func(c *ws.Conn) { sessions.Serve(c) })
}

func debugHTTP() {
	health.Set(func() (health.Status, []byte) {
		return health.Passing, []byte("OK")
	})
	httpi.Start(env.Address(""))
}

func demoServer() {
	fs := http.FileServer(http.Dir("./demo/"))
	http.Handle("/", fs)
	http.ListenAndServe(env.Address(demoPortLabel), nil)
}
