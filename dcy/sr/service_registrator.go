package sr

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
)

const (
	Passing = iota
	Warn
	Fail
)

// Name sets the service name.
// Default is application name.
func Name(name string) func(*serviceRegistrator) {
	return func(s *serviceRegistrator) {
		s.name = name
	}
}

// HealthCheck sets the health check handler.
func HealthCheck(handler HealthCheckHandler) func(*serviceRegistrator) {
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
	setStatus chan int
	handler   HealthCheckHandler
}

type HealthCheckHandler func() (int, string)

// Register ...
func New(port int, opts ...func(*serviceRegistrator)) (*serviceRegistrator, error) {
	s := &serviceRegistrator{
		name:      env.AppName(),
		port:      port,
		ttl:       10,
		interval:  9,
		close:     make(chan bool),
		closed:    make(chan struct{}),
		setStatus: make(chan int),
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

// SetStatus postavi novi status u consulu, moguci statusi su const-ovi ServiceStatus*
func (s *serviceRegistrator) SetStatus(status int) {
	if status < Passing || status > Fail {
		status = Fail
	}
	s.setStatus <- status
}

func (s *serviceRegistrator) Passing() {
	s.SetStatus(Passing)
}

func (s *serviceRegistrator) Warn() {
	s.SetStatus(Warn)
}

func (s *serviceRegistrator) Fail() {
	s.SetStatus(Fail)
}

// Deregister zaustavi registrator, odjavi servis
func (s *serviceRegistrator) Deregister() {
	s.close <- true
	<-s.closed
}

// Stop zaustavi registrator
func (s *serviceRegistrator) Stop() {
	s.close <- false
}

func (s *serviceRegistrator) loop() {
	status := Passing
	note := ""

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

func (s *serviceRegistrator) updateStatus(status int, note string) {
	fn := s.agent.FailTTL
	switch status {
	case Passing:
		fn = s.agent.PassTTL
	case Warn:
		fn = s.agent.WarnTTL
	}
	err := fn(s.checkId, note)
	if err != nil {
		log.Printf("error: %s", err)
	}
}
