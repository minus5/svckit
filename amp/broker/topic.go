package broker

import (
	"math"
	"sort"
	"time"

	"github.com/minus5/svckit/amp"
)

var (
	tsNone = int64(math.MinInt64)
)

type topic struct {
	messages  chan *amp.Msg
	loopWork  chan func()
	consumers map[amp.Subscriber]int64
	closed    chan struct{}
	full      *amp.Msg   // last full message
	diffs     []*amp.Msg // previous diff messages
	updatedAt time.Time
}

func newTopic() *topic {
	t := &topic{
		messages:  make(chan *amp.Msg, 128),
		consumers: make(map[amp.Subscriber]int64),
		closed:    make(chan struct{}),
		loopWork:  make(chan func()),
		diffs:     make([]*amp.Msg, 0),
	}
	go t.loop()
	return t
}

func (t *topic) publish(m *amp.Msg) {
	t.messages <- m
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
		t.sendMany(c, t.findForSubscribe(ts))
	}
}

// unsubscribe vraca true ako vise nema niti jednog consumera.
func (t *topic) unsubscribe(c amp.Subscriber) bool {
	empty := make(chan bool)
	t.loopWork <- func() {
		delete(t.consumers, c)
		empty <- len(t.consumers) == 0
	}
	return <-empty
}

func (t *topic) sendMany(c amp.Subscriber, msgs []*amp.Msg) {
	for _, m := range msgs {
		t.send(c, m)
	}
}

func (t *topic) send(c amp.Subscriber, m *amp.Msg) {
	t.consumers[c] = m.Ts
	c.Send(m)
}

// message for subsribers after he subscribes with ts
func (t *topic) findForSubscribe(ts int64) []*amp.Msg {
	if len(t.diffs) > 0 && ts >= t.diffs[0].Ts && ts <= t.diffs[len(t.diffs)-1].Ts {
		return t.diffsAfter(ts)
	}
	if t.full == nil {
		return nil
	}
	return t.current()
}

// updateCache adds new message to the caches t.full or t.diffs
func (t *topic) updateCache(m *amp.Msg) {
	if m.IsFull() {
		if m.IsReplay() && t.full != nil {
			return
		}
		if t.full != nil { // preserve all after previous full
			t.compactDiffs(t.full.Ts)
		}
		t.full = m
		return
	}

	t.diffs = append(t.diffs, m)
	if len(t.diffs) > 1 {
		prev := len(t.diffs) - 2
		if m.Ts <= t.diffs[prev].Ts {
			t.sortDiffs()
		}
	}
}

// compactDiffs preserves only diffs with Ts greater than input ts
func (t *topic) compactDiffs(ts int64) {
	var n []*amp.Msg
	for _, m := range t.diffs {
		if m.Ts >= ts {
			n = append(n, m)
		}
	}
	t.diffs = n
}

// sortDiffs sorts and removes duplicates in t.diffs
func (t *topic) sortDiffs() {
	sort.Slice(t.diffs, func(i, j int) bool {
		return t.diffs[i].Ts < t.diffs[j].Ts
	})
	// remove duplicates
	for i := 0; i < len(t.diffs)-1; i++ {
		m1 := t.diffs[i]
		m2 := t.diffs[i+1]
		if m1.Ts == m2.Ts {
			if m1.IsReplay() {
				t.diffs = append(t.diffs[:i], t.diffs[i+1:]...) //remove i
				continue
			}
			j := i + 1
			t.diffs = append(t.diffs[:j], t.diffs[j+1:]...) //remove i+1
		}
	}
}

func (t *topic) diffsAfter(ts int64) []*amp.Msg {
	var d []*amp.Msg
	for _, m := range t.diffs {
		if m.Ts > ts {
			d = append(d, m)
		}
	}
	return d
}

func (t *topic) current() []*amp.Msg {
	return append([]*amp.Msg{t.full}, t.diffsAfter(t.full.Ts)...)
}

func (t *topic) onMessage(m *amp.Msg) {
	t.updateCache(m)
	t.sendToConsumers(m)
	t.updatedAt = time.Now()
}

func (t *topic) sendToConsumers(m *amp.Msg) {
	if m.IsFull() {
		var current []*amp.Msg
		for c, cNo := range t.consumers {
			if cNo != tsNone {
				continue
			}
			if current == nil {
				current = t.current()
			}
			t.sendMany(c, current)
		}
		return
	}
	for c, cTs := range t.consumers {
		if cTs == m.Ts || cTs == tsNone {
			continue // ovaj consumer je vec dobio ovu ili jos nije dobio full
		}
		if m.IsReplay() && cTs >= m.Ts { // nemoj ponavljati replay poruke onima koji ih vec imaju
			continue
		}
		t.send(c, m)
	}
}

func (t *topic) replay() []*amp.Msg {
	ret := make(chan []*amp.Msg, 1)
	t.loopWork <- func() {
		ret <- t.current()
	}
	msgs := <-ret
	var rmsgs []*amp.Msg
	for _, m := range msgs {
		rmsgs = append(rmsgs, m.AsReplay())
	}
	return rmsgs
}

func (t *topic) metrics() (diffs, firstDiffTs, lastDiffTs, fullTs int64) {
	done := make(chan struct{})
	t.loopWork <- func() {
		diffs = int64(len(t.diffs))
		if len(t.diffs) > 0 {
			firstDiffTs = t.diffs[0].Ts
			lastDiffTs = t.diffs[len(t.diffs)-1].Ts
		}
		if t.full != nil {
			fullTs = t.full.Ts
		}
		close(done)
	}
	<-done
	return
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
