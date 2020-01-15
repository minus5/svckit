package broker

import (
	"fmt"
	"math"
	"strings"
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
	messages        chan *amp.Msg
	loopWork        chan func()
	consumers       map[amp.Subscriber]int64
	closed          chan struct{}
	cache           cache
	updatedAt       time.Time
	metricName      string
	mOnMsgDuration  string
	mOnMsgConsumers string
	mOnMsgMsgCount  string
	mOnMsgPerMsg    string
	mSubWait        string
	mSubDuration    string
	mSubMsgCount    string
	mSubPerMsg      string
}

func newTopic(name string) *topic {
	t := &topic{
		messages:   make(chan *amp.Msg, 128),
		consumers:  make(map[amp.Subscriber]int64),
		closed:     make(chan struct{}),
		loopWork:   make(chan func()),
		metricName: "other",
	}
	if strings.HasPrefix(name, "sportsbook/") {
		t.metricName = name[11:12]
	}
	t.mOnMsgDuration = fmt.Sprintf("topic.onMessage.%s.duration", t.metricName)
	t.mOnMsgConsumers = fmt.Sprintf("topic.onMessage.%s.consumers", t.metricName)
	t.mOnMsgMsgCount = fmt.Sprintf("topic.onMessage.%s.msgCount", t.metricName)
	t.mOnMsgPerMsg = fmt.Sprintf("topic.onMessage.%s.perMsg", t.metricName)
	t.mSubWait = fmt.Sprintf("topic.sub.%s.wait", t.metricName)
	t.mSubDuration = fmt.Sprintf("topic.sub.%s.duration", t.metricName)
	t.mSubMsgCount = fmt.Sprintf("topic.sub.%s.msgCount", t.metricName)
	t.mSubPerMsg = fmt.Sprintf("topic.sub.%s.perMsg", t.metricName)
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
	enter := time.Now()
	defer func() {
		metric.Time("topic.close", int(time.Now().Sub(enter).Nanoseconds()))
	}()
	close(t.messages)
	<-t.closed
}

func (t *topic) subscribe(c amp.Subscriber, ts int64) {
	call := time.Now()
	t.loopWork <- func() {
		enter := time.Now()
		msgCount := 0
		defer func() {
			if msgCount == 0 {
				return
			}
			duration := int(time.Now().Sub(enter).Nanoseconds())
			metric.Time(t.mSubWait, int(enter.Sub(call).Nanoseconds()))
			metric.Time(t.mSubDuration, duration)
			metric.Time(t.mSubMsgCount, msgCount)
			metric.Time(t.mSubPerMsg, duration/msgCount)
		}()
		if ts <= 0 {
			ts = tsNone
		}
		t.consumers[c] = ts
		if t.cache != nil {
			msgs := t.cache.Find(ts)
			t.sendMany(c, msgs)
			msgCount = len(msgs)
		}
	}
}

// unsubscribe vraca true ako vise nema niti jednog consumera.
func (t *topic) unsubscribe(c amp.Subscriber) bool {
	empty := make(chan bool)
	call := time.Now()
	t.loopWork <- func() {
		enter := time.Now()
		metric.Time("topic.unsubscribe.wait", int(enter.Sub(call).Nanoseconds()))
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
	start := time.Now()
	msgCount := 0
	defer func() {
		if msgCount == 0 {
			return
		}
		duration := int(time.Now().Sub(start).Nanoseconds())
		metric.Time(t.mOnMsgDuration, duration)
		metric.Time(t.mOnMsgConsumers, len(t.consumers))
		metric.Time(t.mOnMsgMsgCount, msgCount)
		metric.Time(t.mOnMsgPerMsg, duration/msgCount)
	}()
	if m.UpdateType == amp.Event {
		for c := range t.consumers {
			t.send(c, m)
		}
		return
	}

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
			msgCount++
		case sendCurrent:
			current := t.cache.Current()
			t.sendMany(c, current)
			msgCount += len(current)
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
