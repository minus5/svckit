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
	consumers       map[amp.Sender]int64
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
		consumers:  make(map[amp.Sender]int64),
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

func (t *topic) subscribe(c amp.Sender, ts int64) {
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
			ms := t.cache.Find(ts)
			msgCount = len(ms)
			if msgCount > 0 {
				t.send(c, burst(ms))
			}
		}
	}
}

func (t *topic) unsubscribe(c amp.Sender) {
	call := time.Now()
	t.loopWork <- func() {
		metric.Time("topic.unsubscribe.wait", int(time.Now().Sub(call).Nanoseconds()))
		delete(t.consumers, c)
	}
}

func burst(ms []*amp.Msg) []*amp.Msg {
	l := len(ms)
	if l <= 2 {
		return ms
	}
	res := []*amp.Msg{
		ms[0].BurstStart(),
	}
	res = append(res, ms...)
	res = append(res, ms[l-1].BurstEnd())
	return res
}

func (t *topic) send(c amp.Sender, ms []*amp.Msg) {
	t.consumers[c] = ms[len(ms)-1].Ts
	c.SendMsgs(ms)
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
	ms := []*amp.Msg{m}
	if m.UpdateType == amp.Event {
		for c := range t.consumers {
			t.send(c, ms)
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
	var current []*amp.Msg
	for c, cTs := range t.consumers {
		switch t.cache.FindFor(cTs, m) {
		case sendMsg:
			t.send(c, ms)
			msgCount++
		case sendCurrent:
			if current == nil {
				current = burst(t.cache.Current())
			}
			t.send(c, current)
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
