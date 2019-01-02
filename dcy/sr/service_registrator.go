package sr

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/mnu5/svckit/dcy"
	"github.com/mnu5/svckit/env"
	"github.com/mnu5/svckit/health"
	"github.com/mnu5/svckit/log"
)

// Name sets the service name.
// Default is application name.
func Name(name string) func(*serviceRegistrator) {
	return func(s *serviceRegistrator) {
		s.name = name
	}
}

// HealthCheck sets the health check handler.
func HealthCheck(handler healthCheckHandler) func(*serviceRegistrator) {
	return func(s *serviceRegistrator) {
		s.handler = handler
	}
}

type serviceRegistrator struct {
	id        string
	name      string
	port      int
	ttl       int
	interval  int
	agent     *api.Agent
	checkId   string
	close     chan bool
	closed    chan struct{}
	setStatus chan health.Status
	handler   healthCheckHandler
}

type healthCheckHandler func() (health.Status, []byte)

// New ...
func New(port int, opts ...func(*serviceRegistrator)) (*serviceRegistrator, error) {
	s := &serviceRegistrator{
		name:      env.AppName(),
		port:      port,
		ttl:       10,
		interval:  9,
		close:     make(chan bool),
		closed:    make(chan struct{}),
		setStatus: make(chan health.Status),
	}
	// apply options
	for _, opt := range opts {
		opt(s)
	}
	// ids
	s.id = fmt.Sprintf("%s:%d", s.name, s.port)
	s.checkId = fmt.Sprintf("%s_ttl_check", s.id)
	// register
	if err := s.register(); err != nil {
		return nil, err
	}
	go s.loop()
	return s, nil
}

// Passing sets status to passing.
func (s *serviceRegistrator) Passing() {
	s.setStatus <- health.Passing
}

// Warn sets status to warn.
func (s *serviceRegistrator) Warn() {
	s.setStatus <- health.Warn
}

// Fail sets status to fail.
func (s *serviceRegistrator) Fail() {
	s.setStatus <- health.Fail
}

// Deregister service in consul.
func (s *serviceRegistrator) Deregister() {
	s.close <- true
	<-s.closed
}

// Close alias for Deregister.
func (s *serviceRegistrator) Close() {
	s.Deregister()
}

// Stop sending ttl to consul without deregister.
func (s *serviceRegistrator) Stop() {
	s.close <- false
	<-s.closed
}

func (s *serviceRegistrator) loop() {
	status := health.Passing
	var note []byte

	readAndUpdateStatus := func() {
		if s.handler != nil {
			status, note = s.handler()
		}
		s.updateStatus(status, note)
	}

	readAndUpdateStatus()
	for {
		select {
		case <-time.After(time.Duration(s.interval) * time.Second):
			readAndUpdateStatus()
		case newStatus := <-s.setStatus:
			if status != newStatus {
				status = newStatus
				s.updateStatus(status, note)
			}
		case dereg := <-s.close:
			if dereg {
				s.deregister()
			}
			close(s.closed)
			return
		}
	}
}

func (s *serviceRegistrator) deregister() {
	_ = s.agent.ServiceDeregister(s.id)
}

func (s *serviceRegistrator) register() error {
	s.agent = dcy.Agent()

	service := &api.AgentServiceRegistration{
		ID:   s.id,
		Name: s.name,
		Port: s.port,
	}
	check := &api.AgentCheckRegistration{
		ID:        s.checkId,
		Name:      fmt.Sprintf("Service '%s' ttl check", s.name),
		Notes:     "",
		ServiceID: service.ID,
		AgentServiceCheck: api.AgentServiceCheck{
			TTL: fmt.Sprintf("%ds", s.ttl),
			//Status: "passing",
		},
	}

	if err := s.agent.ServiceRegister(service); err != nil {
		return err
	}
	if err := s.agent.CheckRegister(check); err != nil {
		return err
	}
	return nil
}

func (s *serviceRegistrator) updateStatus(status health.Status, note []byte) {
	fn := s.agent.FailTTL
	switch status {
	case health.Passing:
		fn = s.agent.PassTTL
	case health.Warn:
		fn = s.agent.WarnTTL
	}
	err := fn(s.checkId, string(note))
	if err != nil {
		log.Error(err)
	}
}
