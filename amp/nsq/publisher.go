package nsq

import (
	"github.com/minus5/svckit/amp"
	"github.com/minus5/svckit/nsq"
)

func Publish(topic string, in <-chan *amp.Msg) chan *amp.Msg {
	pub := nsq.Pub(topic)
	publish := func(m *amp.Msg) {
		pub.Publish(m.Marshal())
	}
	out := make(chan *amp.Msg, 16)
	go func() {
		defer close(out)
		for m := range in {
			publish(m)
			out <- m
		}
	}()
	return out
}

type Publisher struct {
	done chan struct{}
}

func (p *Publisher) Wait() {
	<-p.done
}

func (p *Publisher) loop(in <-chan *amp.Msg) {
	defer close(p.done)

	pub := nsq.Pub("")
	publish := func(m *amp.Msg) {
		pub.PublishTo(m.Topic(), m.Marshal())
	}

	for m := range in {
		publish(m)
	}
}

func NewPublisher(in <-chan *amp.Msg) *Publisher {
	p := &Publisher{
		done: make(chan struct{}),
	}
	go p.loop(in)
	return p
}
