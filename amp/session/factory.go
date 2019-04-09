package session

import (
	"context"
	"sync"
	"time"

	"github.com/minus5/svckit/amp"
)

type requester interface {
	Send(amp.Subscriber, *amp.Msg) // send request
	Unsubscribe(amp.Subscriber)    // stop waiting for responses
	Wait()                         // wait for clean exit
}

type broker interface {
	Subscribe(amp.Subscriber, map[string]int64) // subscribe to the topics
	Unsubscribe(amp.Subscriber)                 // unsubscribe from all topics
	Wait()                                      // wait for clean exit
}

type connection interface {
	Read() ([]byte, error)                     // get client message
	Write(payload []byte, deflated bool) error // send message to the client
	DeflateSupported() bool                    // does websocket connection support per message deflate
	Headers() map[string]string                // http headers we got on connection open
	No() uint64                                // connection identifier (for grouping logs)
	Close() error                              // close connection
}

// Sessions is a session factory
type Sessions struct {
	broker    broker
	requester requester
	cancelSig context.Context
	closed    chan struct{}
	wg        sync.WaitGroup
}

// Factory creates new seessions factory.
func Factory(ctx context.Context, broker broker, requester requester) *Sessions {
	cancelSig, cancelSessions := context.WithCancel(context.Background())
	s := &Sessions{
		broker:    broker,
		requester: requester,
		cancelSig: cancelSig,
		closed:    make(chan struct{}),
	}

	go s.waitDone(ctx, cancelSessions)
	return s
}

// Serve creates new session for connection.
func (s *Sessions) Serve(conn connection) {
	s.wg.Add(1)
	go func() {
		serve(s.cancelSig, conn, s.requester, s.broker)
		s.wg.Done()
	}()
}

func (s *Sessions) waitDone(ctx context.Context, cancelSessions func()) {
	<-ctx.Done()       // wait for application interupt signal
	s.requester.Wait() // wait for clean exit of requester
	s.broker.Wait()    //   and broker
	cancelSessions()   // request cancel of all session
	s.wg.Wait()        // wait for all sessions to exit
	close(s.closed)    // signal that I'am closed
}

// Wait blocks until all sessions are closed.
func (s *Sessions) Wait() {
	<-s.closed
}

var (
	poolInterval     = 32 * time.Second
	waitManyInterval = 2 * time.Millisecond
)

// Pool gets response messages for long pooling interface
func (s *Sessions) Pool(m *amp.Msg) []*amp.Msg {
	s.wg.Add(1)
	defer s.wg.Done()

	switch m.Type {
	case amp.Ping:
		return []*amp.Msg{m.Pong()}
	case amp.Request:
		p := newPooler()
		s.requester.Send(p, m)
		p.waitOne(s.cancelSig, poolInterval)
		s.requester.Unsubscribe(p)
		return p.msgs
	case amp.Subscribe:
		p := newPooler()
		s.broker.Subscribe(p, m.Subscriptions)
		p.wait(s.cancelSig, poolInterval)
		s.broker.Unsubscribe(p)
		return p.msgs
	}
	return nil
}

func newPooler() *pooler {
	ctx, cancel := context.WithCancel(context.Background())
	return &pooler{
		msgWait: ctx,
		onMsg:   cancel,
	}
}

type pooler struct {
	msgs    []*amp.Msg
	onMsg   func()
	msgWait context.Context
	sync.Mutex
}

func (p *pooler) Send(m *amp.Msg) {
	p.Lock()
	p.msgs = append(p.msgs, m)
	p.Unlock()
	p.onMsg()
}

func (p *pooler) waitOne(app context.Context, interval time.Duration) {
	select {
	case <-app.Done():
	case <-time.After(interval):
	case <-p.msgWait.Done():
	}
}

func (p *pooler) wait(app context.Context, interval time.Duration) {
	select {
	case <-app.Done():
	case <-time.After(interval):
	case <-p.msgWait.Done():
		select {
		case <-app.Done():
		case <-time.After(waitManyInterval):
		}
	}
}
