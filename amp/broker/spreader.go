package broker

import (
	"github.com/minus5/svckit/amp"
)

type spreader struct {
	topicCount     int
	topics         []*topic
	consumerTopics map[amp.MulSubscriber]*topic
	pos            int
}

func newSpreader(name string, topicCount int) *spreader {
	s := &spreader{
		topicCount:     topicCount,
		topics:         []*topic{},
		consumerTopics: make(map[amp.MulSubscriber]*topic),
	}
	for i := 0; i < topicCount; i++ {
		s.topics = append(s.topics, newTopic(name))
	}
	return s
}

func (spr *spreader) findTopic(c amp.MulSubscriber) *topic {
	if t, ok := spr.consumerTopics[c]; ok {
		return t
	}
	t := spr.topics[spr.pos]
	spr.pos = (spr.pos + 1) % spr.topicCount
	spr.consumerTopics[c] = t
	return t
}

func (spr *spreader) subscribe(c amp.MulSubscriber, ts int64) {
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

func (spr *spreader) unsubscribe(c amp.MulSubscriber) bool {
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
