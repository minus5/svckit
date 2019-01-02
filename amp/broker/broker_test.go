package broker

import (
	"sync"
	"testing"

	"github.com/mnu5/svckit/amp"
	"github.com/mnu5/svckit/log"
	"github.com/stretchr/testify/assert"
)

type testConsumer struct {
	topics   map[string]int64
	messages []*amp.Msg
	sync.Mutex
}

func (c *testConsumer) Send(m *amp.Msg) {
	c.Lock()
	defer c.Unlock()
	c.messages = append(c.messages, m)
}

func TestDvaTopica(t *testing.T) {
	log.Discard()
	s := New()
	c := &testConsumer{topics: map[string]int64{"1": 0, "2": 0}}
	s.Subscribe(c, c.topics)
	m10 := &amp.Msg{Topic: "1", Ts: 1, UpdateType: amp.Full}
	m11 := &amp.Msg{Topic: "1", Ts: 2, UpdateType: amp.Diff}
	m20 := &amp.Msg{Topic: "2", Ts: 1, UpdateType: amp.Full}
	m30 := &amp.Msg{Topic: "3", Ts: 1, UpdateType: amp.Full}
	s.Publish(m10)
	s.Publish(m11)
	s.Publish(m20)
	s.Publish(m30)
	s.waitClose()

	// provjeri da dobije poruke samo iz  topica 1 i 2
	assert.Len(t, c.messages, 3)
	// if c.messages[0] == m10 {
	// 	assert.Equal(t, m10, c.messages[0])
	// 	assert.Equal(t, m11, c.messages[1])
	// 	assert.Equal(t, m20, c.messages[2])
	// } else {
	// 	assert.Equal(t, m20, c.messages[0])
	// 	assert.Equal(t, m10, c.messages[1])
	// 	assert.Equal(t, m11, c.messages[2])
	// }
}

func TestSubscribe(t *testing.T) {
	s := New()
	c := &testConsumer{topics: map[string]int64{"1": 0, "2": 0}}

	assert.Len(t, s.topics, 0)
	s.Subscribe(c, c.topics)
	s.inLoop(func() {})

	// dodaj jedan
	c.topics["3"] = 0
	s.Subscribe(c, c.topics)
	s.inLoop(func() {})
	assert.Len(t, s.topics, 3)

	// izbaci 3 dodaj jedan
	c.topics = map[string]int64{"4": 0}
	s.Subscribe(c, c.topics)
	s.inLoop(func() {})
	assert.Len(t, s.topics, 1)
}

func TestDobijeFullNakonSubscribe(t *testing.T) {
	s := New()
	m10 := &amp.Msg{Topic: "1", Ts: 1, UpdateType: amp.Full}
	m11 := &amp.Msg{Topic: "1", Ts: 2, UpdateType: amp.Diff}
	s.Publish(m10)
	s.Publish(m11)
	s.wait("1")

	c2 := &testConsumer{topics: map[string]int64{"1": 0, "2": 0}}
	c3 := &testConsumer{topics: map[string]int64{"1": 2, "2": 0}}
	c4 := &testConsumer{topics: map[string]int64{"1": 3, "2": 0}}
	s.Subscribe(c2, c2.topics)
	s.Subscribe(c3, c3.topics)
	s.Subscribe(c4, c4.topics)
	m13 := &amp.Msg{Topic: "1", Ts: 3, UpdateType: amp.Diff}
	s.Publish(m13)
	s.waitClose()

	// c2 je dobio full i sve nakon
	assert.Len(t, c2.messages, 3)
	assert.Equal(t, m10, c2.messages[0])
	assert.Equal(t, m11, c2.messages[1])
	assert.Equal(t, m13, c2.messages[2])

	// c3 je dobio samo diff jer je vec bio na no 2 u trenutku subscribe
	assert.Len(t, c3.messages, 1)
	assert.Equal(t, m13, c3.messages[0])

	// c4 je dobio sve jer je bio van range-a u trenutku subscribe
	assert.Len(t, c4.messages, 3)
}

func TestSubscribeNaPrazanTopic(t *testing.T) {
	s := New()
	c := &testConsumer{topics: map[string]int64{"1": 100, "2": 0}}
	s.Subscribe(c, c.topics)
	c2 := &testConsumer{topics: map[string]int64{"1": 101, "2": 0}}
	s.Subscribe(c2, c2.topics)

	m1 := &amp.Msg{Topic: "1", Ts: 101, UpdateType: amp.Diff}
	s.Publish(m1)
	s.wait("1")

	assert.Len(t, c.messages, 1, "dobije full")
	assert.Len(t, c2.messages, 0, "vec ima ne dobije nista")
	assert.Equal(t, c.messages[0], m1)
}

func TestDobijePropusteneDiffOveNaSubscribe(t *testing.T) {
	s := New()
	c0 := &testConsumer{topics: map[string]int64{"1": 0, "2": 0}}
	s.Subscribe(c0, c0.topics)

	m1 := &amp.Msg{Topic: "1", Ts: 101, UpdateType: amp.Full}
	m2 := &amp.Msg{Topic: "1", Ts: 105, UpdateType: amp.Diff}
	m3 := &amp.Msg{Topic: "1", Ts: 107, UpdateType: amp.Diff}
	m4 := &amp.Msg{Topic: "1", Ts: 111, UpdateType: amp.Diff}
	s.Publish(m1)
	s.Publish(m2)
	s.Publish(m3)
	s.Publish(m4)
	s.wait("1")

	// spoji se consumer koji je zadnje uspjesno dobio 105
	// dobija slijedeca dva diff-a
	c := &testConsumer{topics: map[string]int64{"1": 105, "2": 0}}
	s.Subscribe(c, c.topics)
	s.wait("1")
	assert.Len(t, c.messages, 2)
	assert.Equal(t, m3, c.messages[0])
	assert.Equal(t, m4, c.messages[1])

	c = &testConsumer{topics: map[string]int64{"1": 101, "2": 0}}
	s.Subscribe(c, c.topics)
	s.wait("1")
	assert.Len(t, c.messages, 3)

	// van ranga dobije sve
	c = &testConsumer{topics: map[string]int64{"1": 1000, "2": 0}}
	s.Subscribe(c, c.topics)
	s.wait("1")
	assert.Len(t, c.messages, 4)
}

func TestReplay(t *testing.T) {
	s := New()
	m1 := &amp.Msg{Topic: "1", Ts: 101, UpdateType: amp.Full}
	m2 := &amp.Msg{Topic: "1", Ts: 105, UpdateType: amp.Diff}
	m3 := &amp.Msg{Topic: "1", Ts: 107, UpdateType: amp.Diff}
	m4 := &amp.Msg{Topic: "1", Ts: 111, UpdateType: amp.Diff}
	m21 := &amp.Msg{Topic: "2", Ts: 110, UpdateType: amp.Full}
	m22 := &amp.Msg{Topic: "2", Ts: 109, UpdateType: amp.Diff} // manji ts od full
	m23 := &amp.Msg{Topic: "2", Ts: 111, UpdateType: amp.Diff}
	s.Publish(m1)
	s.Publish(m2)
	s.Publish(m3)
	s.Publish(m4)
	s.Publish(m21)
	s.Publish(m22)
	s.Publish(m23)
	s.wait("1")

	msgs := s.Replay("1")
	assert.Len(t, msgs, 4)
	assert.Equal(t, m1, msgs[0])
	assert.Equal(t, m2, msgs[1])
	assert.Equal(t, m3, msgs[2])
	assert.Equal(t, m4, msgs[3])

	msgs = s.Replay("")
	assert.Len(t, msgs, 6)
}
