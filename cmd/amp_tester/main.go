package main

import (
	"context"
	"time"

	"github.com/minus5/svckit/amp/broker"
	"github.com/minus5/svckit/amp/nsq"
	"github.com/minus5/svckit/amp/session"
	"github.com/minus5/svckit/amp/ws"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/health"
	"github.com/minus5/svckit/httpi"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"

	"github.com/minus5/svckit/metric/statsd"
	"github.com/minus5/svckit/signal"

	_ "github.com/minus5/svckit" // adding svckit.stats to expvar
)

var (
	inputTopics    = []string{"math.v1"}
	debugPortLabel = "debug"
	wsPortLabel    = "ws"
	appPortLabel   = "app"
	appRoot        = "./app/"
)

func main() {
	log.Debug("starting")
	defer log.Debug("stopped")

	statsd.TryDial(statsd.MetricPrefix(env.AppName()))
	defer statsd.Close()

	tcpListener := ws.MustOpen(env.Port(wsPortLabel))
	interupt := signal.InteruptContext()
	requester := nsq.MustRequester(interupt)
	broker := broker.New(requester.Current)
	broker.Consume(nsq.Subscribe(interupt, inputTopics))
	sessions := session.Factory(interupt, broker, requester)
	defer sessions.Wait()

	go debugHTTP()
	go appHTTPServer(interupt, appPortLabel, appRoot, sessions)
	go stats(interupt, sessions)
	ws.Listen(interupt, tcpListener, func(c *ws.Conn) { sessions.Serve(c) })
}

func debugHTTP() {
	health.Set(func() (health.Status, []byte) {
		return health.Passing, []byte("OK")
	})
	httpi.Start(env.Address(debugPortLabel))
}

func stats(interupt context.Context, sessions *session.Sessions) {
	for {
		select {
		case <-interupt.Done():
			return
		case <-time.After(10 * time.Second):
			ws, pooling := sessions.ConnectionsCount()
			metric.Gauge("sessions.ws", ws)
			metric.Gauge("sessions.pooling", pooling)
		}
	}
}
