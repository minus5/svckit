package broker

import (
	"testing"

	"github.com/minus5/svckit/amp"
	"github.com/stretchr/testify/assert"
)

func TestTopicReplay(t *testing.T) {
	topic := newTopic()
	m1 := &amp.Msg{Ts: 10, UpdateType: amp.Full}
	m2 := &amp.Msg{Ts: 11, UpdateType: amp.Diff}
	m3 := &amp.Msg{Ts: 12, UpdateType: amp.Diff}
	m4 := &amp.Msg{Ts: 13, UpdateType: amp.Diff}
	topic.publish(m1)
	topic.publish(m2)
	topic.publish(m4)
	topic.publish(m3)
	topic.wait()
	msgs := topic.replay()
	assert.Len(t, msgs, 4)
	assert.Equal(t, m1.Ts, msgs[0].Ts)
	assert.Equal(t, m2.Ts, msgs[1].Ts)
	assert.Equal(t, m3.Ts, msgs[2].Ts)
	assert.Equal(t, m4.Ts, msgs[3].Ts)
}
