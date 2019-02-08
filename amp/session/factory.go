package session

import (
	"context"
	"sync"

	"github.com/mnu5/svckit/amp"
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

type callbacks interface {
}

// Sessions is a session factory
type Sessions struct {
	broker    broker
	requester requester
	cancelSig context.Context
	cancelFn  context.CancelFunc
	wg        sync.WaitGroup
}

// Factory creates new seessions factory.
func Factory(broker broker, requester requester) *Sessions {
	cancelSig, cancelFn := context.WithCancel(context.Background())
	return &Sessions{
		broker:    broker,
		requester: requester,
		cancelSig: cancelSig,
		cancelFn:  cancelFn,
	}
}

// Serve creates new session for connection.
func (s *Sessions) Serve(conn connection) {
	s.wg.Add(1)
	go func() {
		serve(s.cancelSig, conn, s.requester, s.broker)
		s.wg.Done()
	}()
}

// Wait blocks until all sessions are closed.
func (s *Sessions) Wait() {
	s.requester.Wait()
	s.broker.Wait()
	s.cancelFn()
	s.wg.Wait()
}
