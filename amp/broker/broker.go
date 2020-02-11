//Package broker prosljedjuje poruke svim consumerima nekog topica.
//Garantira poredak po topicu.
//Clean concurency and exit.
// Reference: https://www.enterpriseintegrationpatterns.com/patterns/messaging/MessageBroker.html
package broker

import (
	"github.com/minus5/svckit/amp"
	"github.com/minus5/svckit/log"
	"time"
)

// Broker type
type Broker struct {
	messages      chan *amp.Msg
	loopWork      chan func()
	closed        chan struct{}
	spreaders     map[string]*spreader
	consumerNames map[amp.Subscriber]map[string]int64
	current       func(string)
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
func New(current func(string)) *Broker {
	s := &Broker{
		messages:      make(chan *amp.Msg, 1024),
		loopWork:      make(chan func()),
		closed:        make(chan struct{}),
		spreaders:     make(map[string]*spreader),
		consumerNames: make(map[amp.Subscriber]map[string]int64),
		current:       current,
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
func (s *Broker) Replay(name string) []*amp.Msg {
	log.Debug("replay start")
	var msgs []*amp.Msg
	s.inLoopWait(func() {
		if name == "" || name == "*" {
			for _, spr := range s.spreaders {
				msgs = append(msgs, spr.replay()...)
			}
			return
		}
		spr, ok := s.spreaders[name]
		if ok {
			msgs = append(msgs, spr.replay()...)
		}
	})
	log.I("msgs", len(msgs)).Debug("replay end")
	return msgs
}

// Subscribe consumer to topics defined c.Topics()
// amp.Subscriber should call this on each change ih his Topics list.
func (s *Broker) Subscribe(c amp.Subscriber, newNames map[string]int64) {
	metric.Time("broker.subscribe.len", len(newNames))
	s.inLoop(func() {
		oldNames, ok := s.consumerNames[c]
		s.consumerNames[c] = copyMap(newNames)

		if !ok {
			for name, ts := range newNames {
				s.find(name, true).subscribe(c, ts)
			}
			return
		}

		// proizvedi mapu promjena za one koje treba dodati true,
		// za one koje treba maknuti false
		updMap := make(map[string]bool)
		for t := range oldNames {
			updMap[t] = false
		}
		for name := range newNames {
			if _, ok := updMap[name]; ok {
				delete(updMap, name)
			} else {
				updMap[name] = true
			}
		}

		// obradi mapu promjena
		for name, v := range updMap {
			if v == true {
				s.find(name, true).subscribe(c, newNames[name])
				continue
			}
			spr, ok := s.spreaders[name]
			if !ok {
				continue
			}
			if spr.unsubscribe(c) {
				log.S("topic", name).Info("delete from uns")
				delete(s.spreaders, name) // there is no one subscribed to this topic
				spr.close()
			}
		}
	})
}

func (s *Broker) find(name string, currentOnNew bool) *spreader {
	spr, ok := s.spreaders[name]
	if !ok {
		start := time.Now()
		topicCount := 1
		if name == "sportsbook/m" {
			topicCount = 16
		}
		spr = newSpreader(name, topicCount)
		s.spreaders[name] = spr
		if currentOnNew && s.current != nil {
			log.S("topic", name).I("count", topicCount).Info("new top current")
			go s.current(name)
		} else {
			log.S("topic", name).I("count", topicCount).Info("new topic")
		}
		metric.Time("topic.new", int(time.Now().Sub(start).Nanoseconds()))
	}
	return spr
}

// Unsubscribe from all topics
func (s *Broker) Unsubscribe(c amp.Subscriber) {
	s.inLoopWait(func() {
		oldNames := s.consumerNames[c]
		delete(s.consumerNames, c)
		for name := range oldNames {
			spr, ok := s.spreaders[name]
			if !ok {
				continue
			}
			spr.unsubscribe(c)
		}
	})
}

func (s *Broker) inLoop(f func()) {
	select {
	case <-s.closed:
		return
	default:
	}
	call := time.Now()
	s.loopWork <- func() {
		enter := time.Now()
		metric.Time("broker.subscribe.wait", int(enter.Sub(call).Nanoseconds()))
		defer func() {
			metric.Time("broker.subscribe.run", int(time.Now().Sub(enter).Nanoseconds()))
		}()
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
	call := time.Now()
	s.loopWork <- func() {
		enter := time.Now()
		metric.Time("broker.subscribe.wait", int(enter.Sub(call).Nanoseconds()))
		defer func() {
			metric.Time("broker.subscribe.run", int(time.Now().Sub(enter).Nanoseconds()))
		}()
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
	for _, spr := range s.spreaders {
		spr.close()
	}
	s.spreaders = make(map[string]*spreader)
	close(s.closed)
}

func (s *Broker) loop() {
	for {
		select {
		case m := <-s.messages:
			start := time.Now()
			if m == nil {
				s.close()
				return
			}
			name := m.URI
			spr := s.find(name, !m.IsFull())
			if m.IsTopicClose() {
				log.S("topic", name).Info("delete from msg")
				delete(s.spreaders, name)
				spr.close()
			} else {
				spr.publish(m)
			}
			metric.Time("broker.loop.msg", int(time.Now().Sub(start).Nanoseconds()))
		case f := <-s.loopWork:
			start := time.Now()
			f()
			metric.Time("broker.loop.work", int(time.Now().Sub(start).Nanoseconds()))
		}
	}
}

// cekaj da se procesiraju poruke koje smo publish-ali
// samo za testove
func (s *Broker) wait(name string) {
	for {
		ch := make(chan int)
		s.loopWork <- func() {
			ch <- len(s.messages)
		}
		if 0 == <-ch {
			s.spreaders[name].wait()
			return
		}
	}
}

func (s *Broker) waitClose() {
	s.signalClose()
	s.Wait()
}

// func (s *Broker) Expvar() interface{} {
// 	m := make(map[string]interface{})
// 	s.inLoopWait(func() {
// 		for k, t := range s.topics {
// 			tm := make(map[string]interface{})
// 			diffs, firstDiffTs, lastDiffTs, fullTs := t.metrics()
// 			tm["diffs"] = diffs
// 			tm["firstDiffTs"] = firstDiffTs
// 			tm["lastDiffTs"] = lastDiffTs
// 			tm["fullTs"] = fullTs
// 			m[k] = tm
// 		}
// 	})
// 	return m
// }

func (s *Broker) Gauges() (int, int, int) {
	return len(s.messages), len(s.spreaders), len(s.consumerNames)
}
