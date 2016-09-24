package dcy

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/minus5/svckit/env"
)

const (
	Passing = iota
	Warn
	Fail
)

type ServiceRegistrator struct {
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

func Register(port int) (*ServiceRegistrator, error) {
	return NewServiceRegistrator(env.AppName(), port, nil)
}

//NewServiceRegistrator prijavi servis tog imena i na tom portu lokalnom consulu
//Registrator ce svakih interval sekundi pozvati handler i status koji on vrati poslati consulu.
func NewServiceRegistrator(name string, port int, handler HealthCheckHandler) (*ServiceRegistrator, error) {
	id := fmt.Sprintf("%s:%d", name, port)
	s := &ServiceRegistrator{
		id:        id,
		name:      name,
		port:      port,
		ttl:       10,
		interval:  9,
		close:     make(chan bool),
		closed:    make(chan struct{}),
		setStatus: make(chan int),
		checkId:   fmt.Sprintf("%s_ttl_check", id),
		handler:   handler,
	}
	if err := s.register(); err != nil {
		return nil, err
	}
	go s.loop()
	return s, nil
}

// SetStatus postavi novi status u consulu, moguci statusi su const-ovi ServiceStatus*
func (s *ServiceRegistrator) SetStatus(status int) {
	if status < Passing || status > Fail {
		status = Fail
	}
	s.setStatus <- status
}

func (s *ServiceRegistrator) Passing() {
	s.SetStatus(Passing)
}

func (s *ServiceRegistrator) Warn() {
	s.SetStatus(Warn)
}

func (s *ServiceRegistrator) Fail() {
	s.SetStatus(Fail)
}

// Deregister zaustavi registrator, odjavi servis
func (s *ServiceRegistrator) Deregister() {
	s.close <- true
	<-s.closed
}

// Stop zaustavi registrator
func (s *ServiceRegistrator) Stop() {
	s.close <- false
}

func (s *ServiceRegistrator) loop() {
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

func (s *ServiceRegistrator) deregister() {
	s.agent.ServiceDeregister(s.id)
}

func (s *ServiceRegistrator) register() error {
	s.agent = consul.Agent()

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

func (s *ServiceRegistrator) updateStatus(status int, note string) {
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
