package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/minus5/svckit/amp"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
)

var (
	maxWriteQueueDepth = 256              // max number of messages in outQueue
	aliveInterval      = 32 * time.Second // interval for sending alive message
	startupInterval    = 2 * time.Second  // grace period for having more then max messages in outQueue
)

type session struct {
	conn            connection      // client websocket connection
	broker          broker          // broker for subscribe on published messages
	requester       requester       // requester for request / response messages
	outQueue        []*amp.Msg      // output messages queue
	outQueueChanged chan (struct{}) // signal that queue changed
	stats           struct {        // sessions stats counters
		start         time.Time
		outMessages   int
		inMessages    int
		aliveMessages int
		maxQueueLen   int
	}
	compatibilityVersion uint8
	started              bool
	closed               bool
	sync.Mutex
}

// serve starts new session
// Blocks until session is finished.
func serve(cancelSig context.Context, conn connection, req requester, brk broker,
	compatibilityVersion uint8) {
	s := &session{
		conn:                 conn,
		requester:            req,
		broker:               brk,
		outQueue:             make([]*amp.Msg, 0),
		outQueueChanged:      make(chan struct{}),
		compatibilityVersion: compatibilityVersion,
	}
	s.stats.start = time.Now()
	s.loop(cancelSig)
}

func (s *session) loop(cancelSig context.Context) {
	inMessages := s.readLoop()            // messages from the client
	outMessages := make(chan *amp.Msg, 1) // messages to the client
	exitSig := cancelSig.Done()           // aplication exit signal

	// timer for alive messages
	alive := time.NewTimer(aliveInterval)
	// if there is no other messages send alive
	sendAlive := func() {
		s.Lock()
		defer s.Unlock()
		if len(s.outQueue) == 0 {
			s.outQueue = append(s.outQueue, amp.NewAlive())
			s.stats.aliveMessages++
		}
	}

	// if there is anything in queue waiting for sending put it inot outMessages chan
	tryPopQueue := func() {
		s.Lock()
		defer s.Unlock()
		if len(s.outQueue) > 0 {
			select { /// non blocking write
			case outMessages <- s.outQueue[0]:
				s.outQueue = s.outQueue[1:]
			default:
			}
		}
	}

	defer s.logStats()

	for {
		tryPopQueue()

		select {
		case <-s.outQueueChanged:
			// just start another loop iteration
		case <-alive.C:
			sendAlive()
		case msg := <-outMessages:
			s.connWrite(msg)
			alive.Reset(aliveInterval)
			s.stats.outMessages++
		case msg, ok := <-inMessages:
			if !ok {
				s.unsubscribe()
				return
			}
			s.receive(msg)
			s.stats.inMessages++
		case <-exitSig:
			_ = s.conn.Close()
			exitSig = nil // fire once
		}
	}
}

func (s *session) logStats() {
	s.Lock()
	defer s.Unlock()
	duration := int(time.Now().Sub(s.stats.start) / time.Millisecond)
	s.log().I("inMessages", s.stats.inMessages).
		I("outMessages", s.stats.outMessages).
		I("aliveMessages", s.stats.aliveMessages).
		I("durationMs", duration).
		Debug("stats")
	metric.Time("inMessages", s.stats.inMessages)
	metric.Time("outMessages", s.stats.outMessages)
	metric.Time("aliveMessages", s.stats.aliveMessages)
	metric.Time("duration", duration)
}

func (s *session) unsubscribe() {
	s.broker.Unsubscribe(s)
	s.requester.Unsubscribe(s)
}

func (s *session) readLoop() chan *amp.Msg {
	in := make(chan *amp.Msg)
	go func() {
		defer close(in)
		for {
			buf, err := s.conn.Read()
			if err != nil {
				return
			}
			if m := amp.ParseCompatibility(buf, s.compatibilityVersion); m != nil {
				in <- m
			}
		}
	}()
	return in
}

// receive gets client messages
func (s *session) receive(m *amp.Msg) {
	switch m.Type {
	case amp.Ping:
		s.Send(m.Pong())
	case amp.Request:
		// TODO what URI-a are ok, make filter
		m.Meta = s.conn.Meta()
		s.requester.Send(s, m)
	case amp.Subscribe:
		s.broker.Subscribe(s, m.Subscriptions)
	}
}

// Send message to the clinet
// Implements amp.Subscriber interface.
func (s *session) Send(m *amp.Msg) {
	// add to queue
	s.Lock()
	defer s.Unlock()

	if s.isStarted() {
		queueLen := len(s.outQueue)
		if s.stats.maxQueueLen < queueLen {
			s.stats.maxQueueLen = queueLen
		}
		// check for queue overflow
		if queueLen >= maxWriteQueueDepth {
			if !s.closed {
				s.conn.Close()
				s.logOutQueueOverflow()
			}
			s.closed = true
			return
		}
	}

	s.outQueue = append(s.outQueue, m)
	// signal queue changed
	select {
	case s.outQueueChanged <- struct{}{}:
	default:
	}
}

// should be called during s.Lock
func (s *session) isStarted() bool {
	if !s.started {
		s.started = time.Now().Sub(s.stats.start) > startupInterval
	}
	return s.started
}

// should be called during s.Lock
func (s *session) logOutQueueOverflow() {
	s.log().I("len", len(s.outQueue)).
		S("start", fmt.Sprintf("%v", s.stats.start)).
		I("inMessages", s.stats.inMessages).
		I("outMessages", s.stats.outMessages).
		I("aliveMessages", s.stats.aliveMessages).
		I("durationMs", int(time.Now().Sub(s.stats.start)/time.Millisecond)).
		Info("out queue overflow")
	for i, m := range s.outQueue {
		s.log().I("i", i).I("type", int(m.Type)).S("uri", m.URI).I("updateType", int(m.UpdateType)).Info("queue content")
	}
}

func (s *session) connWrite(m *amp.Msg) {
	var payload []byte
	deflated := false
	if s.conn.DeflateSupported() {
		payload, deflated = m.MarshalDeflateCompatiblity(s.compatibilityVersion)
	} else {
		payload = m.MarshalCompatiblity(s.compatibilityVersion)
	}
	err := s.conn.Write(payload, deflated)
	if err != nil {
		s.conn.Close()
	}
}

func (s *session) log() *log.Agregator {
	return log.I("no", int(s.conn.No()))
}
