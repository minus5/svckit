package broker

import "github.com/minus5/svckit/amp"

type ReplayBroker struct {
	messages chan *amp.Msg
	broker   *Broker
}

func NewWithReplay() *ReplayBroker {
	return &ReplayBroker{
		messages: make(chan *amp.Msg),
		broker:   New(nil, nil),
	}
}

func (r *ReplayBroker) Pipe(in <-chan *amp.Msg) <-chan *amp.Msg {
	out := make(chan *amp.Msg)

	go func() {
		defer close(out)
		for {
			select {
			case m, ok := <-in:
				if !ok {
					r.broker.signalClose()
					go func() {
						// drain the r.messages chan
						for range r.messages {
						}
					}()
					return
				}
				r.broker.Publish(m)
				out <- m
			case m := <-r.messages:
				out <- m
			}
		}
	}()

	return out
}

func (r *ReplayBroker) Replay(topic string) {
	for _, m := range r.broker.Replay(topic) {
		r.messages <- m
	}
}
