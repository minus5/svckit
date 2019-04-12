package main

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/minus5/svckit/amp"
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
	inputTopics   = []string{"math.v1", "chat"}
	wsPortLabel   = "ws"
	demoPortLabel = "demo"
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
	go demoServer()
	go poolingHTTP(interupt, sessions)
	ws.Listen(interupt, tcpListener, func(c *ws.Conn) { sessions.Serve(c) })
}

func poolingHTTP(interupt context.Context, sessions *session.Sessions) {
	srv := &http.Server{Addr: ":9090", Handler: &server{sessions: sessions}}
	go func() {
		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error(err)
		}
	}()
	go func() {
		<-interupt.Done()
		srv.Shutdown(context.Background())
	}()
}

type server struct {
	sessions *session.Sessions
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	defer r.Body.Close()

	m := amp.Parse(buf)
	if m == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rsp := s.sessions.Pool(m)
	for i, r := range rsp {
		w.Write(r.Marshal())
		if i < len(rsp)-1 {
			w.Write([]byte{10, 10})
		}
	}
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
