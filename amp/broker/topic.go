package broker

import (
	"math"
	"time"

	"github.com/minus5/svckit/amp"
)

const (
	sendNothing uint8 = iota
	sendMsg
	sendCurrent
)

var (
	tsNone = int64(math.MinInt64)
)

type cache interface {
	Add(m *amp.Msg)
	Find(ts int64) []*amp.Msg
	FindFor(consumerTs int64, m *amp.Msg) uint8
	Current() []*amp.Msg
}

type topic struct {
	messages  chan *amp.Msg
	loopWork  chan func()
	consumers map[amp.Subscriber]int64
	closed    chan struct{}
	cache     cache
	updatedAt time.Time
}

func newTopic() *topic {
	t := &topic{
		messages:  make(chan *amp.Msg, 128),
		consumers: make(map[amp.Subscriber]int64),
		closed:    make(chan struct{}),
		loopWork:  make(chan func()),
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
		if t.cache != nil {
			t.sendMany(c, t.cache.Find(ts))
		}
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
	burstStartEnd := len(msgs) > 2
	if burstStartEnd {
		t.send(c, msgs[0].BurstStart())
	}
	for _, m := range msgs {
		t.send(c, m)
	}
	if burstStartEnd {
		t.send(c, msgs[len(msgs)-1].BurstEnd())
	}
}

func (t *topic) send(c amp.Subscriber, m *amp.Msg) {
	t.consumers[c] = m.Ts
	c.Send(m)
}

func (t *topic) onMessage(m *amp.Msg) {
	if t.cache == nil {
		if m.UpdateType == amp.Append || m.UpdateType == amp.Update {
			t.cache = newAppendCache()
		} else {
			t.cache = newFullDiffCache()
		}
	}

	t.cache.Add(m)
	for c, cTs := range t.consumers {
		switch t.cache.FindFor(cTs, m) {
		case sendMsg:
			t.send(c, m)
		case sendCurrent:
			t.sendMany(c, t.cache.Current())
		}
	}

	t.updatedAt = time.Now()
}

func (t *topic) replay() []*amp.Msg {
	if t.cache == nil {
		return nil
	}
	ret := make(chan []*amp.Msg, 1)
	t.loopWork <- func() {
		ret <- t.cache.Current()
	}
	msgs := <-ret
	var rmsgs []*amp.Msg
	for _, m := range msgs {
		rmsgs = append(rmsgs, m.AsReplay())
	}
	return rmsgs
}

// func (t *topic) metrics() (diffs, firstDiffTs, lastDiffTs, fullTs int64) {
// 	done := make(chan struct{})
// 	t.loopWork <- func() {
// 		diffs = int64(len(t.diffs))
// 		if len(t.diffs) > 0 {
// 			firstDiffTs = t.diffs[0].Ts
// 			lastDiffTs = t.diffs[len(t.diffs)-1].Ts
// 		}
// 		if t.full != nil {
// 			fullTs = t.full.Ts
// 		}
// 		close(done)
// 	}
// 	<-done
// 	return
// }

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
