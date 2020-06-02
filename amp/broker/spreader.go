package broker

import (
	"time"

	"github.com/minus5/svckit/amp"
)

type spreader struct {
	topicCount     int
	topics         []*topic
	consumerTopics map[amp.Sender]*topic
	pos            int
	lastUsed       time.Time
}

func newSpreader(name string, topicCount int) *spreader {
	s := &spreader{
		topicCount:     topicCount,
		topics:         []*topic{},
		consumerTopics: make(map[amp.Sender]*topic),
		lastUsed:       time.Now(),
	}
	for i := 0; i < topicCount; i++ {
		s.topics = append(s.topics, newTopic(name))
	}
	return s
}

func (spr *spreader) isExpired(expirePeriod time.Duration) bool {
	return len(spr.consumerTopics) == 0 && time.Now().Sub(spr.lastUsed) > expirePeriod
}

func (spr *spreader) findTopic(c amp.Sender) *topic {
	if t, ok := spr.consumerTopics[c]; ok {
		return t
	}
	t := spr.topics[spr.pos]
	spr.pos = (spr.pos + 1) % spr.topicCount
	spr.consumerTopics[c] = t
	return t
}

func (spr *spreader) subscribe(c amp.Sender, ts int64) {
	spr.lastUsed = time.Now()
	t := spr.findTopic(c)
	t.subscribe(c, ts)
}

func (spr *spreader) publish(m *amp.Msg) {
	spr.lastUsed = time.Now()
	for _, t := range spr.topics {
		t.messages <- m
	}
}

func (spr *spreader) close() {
	for _, t := range spr.topics {
		t.close()
	}
}

func (spr *spreader) unsubscribe(c amp.Sender) {
	if t := spr.consumerTopics[c]; t != nil {
		t.unsubscribe(c)
		delete(spr.consumerTopics, c)
	}
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
