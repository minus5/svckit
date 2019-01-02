//Package broker prosljedjuje poruke svim consumerima nekog topica.
//Garantira poredak po topicu.
//Clean concurency and exit.
// Reference: https://www.enterpriseintegrationpatterns.com/patterns/messaging/MessageBroker.html
package broker

import (
	"pkg/amp/light/amp"

	"github.com/minus5/svckit/log"
)

// Broker type
type Broker struct {
	messages       chan *amp.Msg
	loopWork       chan func()
	closed         chan struct{}
	topics         map[string]*topic
	consumerTopics map[amp.Subscriber]map[string]int64
}

// Consume consumes all msgs from in channel.
func (s *Broker) Consume(in <-chan *amp.Msg) {
	go func() {
		defer s.signalClose()
		for m := range in {
			s.Publish(m)
		}
	}()
}

// Wait blocks until broker is finished
func (s *Broker) Wait() {
	<-s.closed
}

// New creates new scatter
func New() *Broker {
	s := &Broker{
		messages:       make(chan *amp.Msg, 1024),
		loopWork:       make(chan func()),
		closed:         make(chan struct{}),
		topics:         make(map[string]*topic),
		consumerTopics: make(map[amp.Subscriber]map[string]int64),
	}
	go s.loop()
	return s
}

func copyMap(o map[string]int64) map[string]int64 {
	n := make(map[string]int64)
	for k, v := range o {
		n[k] = v
	}
	return n
}

// Replay collects all current messages.
func (s *Broker) Replay(topic string) []*amp.Msg {
	var msgs []*amp.Msg
	s.inLoopWait(func() {
		if topic == "" || topic == "*" {
			for _, t := range s.topics {
				msgs = append(msgs, t.replay()...)
			}
			return
		}
		t, ok := s.topics[topic]
		if ok {
			msgs = append(msgs, t.replay()...)
		}
	})
	return msgs
}

// Subscribe consumer to topics defined c.Topics()
// amp.Subscriber should call this on each change ih his Topics list.
func (s *Broker) Subscribe(c amp.Subscriber, newTopics map[string]int64) {
	s.inLoop(func() {
		oldTopics, ok := s.consumerTopics[c]
		s.consumerTopics[c] = copyMap(newTopics)

		if !ok {
			for topic, ts := range newTopics {
				s.find(topic).subscribe(c, ts)
			}
			return
		}

		// proizvedi mapu promjena za one koje treba dodati true,
		// za one koje treba maknuti false
		updMap := make(map[string]bool)
		for t := range oldTopics {
			updMap[t] = false
		}
		for t := range newTopics {
			if _, ok := updMap[t]; ok {
				delete(updMap, t)
			} else {
				updMap[t] = true
			}
		}

		// obradi mapu promjena
		for t, v := range updMap {
			if v == true {
				s.find(t).subscribe(c, newTopics[t])
				continue
			}
			topic, ok := s.topics[t]
			if !ok {
				continue
			}
			if topic.unsubscribe(c) {
				delete(s.topics, t) // there is no one subscribed to this topic
				topic.close()
			}
		}
	})
}

func (s *Broker) find(topic string) *topic {
	t, ok := s.topics[topic]
	if !ok {
		log.S("topic", topic).Debug("new topic")
		t = newTopic()
		s.topics[topic] = t
	}
	return t
}

// Unsubscribe from all topics
func (s *Broker) Unsubscribe(c amp.Subscriber) {
	s.inLoopWait(func() {
		oldTopics := s.consumerTopics[c]
		delete(s.consumerTopics, c)
		for t := range oldTopics {
			topic, ok := s.topics[t]
			if !ok {
				continue
			}
			topic.unsubscribe(c)
		}
	})
}

func (s *Broker) inLoop(f func()) {
	select {
	case <-s.closed:
		return
	default:
	}
	s.loopWork <- func() {
		f()
	}
}

func (s *Broker) inLoopWait(f func()) {
	select {
	case <-s.closed:
		return
	default:
	}
	done := make(chan struct{})
	s.loopWork <- func() {
		f()
		close(done)
	}
	<-done
}

// Publish is interface for publisher.
func (s *Broker) Publish(m *amp.Msg) {
	s.messages <- m
}

func (s *Broker) signalClose() {
	close(s.messages)
}

func (s *Broker) close() {
	for _, t := range s.topics {
		t.close()
	}
	s.topics = make(map[string]*topic)
	close(s.closed)
}

func (s *Broker) loop() {
	for {
		select {
		case m := <-s.messages:
			if m == nil {
				s.close()
				return
			}
			t := m.Topic
			topic := s.find(t)
			if m.IsTopicClose() {
				log.S("topic", t).Debug("delete")
				delete(s.topics, t)
				topic.close()
			} else {
				topic.publish(m)
			}
		case f := <-s.loopWork:
			f()
		}
	}
}

// cekaj da se procesiraju poruke koje smo publish-ali
// samo za testove
func (s *Broker) wait(topic string) {
	for {
		ch := make(chan int)
		s.loopWork <- func() {
			ch <- len(s.messages)
		}
		if 0 == <-ch {
			s.topics[topic].wait()
			return
		}
	}
}

func (s *Broker) waitClose() {
	s.signalClose()
	s.Wait()
}
