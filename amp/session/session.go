package session

import (
	"context"
	"time"

	"github.com/mnu5/svckit/amp"
	"github.com/mnu5/svckit/log"
)

var (
	maxWriteQueueDepth = 128              // max number of messages in client write queue
	aliveInterval      = 32 * time.Second // interval for sending alive messages to the client
)

type session struct {
	conn        connection    // client websocket connection
	broker      broker        // broker for subscribe on published messages
	requester   requester     // requester for request / response messages
	outMessages chan *amp.Msg // messages to send to the client
}

// serve starts new session
// Blocks until session is finished.
func serve(cancelSig context.Context, conn connection, req requester, brk broker) {
	s := &session{
		conn:        conn,
		requester:   req,
		broker:      brk,
		outMessages: make(chan *amp.Msg),
	}
	s.loop(cancelSig)
}

func (s *session) loop(cancelSig context.Context) {
	go s.writeLoop()
	readDone := s.readLoop()
	// wait and cleanup
	select {
	case <-readDone: // if connection if broken
		s.unsubscribe()
	case <-cancelSig.Done(): // if aplication is closing
		_ = s.conn.Close()
		<-readDone
	}
	close(s.outMessages)
}

func (s *session) unsubscribe() {
	s.broker.Unsubscribe(s)
	s.requester.Unsubscribe(s)
}

func (s *session) readLoop() chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			buf, err := s.conn.Read()
			if err != nil {
				return
			}
			if m := amp.Parse(buf); m != nil {
				s.receive(m)
			}
		}
	}()
	return done
}

// receive gets client messages
func (s *session) receive(m *amp.Msg) {
	switch m.Type {
	case amp.Ping:
		s.outMessages <- m.Pong()
	case amp.Request:
		// TODO what URI-a are ok, make filter
		s.requester.Send(s, m)
	case amp.Subscribe:
		s.broker.Subscribe(s, m.Subscriptions)
	}
}

// Send message to the clinet
// Implements amp.Subscriber interface.
func (s *session) Send(m *amp.Msg) {
	s.outMessages <- m
}

// writeLoop is all about not blocking outMessages chan.
// It ensures that one slow client is not blocking the rest of the app.
func (s *session) writeLoop() {
	queue := make([]*amp.Msg, 0)   // we are queuing messges in slice so chan is not blocked
	out := make(chan *amp.Msg)     // out is processin messages from the queue
	outDone := make(chan struct{}) // signal that out loop is done

	go func() { // out loop
		defer close(outDone)
		for m := range out {
			if err := s.connWrite(m); err != nil {
				return
			}
		}
	}()

	// timer for alive messages, if there is no other messages we send alive
	alive := aliveInterval
	t := time.NewTimer(alive)

	stopQueueing := func() {
		queue = nil
	}
	enqueue := func(m *amp.Msg) {
		if queue == nil {
			return
		}
		queue = append(queue, m)
		if len(queue) > maxWriteQueueDepth {
			s.log().I("queue", len(queue)).Debug("queue overflow")
			_ = s.conn.Close()
			stopQueueing()
		}
	}

	// main loop
	for {
		if len(queue) > 0 { // if there is messages in the queue
			select {
			case out <- queue[0]: // try to send to out
				queue = queue[1:]
			case m, ok := <-s.outMessages: // or receive new
				if !ok {
					return
				}
				enqueue(m)
			case <-outDone:
				stopQueueing()
			}
		} else { // if the queue is empty
			select {
			case m, ok := <-s.outMessages: // receive new
				if !ok {
					return
				}
				enqueue(m)
			case <-t.C: // or react to timer
				enqueue(amp.NewAlive())
			}
		}
		t.Reset(alive)
	}
}

func (s *session) connWrite(m *amp.Msg) error {
	var payload []byte
	deflated := false
	if s.conn.DeflateSupported() {
		payload, deflated = m.MarshalDeflate()
	} else {
		payload = m.Marshal()
	}
	return s.conn.Write(payload, deflated)
}

func (s *session) log() *log.Agregator {
	return log.I("no", int(s.conn.No()))
}
