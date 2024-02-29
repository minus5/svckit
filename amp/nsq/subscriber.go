package nsq

import (
	"context"
	"sync"

	"github.com/minus5/svckit/amp"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/nsq"
	"github.com/pkg/errors"
)

type subscriber struct {
	subs []*nsq.Consumer
	out  chan *amp.Msg
	msgs sync.WaitGroup
}

func (s *subscriber) onMessage(m *nsq.Message) error {
	s.msgs.Add(1)
	defer s.msgs.Done()
	am := amp.ParseFromBackend(m.Body)
	if am == nil {
		return nil
	}
	if am.IsAlive() {
		log.Info("alive")
		return nil
	}
	s.out <- am
	return nil
}

func Subscribe(ctx context.Context, topics []string) <-chan *amp.Msg {
	out := make(chan *amp.Msg, 16)
	s := &subscriber{
		out: out,
	}
	if err := s.subscribe(topics); err != nil {
		log.Fatal(err)
	}
	go s.waitClose(ctx)
	return out
}

func (s *subscriber) waitClose(ctx context.Context) {
	<-ctx.Done()
	s.close()
	s.msgs.Wait()
	close(s.out)
}

func (s *subscriber) subscribe(topics []string) error {
	for _, topic := range topics {
		sub, err := nsq.NewConsumer(topic, s.onMessage, nsq.Ordered())
		if err != nil {
			return errors.WithStack(err)
		}
		s.subs = append(s.subs, sub)
	}
	return nil
}

func (s *subscriber) close() {
	for _, sub := range s.subs {
		sub.Close()
	}
}

// func responder(in <-chan *amp.Msg) {
// 	pub := nsq.Pub("")
// 	publish := func(m *amp.Msg) {
// 		topic := "dead.letter"
// 		if m.ReplyTopic != "" {
// 			topic = m.ReplyTopic
// 		}
// 		pub.PublishTo(topic, m.Marshal())
// 	}
// 	go func() {
// 		defer pub.Close()
// 		for m := range in {
// 			publish(m)
// 		}
// 	}()
// }
