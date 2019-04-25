package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
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
	inputTopics  = []string{}
	wsPortLabel  = "ws"
	logPortLabel = "log"
	appPortLabel = "app"
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
	go appServer()
	ws.Listen(interupt, tcpListener, func(c *ws.Conn) { sessions.Serve(c) })
}

func debugHTTP() {
	health.Set(func() (health.Status, []byte) {
		return health.Passing, []byte("OK")
	})
	httpi.Start(env.Address(""))
}

func logHTTP(interupt context.Context) {
	// TODO provjeri da ovdje nema debug interface
	//      uzmi zadnju verziju negroni i sto vec koristi httpi

	srv := &http.Server{Addr: env.Address(logPortLabel), Handler: &logger{}}
	go func() {
		<-interupt.Done()
		srv.Shutdown(context.Background())
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error(err)
	}
}

type logger struct{}

func (logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if path == "health_check" || path == "ping" {
		w.WriteHeader(http.StatusOK)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	hm := make(map[string]interface{})
	for k, v := range r.Header {
		if len(v) == 1 {
			hm[k] = v[0]
		} else {
			hm[k] = v
		}
	}
	h, err := json.Marshal(hm)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	l := log.Jc("body", body).
		Jc("header", h)
	if path == "error" {
		l.Error(nil)
	} else {
		l.Info(path)
	}

	w.WriteHeader(http.StatusOK)
}

func appServer() {
	fs := http.FileServer(http.Dir("./app/"))
	http.Handle("/", fs)
	http.ListenAndServe(env.Address(appPortLabel), nil)
}
