package broker

import (
	"math"
	"sort"
	"time"

	"github.com/mnu5/svckit/amp"
)

var (
	tsNone = int64(math.MinInt64)
)

type topic struct {
	messages  chan *amp.Msg
	loopWork  chan func()
	consumers map[amp.Subscriber]int64
	closed    chan struct{}
	prev      []*amp.Msg // previous messages, full and all diffs after that full
	updatedAt time.Time
}

func newTopic() *topic {
	t := &topic{
		messages:  make(chan *amp.Msg, 128),
		consumers: make(map[amp.Subscriber]int64),
		closed:    make(chan struct{}),
		loopWork:  make(chan func()),
		prev:      make([]*amp.Msg, 0),
	}
	go t.loop()
	return t
}

func (t *topic) publish(m *amp.Msg) {
	t.messages <- m
}

func (t *topic) close() {
	close(t.messages)
	<-t.closed
}

func (t *topic) subscribe(c amp.Subscriber, ts int64) {
	t.loopWork <- func() {
		if ts <= 0 {
			ts = tsNone
		}
		t.consumers[c] = ts
		msgs := t.findForSubscribe(ts)
		for _, m := range msgs {
			t.send(c, m)
		}
	}
}

func (t *topic) send(c amp.Subscriber, m *amp.Msg) {
	t.consumers[c] = m.Ts
	c.Send(m)
}

func (t *topic) prevTs() (int64, int64) {
	lastFull := tsNone
	lastMsg := tsNone
	if len(t.prev) > 0 {
		lastFull = t.prev[0].Ts
		lastMsg = t.prev[len(t.prev)-1].Ts
	}
	return lastFull, lastMsg
}

// poruke koje dobije kada napravi subscribe
func (t *topic) findForSubscribe(ts int64) []*amp.Msg {
	lastFull, lastMsg := t.prevTs()
	if ts < lastFull {
		return t.prev
	}
	if ts > lastMsg && lastMsg != tsNone {
		return t.prev
	}

	var msgs []*amp.Msg
	for _, m := range t.prev {
		if m.Ts > ts {
			msgs = append(msgs, m)
		}
	}
	return msgs

}

// Unsubscribe vraca true ako vise nema niti jednog consumera.
func (t *topic) unsubscribe(c amp.Subscriber) bool {
	empty := make(chan bool)
	t.loopWork <- func() {
		delete(t.consumers, c)
		empty <- len(t.consumers) == 0
	}
	return <-empty
}

func (t *topic) loop() {
	for {
		select {
		case m, ok := <-t.messages:
			if !ok {
				close(t.closed)
				return
			}
			t.onMessage(m)
		case f := <-t.loopWork:
			f()
		}
	}
}

func (t *topic) sortPrev() {
	sort.Slice(t.prev, func(i, j int) bool {
		return t.prev[i].Ts < t.prev[j].Ts
	})
}

func (t *topic) onMessage(m *amp.Msg) {
	ts := m.Ts
	if m.IsFull() {
		// sacuvaj full i sve diffove nakon njega
		p := make([]*amp.Msg, 0)
		p = append(p, m)
		for _, pm := range t.prev {
			if pm.Ts > m.Ts {
				p = append(p, pm)
			}
		}
		t.prev = p
		t.sortPrev()
		// posalji full svima koji jos nisu nista dobili
		for c, cNo := range t.consumers {
			if cNo == tsNone {
				t.send(c, m)
			}
		}
		return
	}
	for c, cTs := range t.consumers {
		if cTs == ts || cTs == tsNone {
			continue // ovaj consumer je vec dobio ovu ili jos nije dobio full
		}
		t.send(c, m)
	}
	if len(t.prev) == 0 {
		return
	}
	lastFull, lastMsg := t.prevTs()
	if m.Ts > lastFull {
		t.prev = append(t.prev, m)
		if m.Ts < lastMsg {
			t.sortPrev()
		}
	}
	t.updatedAt = time.Now()
}

func (t *topic) replay() []*amp.Msg {
	ret := make(chan []*amp.Msg, 1)
	t.loopWork <- func() {
		var msgs []*amp.Msg
		for _, m := range t.prev {
			msgs = append(msgs, m)
		}
		ret <- msgs
	}
	return <-ret
}

// samo za testove
func (t *topic) wait() {
	for {
		ch := make(chan int)
		t.loopWork <- func() {
			ch <- len(t.messages)
		}
		if 0 == <-ch {
			return
		}
	}
}
