package main

import (
	"github.com/minus5/svckit/example/nsq_rr/api"
	"github.com/minus5/svckit/example/nsq_rr/api/nsq"
	"github.com/minus5/svckit/signal"
)

var topic = "nsq_rr.req"

func main() {
	svc := NewService()
	srv := nsq.NewServer(svc)

	signal.WaitForInterupt()
	srv.Close()
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Add(x, y int) (int, error) {
	if x+y > 256 {
		return 0, api.ErrOverflow
	}
	return x + y, nil
}
