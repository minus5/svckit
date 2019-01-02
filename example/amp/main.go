package main

import (
	"expvar"

	"github.com/mnu5/svckit/example/amp/session"

	"github.com/mnu5/svckit/amp/broker"
	"github.com/mnu5/svckit/amp/nsq"
	"github.com/mnu5/svckit/amp/ws"

	"github.com/mnu5/svckit/health"
	"github.com/mnu5/svckit/httpi"
	"github.com/mnu5/svckit/log"
	"github.com/mnu5/svckit/signal"

	_ "github.com/mnu5/svckit" // adding svckit.stats to expvar
)

var (
	wsPort    = "8080"
	debugPort = "8081"
)

func main() {
	log.Debug("starting")

	tcpListener := ws.MustOpen(wsPort)      // try to open ws port
	interuptCtx := signal.InteruptContext() // application interupt signal

	requester := nsq.MustRequester(topicForMsgType)
	broker := broker.New()
	broker.Consume(nsq.Subscribe(interuptCtx, []string{"math.topics"}))
	sessions := session.Factory(broker, requester)

	go func() {
		requester.Wait(interuptCtx)
		broker.Wait()
		sessions.Close()
	}()
	go debugHTTP()
	expvar.Publish("svckit.amp.broker", expvar.Func(broker.Expvar))

	ws.Listen(interuptCtx, tcpListener, func(c *ws.Conn) { sessions.Serve(c) })
	sessions.Wait()
	log.Debug("stopped")
}

func debugHTTP() {
	health.Set(func() (health.Status, []byte) {
		return health.Passing, []byte("OK")
	})
	httpi.Start(":" + debugPort)
}

func topicForMsgType(msgType string) string {
	switch msgType {
	case "add":
		return "math.req"
	default:
		return "dead.letter"
	}
}
