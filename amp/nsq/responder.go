package nsq

import (
	"context"

	"github.com/ianic/svckit/log"
	"github.com/mnu5/svckit/amp"
	"github.com/mnu5/svckit/nsq"
)

type Responder struct {
	done    chan struct{}
	handler func(m *amp.Msg) (*amp.Msg, error)
}

func NewResponder(ctx context.Context,
	handler func(m *amp.Msg) (*amp.Msg, error),
	topic string) *Responder {

	r := &Responder{
		done:    make(chan struct{}),
		handler: handler,
	}

	in := Subscribe(ctx, []string{topic})
	go r.loop(in)
	return r
}

func (r *Responder) loop(in <-chan *amp.Msg) {
	defer close(r.done)

	pub := nsq.Pub("")
	defer pub.Close()

	for m := range in {
		rm, err := r.handler(m)
		if err != nil {
			log.Error(err)
			continue
		}
		if rm == nil || m.ReplyTo == "" {
			continue
		}
		if err := pub.PublishTo(m.ReplyTo, rm.Marshal()); err != nil {
			log.Error(err)
		}
	}
}

func (r *Responder) Wait() {
	<-r.done
}
