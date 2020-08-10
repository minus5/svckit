package session

import (
	"context"
	"fmt"
	"time"

	"github.com/minus5/svckit/amp"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
)

var (
	maxWriteQueueDepth = 256              // max number of messages in outQueue
	aliveInterval      = 32 * time.Second // interval for sending alive message
)

type session struct {
	conn        connection      // client websocket connection
	broker      broker          // broker for subscribe on published messages
	requester   requester       // requester for request / response messages
	outMessages chan []*amp.Msg // output messages queue
	stats       struct {        // sessions stats counters
		start         time.Time
		outMessages   int
		inMessages    int
		aliveMessages int
		maxQueueLen   int
	}
	compatibilityVersion uint8
	overflow             chan struct{}
	overflowRead         chan struct{}
}

// serve starts new session
// Blocks until session is finished.
func serve(cancelSig context.Context, conn connection, req requester, brk broker,
	compatibilityVersion uint8) {
	overflow := make(chan struct{}, 1)
	s := &session{
		conn:                 conn,
		requester:            req,
		broker:               brk,
		outMessages:          make(chan []*amp.Msg, 256),
		compatibilityVersion: compatibilityVersion,
		overflow:             overflow,
		overflowRead:         overflow, // read once and set to nil
	}
	s.stats.start = time.Now()
	s.loop(cancelSig)
}

func (s *session) loop(cancelSig context.Context) {
	inMessages := s.readLoop()  // messages from the client
	exitSig := cancelSig.Done() // aplication exit signal

	// timer for alive messages
	alive := time.NewTimer(aliveInterval)
	// if there is no other messages send alive
	sendAlive := func() {
		s.Send(amp.NewAlive())
		s.stats.aliveMessages++
	}

	defer s.logStats()

	for {
		select {
		case <-alive.C:
			sendAlive()
		case msgs := <-s.outMessages:
			for _, msg := range msgs {
				s.connWrite(msg)
			}
			s.stats.outMessages += len(msgs)
			alive.Reset(aliveInterval)
		case msg, ok := <-inMessages:
			if !ok {
				s.unsubscribe()
				return
			}
			s.receive(msg)
			s.stats.inMessages++
		case <-exitSig:
			s.connClose()
			exitSig = nil // fire once
		case <-s.overflowRead:
			s.connClose()
			s.overflowRead = nil // fire once
		}
	}
}

func (s *session) logStats() {
	duration := int(time.Now().Sub(s.stats.start) / time.Millisecond)
	s.log().I("inMessages", s.stats.inMessages).
		I("outMessages", s.stats.outMessages).
		I("aliveMessages", s.stats.aliveMessages).
		I("durationMs", duration).
		Debug("stats")
	metric.Time("duration", duration)
	metric.Time("inMessages", s.stats.inMessages)
	metric.Time("outMessages", s.stats.outMessages)
	metric.Time("aliveMessages", s.stats.aliveMessages)
	metric.Time("maxQueueLen", s.stats.maxQueueLen)
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
	s.SendMsgs([]*amp.Msg{m})
}

func (s *session) SendMsgs(msgs []*amp.Msg) {
	select {
	case s.outMessages <- msgs:
	default:
		select {
		case s.overflow <- struct{}{}:
		default:
		}
	}
}

// should be called during s.Lock
func (s *session) logOutQueueOverflow() {
	s.log().
		S("start", fmt.Sprintf("%v", s.stats.start)).
		I("inMessages", s.stats.inMessages).
		I("outMessages", s.stats.outMessages).
		I("aliveMessages", s.stats.aliveMessages).
		I("durationMs", int(time.Now().Sub(s.stats.start)/time.Millisecond)).
		Info("out queue overflow")
}

func (s *session) connWrite(m *amp.Msg) {
	var payload []byte
	deflated := false
	if s.conn.DeflateSupported() {
		payload, deflated = m.MarshalDeflateCompatiblity(s.compatibilityVersion)
	} else {
		payload = m.MarshalCompatiblity(s.compatibilityVersion)
	}
	if payload == nil {
		return
	}
	err := s.conn.Write(payload, deflated)
	if err != nil {
		s.connClose()
	}
}

func (s *session) log() *log.Agregator {
	return log.I("no", int(s.conn.No()))
}

func (s *session) connClose() {
	metric.Timing("connClose", func() {
		s.conn.Close()
	})
}

func (s *session) Meta() map[string]string {
	return s.conn.Meta()
}
