package session

import (
	"context"
	"time"

	"github.com/mnu5/svckit/amp"
	"github.com/mnu5/svckit/log"
)

type requester interface {
	Send(amp.Subscriber, *amp.Msg)
	Unsubscribe(amp.Subscriber)
}

type broker interface {
	Subscribe(amp.Subscriber, map[string]int64)
	Unsubscribe(amp.Subscriber)
}

type connection interface {
	Read() ([]byte, error)
	Write(payload []byte, deflated bool) error
	DeflateSupported() bool
	Headers() map[string]string
	No() uint64
	Close() error
}

type session struct {
	conn           connection
	broker         broker
	requester      requester
	clientSendMsgs chan *amp.Msg
}

// serve starts new session.
// Blocks until session is finished.
func serve(ctx context.Context, conn connection, req requester, brk broker) {
	s := &session{
		conn:           conn,
		requester:      req,
		broker:         brk,
		clientSendMsgs: make(chan *amp.Msg),
	}
	// s.log().Debug("session start")
	s.loop(ctx)
}

func (s *session) log() *log.Agregator {
	return log.I("no", int(s.conn.No()))
}

func (s *session) connReadLoop() chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			buf, err := s.conn.Read()
			if err != nil {
				return
			}
			m := amp.Parse(buf)
			s.onClientMsg(m)
		}
	}()
	return done
}

func (s *session) loop(ctx context.Context) {
	go s.connWriteLoop()
	connReadDone := s.connReadLoop()
	// wait and cleanup
	select {
	case <-connReadDone: // if connection if broken
		s.unsubscribe()
	case <-ctx.Done(): // if aplication is closing
		_ = s.conn.Close()
		<-connReadDone
	}
	close(s.clientSendMsgs)
}

func (s *session) unsubscribe() {
	s.broker.Unsubscribe(s)
	s.requester.Unsubscribe(s)
}

func (s *session) onClientMsg(m *amp.Msg) {
	switch m.Type {
	case amp.Ping:
		s.clientSendMsgs <- amp.NewPong()
	case amp.Request:
		s.requester.Send(s, m)
	case amp.Subscribe:
		s.broker.Subscribe(s, m.Subscriptions)
	}
}

func (s *session) onBackendMsg(m *amp.Msg) {
	s.clientSendMsgs <- m
}

// Send implements amp.Subscriber interface
func (s *session) Send(m *amp.Msg) {
	s.onBackendMsg(m)
}

var (
	maxQueueLen   = 1024
	aliveInterval = 32 * time.Second
)

// connWriteLoop is all about not blocking clientSendMsgs chan.
// It ensures that one slow client is not blocking the rest of the app.
func (s *session) connWriteLoop() {
	queue := make([]*amp.Msg, 0)
	stopQueueing := func() {
		queue = nil
	}
	enqueue := func(m *amp.Msg) {
		if queue == nil {
			return
		}
		queue = append(queue, m)
		if len(queue) > maxQueueLen {
			s.log().I("queue", len(queue)).Debug("queue overflow")
			_ = s.conn.Close()
			stopQueueing()
		}
	}
	dequeue := func() *amp.Msg {
		m := queue[0]
		queue = queue[1:]
		return m
	}
	empty := func() bool {
		return len(queue) == 0
	}

	alive := aliveInterval
	t := time.NewTimer(alive)

	msgDone := make(chan error, 1)
	sending := false
	for {
		if !sending && !empty() {
			// start sending a message, signal on msgDone when finished
			sending = true
			go func(m *amp.Msg) {
				msgDone <- s.connWrite(m)
			}(dequeue())
		}
		select {
		case m, ok := <-s.clientSendMsgs:
			if ok {
				enqueue(m)
				t.Reset(alive)
				break
			}
			if sending {
				<-msgDone // wait for pending send
				sending = false
			}
			if empty() { // if there is nothing to send exit this loop
				return
			}
		case <-t.C:
			if !sending {
				enqueue(amp.NewAlive())
			}
			t.Reset(alive)
		case err := <-msgDone:
			sending = false
			if err != nil {
				stopQueueing()
			}
		}
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
