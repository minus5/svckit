package main

import (
	"fmt"
	"time"

	"github.com/minus5/svckit/amp"
	"github.com/minus5/svckit/amp/broker"
	"github.com/minus5/svckit/amp/nsq"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/health"
	"github.com/minus5/svckit/httpi"
	"github.com/minus5/svckit/signal"
)

type msg struct {
	Nickname string `json:"nickname"`
	Comment  string `json:"comment"`
}

func chat(in <-chan msg) <-chan msg {
	out := make(chan msg, 1)
	go func() {
		defer close(out)
		for m := range in {
			out <- m
		}
	}()
	return out
}

var (
	inTopics = []string{"chat.req", "chat.current"}
	outTopic = "chat"
	outPath  = "1"
)

func main() {
	interupt := signal.InteruptContext()
	broker := broker.NewWithReplay()
	router := newRouter(broker.Replay)
	responder := nsq.NewResponder(interupt, router.entryPoint, inTopics)
	publisher := nsq.NewPublisher(broker.Pipe(msg2ampMsg(chat(router.in))))

	debugHTTP()
	responder.Wait()
	router.close()
	publisher.Wait()
}

type router struct {
	replay func(string)
	in     chan msg
}

func newRouter(replay func(string)) *router {
	return &router{
		replay: replay,
		in:     make(chan msg),
	}
}

func (r *router) entryPoint(m *amp.Msg) (*amp.Msg, error) {
	if m.IsCurrent() {
		r.replay(m.URI)
		return nil, nil
	}
	if !m.IsRequest() {
		return nil, nil
	}
	switch m.Path() {
	case "add":
		var cm msg
		if err := m.Unmarshal(&cm); err != nil {
			return nil, err
		}
		r.in <- cm
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown path %s", m.Path())
	}
	return nil, nil
}

func (r *router) close() {
	close(r.in)
}

func msg2ampMsg(in <-chan msg) <-chan *amp.Msg {
	out := make(chan *amp.Msg, 1)
	var i int
	go func() {
		defer close(out)
		for m := range in {
			am := amp.NewPublish(outTopic, outPath, time.Now().UnixNano(), amp.Append, m)
			if i%16 == 0 { // postavi tu i tamo, ne mora na svaku poruku
				am.CacheDepth = 16
			}
			i++
			out <- am
		}
	}()
	return out
}

func debugHTTP() {
	health.Set(func() (health.Status, []byte) {
		return health.Passing, []byte("OK")
	})
	httpi.Start(env.Address(""))
}
