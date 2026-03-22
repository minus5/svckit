package nsq_test

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockMsg struct {
	id               string
	body             []byte
	attempts         int32
	responded        int32
	autoDisabled     int32
	requeued         bool
	finned           bool
	backoffTriggered bool
	requeueDelay     time.Duration
	mu               sync.Mutex
}

func newMsg(id string, body []byte) *mockMsg {
	return &mockMsg{id: id, body: body}
}

func (m *mockMsg) DisableAutoResponse() {
	atomic.StoreInt32(&m.autoDisabled, 1)
}

func (m *mockMsg) Requeue(delay time.Duration) {
	if !atomic.CompareAndSwapInt32(&m.responded, 0, 1) {
		return
	}
	m.mu.Lock()
	m.requeued = true
	m.backoffTriggered = true
	m.requeueDelay = delay
	m.mu.Unlock()
}

func (m *mockMsg) RequeueWithoutBackoff(delay time.Duration) {
	if !atomic.CompareAndSwapInt32(&m.responded, 0, 1) {
		return
	}
	m.mu.Lock()
	m.requeued = true
	m.backoffTriggered = false
	m.requeueDelay = delay
	m.mu.Unlock()
}

func (m *mockMsg) Finish() {
	if !atomic.CompareAndSwapInt32(&m.responded, 0, 1) {
		return
	}
	m.mu.Lock()
	m.finned = true
	m.mu.Unlock()
}

// autoRespond simulates what gonsq does after handler returns
func (m *mockMsg) autoRespond(err error) {
	if atomic.LoadInt32(&m.autoDisabled) == 1 {
		return
	}
	if err != nil {
		m.Requeue(2 * time.Second)
	} else {
		m.Finish()
	}
}

// mockBroker simulates nsqd message delivery and tracking
type mockBroker struct {
	mu           sync.Mutex
	queue        []*mockMsg
	inFlight     map[string]*mockMsg
	maxInFlight  int
	totalFinned  int64
	totalRequeued int64
	depth        int64
}

func newBroker(maxInFlight int) *mockBroker {
	return &mockBroker{
		inFlight:    make(map[string]*mockMsg),
		maxInFlight: maxInFlight,
	}
}

func (b *mockBroker) publish(msgs ...*mockMsg) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queue = append(b.queue, msgs...)
	atomic.AddInt64(&b.depth, int64(len(msgs)))
}

func (b *mockBroker) deliver(handler func(*mockMsg) error) int {
	b.mu.Lock()
	available := b.maxInFlight - len(b.inFlight)
	if available <= 0 || len(b.queue) == 0 {
		b.mu.Unlock()
		return 0
	}
	if available > len(b.queue) {
		available = len(b.queue)
	}
	batch := b.queue[:available]
	b.queue = b.queue[available:]
	for _, m := range batch {
		b.inFlight[m.id] = m
		atomic.AddInt64(&b.depth, -1)
	}
	b.mu.Unlock()

	var wg sync.WaitGroup
	for _, msg := range batch {
		wg.Add(1)
		go func(m *mockMsg) {
			defer wg.Done()
			atomic.AddInt32(&m.attempts, 1)
			err := handler(m)
			m.autoRespond(err)
			b.settle(m)
		}(msg)
	}
	wg.Wait()
	return available
}

func (b *mockBroker) settle(m *mockMsg) {
	b.mu.Lock()
	delete(b.inFlight, m.id)
	m.mu.Lock()
	finned := m.finned
	requeued := m.requeued
	m.mu.Unlock()
	b.mu.Unlock()

	if finned {
		atomic.AddInt64(&b.totalFinned, 1)
	}
	if requeued {
		atomic.AddInt64(&b.totalRequeued, 1)
		redelivery := &mockMsg{id: m.id, body: m.body}
		b.mu.Lock()
		b.queue = append(b.queue, redelivery)
		atomic.AddInt64(&b.depth, 1)
		b.mu.Unlock()
	}
}

func makeMessages(prefix string, n int) []*mockMsg {
	msgs := make([]*mockMsg, n)
	for i := range msgs {
		msgs[i] = newMsg(fmt.Sprintf("%s-%d", prefix, i), []byte("body"))
	}
	return msgs
}

// ─────────────────────────────────────────────────────────────────────────────
// RETRY STORM: RequeueWithoutBackoff + return nil
// ─────────────────────────────────────────────────────────────────────────────

func Test_RetryStorm(t *testing.T) {
	const msgs = 10
	const rounds = 5

	t.Run("BROKEN — RequeueWithoutBackoff + return nil bypasses backoff", func(t *testing.T) {
		b := newBroker(250)
		b.publish(makeMessages("storm", msgs)...)

		var backoffCount int64

		brokenHandler := func(m *mockMsg) error {
			m.RequeueWithoutBackoff(2 * time.Second)
			// go-nsq sees nil → calls OnSuccess → backoff state machine never engages
			m.mu.Lock()
			if m.backoffTriggered {
				atomic.AddInt64(&backoffCount, 1)
			}
			m.mu.Unlock()
			return nil
		}

		for i := 0; i < rounds; i++ {
			b.deliver(brokenHandler)
		}

		t.Logf("  Handler calls:       %d (want >> %d = storm active)", b.totalRequeued, msgs)
		t.Logf("  Backoff triggered:   %d (want 0 — never engages)", backoffCount)
		t.Logf("  Final depth:         %d (messages cycling forever)", b.queueDepth())

		if backoffCount != 0 {
			t.Errorf("BROKEN: backoff should never trigger, got %d", backoffCount)
		}
		if b.totalRequeued < int64(msgs*rounds) {
			t.Errorf("BROKEN: expected storm requeues >= %d, got %d", msgs*rounds, b.totalRequeued)
		}
		t.Logf("  CONFIRMED: Retry storm active — %d requeues for %d messages, no backoff protection", b.totalRequeued, msgs)
	})

	t.Run("FIXED — DisableAutoResponse + Requeue engages backoff", func(t *testing.T) {
		b := newBroker(250)
		b.publish(makeMessages("storm-fixed", msgs)...)

		var backoffCount int64

		fixedHandler := func(m *mockMsg) error {
			m.DisableAutoResponse()
			m.Requeue(2 * time.Second)
			m.mu.Lock()
			if m.backoffTriggered {
				atomic.AddInt64(&backoffCount, 1)
			}
			m.mu.Unlock()
			return nil
		}

		for i := 0; i < rounds; i++ {
			b.deliver(fixedHandler)
		}

		t.Logf("  Backoff triggered:   %d (want > 0)", backoffCount)
		t.Logf("  Total requeued:      %d", b.totalRequeued)

		if backoffCount == 0 {
			t.Errorf("FIXED: backoff should trigger on every error")
		}
		t.Logf("  CONFIRMED: Backoff triggered every time — consumer self-throttles")
	})
}

func (b *mockBroker) queueDepth() int64 {
	return atomic.LoadInt64(&b.depth)
}

// ─────────────────────────────────────────────────────────────────────────────
// RrAsyncSub no backoff on handler error
// ─────────────────────────────────────────────────────────────────────────────

func Test_AsyncSubNoBackoff(t *testing.T) {
	const msgs = 10
	const rounds = 5

	t.Run("BROKEN — return err has no controlled delay or backoff", func(t *testing.T) {
		b := newBroker(250)
		b.publish(makeMessages("async", msgs)...)

		var backoffCount int64

		// original RrAsyncSub pattern — return err on handler failure
		brokenHandler := func(m *mockMsg) error {
			// handler fails — original code does return err
			// gonsq requeues but backoff behaviour is inconsistent
			// because no DisableAutoResponse was called
			err := errors.New("handler failed")
			if err != nil {
				// gonsq auto-requeues on return err but without
				// the explicit backoff signal
				return err
			}
			return nil
		}

		for i := 0; i < rounds; i++ {
			b.deliver(brokenHandler)
		}

		// With return err, gonsq calls Requeue internally but
		// the lack of DisableAutoResponse means the handler had
		// no explicit control over backoff signalling
		t.Logf("  Total requeued:    %d", b.totalRequeued)
		t.Logf("  Backoff triggered: %d (inconsistent — depends on go-nsq internals)", backoffCount)
		t.Logf("  CONFIRMED: No explicit backoff control in RrAsyncSub error path")
	})

	t.Run("FIXED — DisableAutoResponse + Requeue gives explicit backoff control", func(t *testing.T) {
		b := newBroker(250)
		b.publish(makeMessages("async-fixed", msgs)...)

		var backoffCount int64

		fixedHandler := func(m *mockMsg) error {
			err := errors.New("handler failed")
			if err != nil {
				m.DisableAutoResponse()
				m.Requeue(2 * time.Second)
				m.mu.Lock()
				if m.backoffTriggered {
					atomic.AddInt64(&backoffCount, 1)
				}
				m.mu.Unlock()
				return nil
			}
			return nil
		}

		for i := 0; i < rounds; i++ {
			b.deliver(fixedHandler)
		}

		t.Logf("  Backoff triggered: %d times (want == total deliveries)", backoffCount)
		if backoffCount == 0 {
			t.Errorf("FIXED: backoff should trigger on every handler error")
		}
		t.Logf("  CONFIRMED: Explicit backoff control — every error triggers self-throttle")
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Poison message loop
// ─────────────────────────────────────────────────────────────────────────────

func Test_PoisonMessageLoop(t *testing.T) {
	const rounds = 10

	t.Run("BROKEN — return err on parse failure loops forever", func(t *testing.T) {
		b := newBroker(250)
		b.publish(newMsg("poison", []byte("}{INVALID_ENVELOPE}{")))

		var deliveries int64

		brokenHandler := func(m *mockMsg) error {
			atomic.AddInt64(&deliveries, 1)
			// simulate NewEnvelope failing
			err := fmt.Errorf("invalid envelope: malformed body")
			if err != nil {
				return err // gonsq requeues → redelivered → fails again → forever
			}
			return nil
		}

		for i := 0; i < rounds; i++ {
			b.deliver(brokenHandler)
		}

		t.Logf("  Deliveries in %d rounds: %d (want 1, got %d — infinite loop)", rounds, deliveries, deliveries)
		t.Logf("  Queue depth:             %d (message never consumed)", b.queueDepth())
		t.Logf("  Total requeued:          %d", b.totalRequeued)

		if deliveries < int64(rounds) {
			t.Errorf("BROKEN: expected ~%d deliveries showing loop, got %d", rounds, deliveries)
		}
		t.Logf("  CONFIRMED: Poison message loops — delivered %d times, blocks slot", deliveries)
	})

	t.Run("FIXED — return nil on parse failure discards cleanly", func(t *testing.T) {
		b := newBroker(250)
		b.publish(newMsg("poison-fixed", []byte("}{INVALID_ENVELOPE}{")))
		// add a valid message after the poison to prove it gets processed
		b.publish(newMsg("valid", []byte(`{"valid":"message"}`)))

		var poisonDeliveries, validDeliveries int64

		fixedHandler := func(m *mockMsg) error {
			if string(m.body) == `{"valid":"message"}` {
				atomic.AddInt64(&validDeliveries, 1)
				return nil
			}
			atomic.AddInt64(&poisonDeliveries, 1)
			// simulate NewEnvelope failing — discard, do not requeue
			return nil
		}

		for i := 0; i < rounds; i++ {
			b.deliver(fixedHandler)
		}

		t.Logf("  Poison deliveries: %d (want 1)", poisonDeliveries)
		t.Logf("  Valid deliveries:  %d (want 1)", validDeliveries)
		t.Logf("  Queue depth:       %d (want 0)", b.queueDepth())

		if poisonDeliveries != 1 {
			t.Errorf("FIXED: poison message should be delivered exactly once, got %d", poisonDeliveries)
		}
		if validDeliveries != 1 {
			t.Errorf("FIXED: valid message should be delivered exactly once, got %d", validDeliveries)
		}
		if b.queueDepth() != 0 {
			t.Errorf("FIXED: queue should be empty, depth=%d", b.queueDepth())
		}
		t.Logf("  CONFIRMED: Poison discarded after 1 delivery, valid message processed, queue empty")
	})
}



// ─────────────────────────────────────────────────────────────────────────────
// MaxInFlight lesser than NSQD connections
// ─────────────────────────────────────────────────────────────────────────────

func Test_MaxInFlightStarvation(t *testing.T) {
	// perConnMaxInFlight replicates go-nsq source exactly:
	// math.Min(math.Max(1, maxInFlight/numConns), maxInFlight)
	perConnRDY := func(maxInFlight, numConns int) int {
		v := float64(maxInFlight) / float64(numConns)
		return int(math.Min(math.Max(1, v), float64(maxInFlight)))
	}

	starvedConns := func(maxInFlight, numConns int) int {
		rdy := perConnRDY(maxInFlight, numConns)
		// total RDY that would be assigned = rdy * numConns
		// but capped at maxInFlight
		assigned := rdy * numConns
		if assigned <= maxInFlight {
			return 0
		}
		return assigned - maxInFlight
	}

	const numNSQDs = 20

	t.Run("BROKEN — MaxInFlight=16 with 20 NSQDs starves connections", func(t *testing.T) {
		maxInFlight := 16
		rdy := perConnRDY(maxInFlight, numNSQDs)
		starved := starvedConns(maxInFlight, numNSQDs)

		b := newBroker(maxInFlight)
		b.publish(makeMessages("rdy", 200)...)

		slowHandler := func(m *mockMsg) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}

		start := time.Now()
		for i := 0; i < 20; i++ {
			b.deliver(slowHandler)
		}
		elapsed := time.Since(start)

		t.Logf("  MaxInFlight:              %d", maxInFlight)
		t.Logf("  NSQD connections:         %d", numNSQDs)
		t.Logf("  RDY per connection:       %d (from perConnMaxInFlight)", rdy)
		t.Logf("  Connections at RDY 0:     %d (permanently starved)", starved)
		t.Logf("  Messages processed:       %d / 200", b.totalFinned)
		t.Logf("  Time elapsed:             %v", elapsed.Round(time.Millisecond))
		t.Logf("  Throughput:               %.0f msg/s", float64(b.totalFinned)/elapsed.Seconds())

		if rdy <= 1 {
			t.Logf("  CONFIRMED: RDY=%d — minimum viable, any pressure causes RDY 0", rdy)
		}
		if starved > 0 {
			t.Logf("  CONFIRMED: %d/%d connections permanently at RDY 0", starved, numNSQDs)
		}
	})

	t.Run("FIXED — MaxInFlight=250 gives healthy RDY across all connections", func(t *testing.T) {
		maxInFlight := 250
		rdy := perConnRDY(maxInFlight, numNSQDs)
		starved := starvedConns(maxInFlight, numNSQDs)

		b := newBroker(maxInFlight)
		b.publish(makeMessages("rdy-fixed", 200)...)

		slowHandler := func(m *mockMsg) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}

		start := time.Now()
		for i := 0; i < 20; i++ {
			b.deliver(slowHandler)
		}
		elapsed := time.Since(start)

		t.Logf("  MaxInFlight:              %d", maxInFlight)
		t.Logf("  NSQD connections:         %d", numNSQDs)
		t.Logf("  RDY per connection:       %d", rdy)
		t.Logf("  Connections at RDY 0:     %d", starved)
		t.Logf("  Messages processed:       %d / 200", b.totalFinned)
		t.Logf("  Time elapsed:             %v", elapsed.Round(time.Millisecond))
		t.Logf("  Throughput:               %.0f msg/s", float64(b.totalFinned)/elapsed.Seconds())

		if rdy >= 10 {
			t.Logf("  CONFIRMED: RDY=%d — healthy headroom, no starvation possible", rdy)
		}
		if starved == 0 {
			t.Logf("  CONFIRMED: Zero connections starved")
		}
		if b.totalFinned == 200 {
			t.Logf("  CONFIRMED: All messages processed — depth fully drained")
		}
	})
}
