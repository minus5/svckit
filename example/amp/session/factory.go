package session

import (
	"context"
	"sync"
)

// Sessions is a session factory
type Sessions struct {
	b        broker
	r        requester
	ctx      context.Context
	ctxClose context.CancelFunc
	wg       sync.WaitGroup
}

// Factory creates new seessions factory.
func Factory(b broker, r requester) *Sessions {
	ctx, ctxClose := context.WithCancel(context.Background())
	return &Sessions{
		b:        b,
		r:        r,
		ctx:      ctx,
		ctxClose: ctxClose,
	}
}

// Serve creates new session for connection.
func (s *Sessions) Serve(c connection) {
	s.wg.Add(1)
	go func() {
		serve(s.ctx, c, s.r, s.b)
		s.wg.Done()
	}()
}

// Close starts closing factory.
func (s *Sessions) Close() {
	s.ctxClose()
}

// Wait blocks until all sessions are closed.
func (s *Sessions) Wait() {
	s.wg.Wait()
}
