package broker

import (
	"testing"

	"github.com/minus5/svckit/amp"
	"github.com/stretchr/testify/assert"
)

func TestTopicFindForSubscribe(t *testing.T) {
	topic := &topic{
		full: &amp.Msg{Ts: 10, UpdateType: amp.Full},
		diffs: []*amp.Msg{
			&amp.Msg{Ts: 10, UpdateType: amp.Diff},
			&amp.Msg{Ts: 11, UpdateType: amp.Diff},
			&amp.Msg{Ts: 12, UpdateType: amp.Diff},
			&amp.Msg{Ts: 13, UpdateType: amp.Diff},
		},
	}

	// nema nista dobije full
	msgs := topic.findForSubscribe(0)
	assert.NotNil(t, msgs)
	assert.Len(t, msgs, 4)

	// rubni dobije sve
	msgs = topic.findForSubscribe(9)
	assert.NotNil(t, msgs)
	assert.Len(t, msgs, 4)

	// nadopunimo ga diff-ovima
	msgs = topic.findForSubscribe(10)
	assert.Len(t, msgs, 3)
	assert.Equal(t, int64(11), msgs[0].Ts)
	assert.Equal(t, int64(12), msgs[1].Ts)
	assert.Equal(t, int64(13), msgs[2].Ts)

	msgs = topic.findForSubscribe(11)
	assert.Len(t, msgs, 2)
	assert.Equal(t, int64(12), msgs[0].Ts)
	assert.Equal(t, int64(13), msgs[1].Ts)

	msgs = topic.findForSubscribe(13)
	assert.Len(t, msgs, 0)

	// ovaj ima neki krivi, preveliki ts, ide od full
	msgs = topic.findForSubscribe(14)
	assert.Len(t, msgs, 4)

	topic.onMessage(&amp.Msg{Ts: 15, UpdateType: amp.Diff})
	msgs = topic.findForSubscribe(14)
	assert.Len(t, msgs, 1)
}

func TestTopicFindForSubscribeBeforeFull(t *testing.T) {
	topic := &topic{
		diffs: []*amp.Msg{
			&amp.Msg{Ts: 10, UpdateType: amp.Diff},
			&amp.Msg{Ts: 11, UpdateType: amp.Diff},
			&amp.Msg{Ts: 12, UpdateType: amp.Diff},
			&amp.Msg{Ts: 13, UpdateType: amp.Diff},
		},
	}
	msgs := topic.findForSubscribe(0)
	assert.Nil(t, msgs)
}

func TestTopicOnMessage(t *testing.T) {
	topic := &topic{
		full: &amp.Msg{Ts: 10, UpdateType: amp.Full},
		diffs: []*amp.Msg{
			&amp.Msg{Ts: 11, UpdateType: amp.Diff},
			&amp.Msg{Ts: 12, UpdateType: amp.Diff},
		},
	}

	// diff se dodaje
	topic.onMessage(&amp.Msg{Ts: 15, UpdateType: amp.Diff})
	assert.Len(t, topic.diffs, 3)

	// full postavlja novo stanje od kojeg krecemo
	topic.onMessage(&amp.Msg{Ts: 15, UpdateType: amp.Full})
	assert.Len(t, topic.diffs, 3)

	// diff koji je isti kao full, njega ne spremamo samo proslijedimo
	topic.onMessage(&amp.Msg{Ts: 15, UpdateType: amp.Diff})
	assert.Len(t, topic.diffs, 3)

	// stari njega preskacemo
	topic.onMessage(&amp.Msg{Ts: 14, UpdateType: amp.Diff})
	assert.Len(t, topic.diffs, 4)

	topic.onMessage(&amp.Msg{Ts: 14, UpdateType: amp.Diff})
	assert.Len(t, topic.diffs, 4)
}

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

func TestSortPrevRemovesDuplicates(t *testing.T) {
	topic := &topic{
		full: &amp.Msg{Ts: 10, UpdateType: amp.Full},
		diffs: []*amp.Msg{
			&amp.Msg{Ts: 10, UpdateType: amp.Diff},
			&amp.Msg{Ts: 12, UpdateType: amp.Diff},
			&amp.Msg{Ts: 12, UpdateType: amp.Diff},
			&amp.Msg{Ts: 15, UpdateType: amp.Diff},
			&amp.Msg{Ts: 15, UpdateType: amp.Diff, Replay: 1},
		},
	}
	topic.sortDiffs()
	assert.Len(t, topic.diffs, 3)
	assert.Equal(t, int64(10), topic.diffs[0].Ts)
	assert.Equal(t, int64(12), topic.diffs[1].Ts)
	assert.Equal(t, int64(15), topic.diffs[2].Ts)
}
