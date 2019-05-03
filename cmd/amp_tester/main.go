package main

import (
	"context"
	"net/http"

	"github.com/minus5/svckit/amp/broker"
	"github.com/minus5/svckit/amp/nsq"
	"github.com/minus5/svckit/amp/ws"
	"github.com/minus5/svckit/env"

	"github.com/minus5/svckit/amp/session"
	"github.com/minus5/svckit/health"
	"github.com/minus5/svckit/httpi"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"

	_ "github.com/minus5/svckit" // adding svckit.stats to expvar
)

var (
	inputTopics      = []string{}
	debugPortLabel   = "debug"
	wsPortLabel      = "ws"
	logPortLabel     = "log"
	appPortLabel     = "app"
	poolingPortLabel = "pooling"
)

func main() {
	log.Debug("starting")
	defer log.Debug("stopped")

	tcpListener := ws.MustOpen(env.Port(wsPortLabel))
	interupt := signal.InteruptContext()
	requester := nsq.MustRequester(interupt)
	broker := broker.New(requester.Current)
	broker.Consume(nsq.Subscribe(interupt, inputTopics))
	sessions := session.Factory(interupt, broker, requester)
	defer sessions.Wait()

	go debugHTTP()
	go logHTTP(interupt)
	go poolingHTTP(interupt, sessions)
	go appServer()
	ws.Listen(interupt, tcpListener, func(c *ws.Conn) { sessions.Serve(c) })
}

func debugHTTP() {
	health.Set(func() (health.Status, []byte) {
		return health.Passing, []byte("OK")
	})
	httpi.Start(env.Address(debugPortLabel))
}

func logHTTP(interupt context.Context) {
	srv := &http.Server{Addr: env.Address(logPortLabel), Handler: &logger{}}
	go func() {
		<-interupt.Done()
		srv.Shutdown(context.Background())
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error(err)
	}
}

func poolingHTTP(interupt context.Context, sessions *session.Sessions) {
	srv := &http.Server{Addr: env.Address(poolingPortLabel), Handler: &pooling{sessions: sessions}}
	go func() {
		<-interupt.Done()
		srv.Shutdown(context.Background())
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error(err)
	}
}

func appServer() {
	fs := http.FileServer(http.Dir("./app/"))
	http.Handle("/", fs)
	http.ListenAndServe(env.Address(appPortLabel), nil)
}
