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
	Created(sender amp.Sender)
	Subscribe(amp.Sender, map[string]int64) // subscribe to the topics
	Unsubscribe(amp.Sender)                 // unsubscribe from all topics
	Wait()                                  // wait for clean exit
}

type connection interface {
	Read() ([]byte, error)                     // get client message
	Write(payload []byte, deflated bool) error // send message to the client
	DeflateSupported() bool                    // does websocket connection support per message deflate
	Headers() map[string]string                // http headers we got on connection open
	No() uint64                                // connection identifier (for grouping logs)
	Close() error                              // close connection
	Meta() map[string]string                   // session metadata, set by the client
	SetMeta(map[string]string)                 // set session metadata
	GetRemoteIp() string
	GetCookie() string
}

type counter struct {
	value int
	max   int
	sync.Mutex
}

func (c *counter) Up() {
	c.Lock()
	defer c.Unlock()
	c.value++
	if c.value > c.max {
		c.max = c.value
	}
}

func (c *counter) Down() {
	c.Lock()
	defer c.Unlock()
	c.value--
}

func (c *counter) Count() int {
	c.Lock()
	defer c.Unlock()
	v := c.value
	if c.max > c.value {
		v = c.max
	}
	c.max = 0
	return v
}

// Sessions is a session factory
type Sessions struct {
	broker             broker
	requester          requester
	cancelSig          context.Context
	closed             chan struct{}
	wg                 sync.WaitGroup
	wsConnections      counter
	poolingConnections counter
	// topicWhitelist is a list of topics that clients can send requests to.
	// Empty value means block all.
	topicWhitelist []string
}

// Factory creates new Sessions factory.
func Factory(ctx context.Context, broker broker, requester requester, topicWhitelist []string) *Sessions {
	cancelSig, cancelSessions := context.WithCancel(context.Background())
	s := &Sessions{
		broker:         broker,
		requester:      requester,
		cancelSig:      cancelSig,
		closed:         make(chan struct{}),
		topicWhitelist: topicWhitelist,
	}

	go s.waitDone(ctx, cancelSessions)
	return s
}

// Serve creates new session for connection.
// Blocks until connection is closed
func (s *Sessions) Serve(conn connection) {
	s.wg.Add(1)
	s.wsConnections.Up()
	serve(s.cancelSig, conn, s.requester, s.broker, s.topicWhitelist, amp.CompatibilityVersionDefault)
	s.wg.Done()
	s.wsConnections.Down()
}

// Serve creates new session for connection.
// Blocks until connection is closed
func (s *Sessions) ServeV1(conn connection) {
	s.wg.Add(1)
	s.wsConnections.Up()
	serve(s.cancelSig, conn, s.requester, s.broker, s.topicWhitelist, amp.CompatibilityVersion1)
	s.wg.Done()
	s.wsConnections.Down()
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
	s.poolingConnections.Up()
	defer s.wg.Done()
	defer s.poolingConnections.Down()

	switch m.Type {
	case amp.Ping:
		return []*amp.Msg{m.Pong()}
	case amp.Request:
		p := newPooler(m.Meta)
		s.requester.Send(p, m)
		p.waitOne(s.cancelSig, poolInterval)
		s.requester.Unsubscribe(p)
		return p.msgs
	case amp.Subscribe:
		p := newPooler(m.Meta)
		s.broker.Subscribe(p, m.Subscriptions)
		p.wait(s.cancelSig, poolInterval)
		s.broker.Unsubscribe(p)
		return p.msgs
	}
	return nil
}

func (s *Sessions) ConnectionsCount() (int, int) {
	return s.wsConnections.Count(), s.poolingConnections.Count()
}

func newPooler(meta map[string]string) *pooler {
	ctx, cancel := context.WithCancel(context.Background())
	return &pooler{
		msgWait: ctx,
		onMsg:   cancel,
	}
}

type pooler struct {
	meta    map[string]string
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

func (p *pooler) SendMsgs(m []*amp.Msg) {
	p.Lock()
	p.msgs = append(p.msgs, m...)
	p.Unlock()
	p.onMsg()
}

func (p *pooler) Meta() map[string]string {
	return p.meta
}

func (p *pooler) Headers() map[string]string {
	return nil
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
