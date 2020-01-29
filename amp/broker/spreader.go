package broker

import (
	"github.com/minus5/svckit/amp"
)

const topicCount = 4

type spreader struct {
	topics         [topicCount]*topic
	consumerTopics map[amp.Subscriber]*topic
	pos            int
}

func newSpreader(name string) *spreader {
	s := &spreader{
		topics:         [topicCount]*topic{},
		consumerTopics: make(map[amp.Subscriber]*topic),
	}
	for i := 0; i < topicCount; i++ {
		s.topics[i] = newTopic(name)
	}
	return s
}

func (spr *spreader) findTopic(c amp.Subscriber) *topic {
	if t, ok := spr.consumerTopics[c]; ok {
		return t
	}
	t := spr.topics[spr.pos]
	spr.pos = (spr.pos + 1) % topicCount
	spr.consumerTopics[c] = t
	return t
}

func (spr *spreader) subscribe(c amp.Subscriber, ts int64) {
	t := spr.findTopic(c)
	t.subscribe(c, ts)
}

func (spr *spreader) publish(m *amp.Msg) {
	for _, t := range spr.topics {
		t.messages <- m
	}
}

func (spr *spreader) close() {
	for _, t := range spr.topics {
		t.close()
	}
}

func (spr *spreader) unsubscribe(c amp.Subscriber) bool {
	t := spr.consumerTopics[c]
	if t != nil {
		t.unsubscribe(c)
		delete(spr.consumerTopics, c)
	}
	return len(spr.consumerTopics) == 0
}

func (spr *spreader) replay() []*amp.Msg {
	return spr.topics[0].replay()
}

// samo za testove
func (spr *spreader) wait() {
	for _, t := range spr.topics {
		t.wait()
	}
}
