package sr

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/minus5/svckit/dcy"
	"github.com/stretchr/testify/assert"
)

func TestServiceRegistrator(t *testing.T) {
	//t.Skip("test depends on running consul instance")
	dcy.MustConnect()
	name := "service_registrator_test"
	port := 12345

	//ovo napravi registrira service napravi check i postavi stanje na passing
	sr, err := New(port, Name("service_registrator_test"))
	assert.Nil(t, err)
	svc, checks := consulService(t, name)
	assert.NotNil(t, svc)
	assert.Equal(t, port, svc.Port)
	assert.Equal(t, 1, len(checks))
	assert.Equal(t, "passing", checks[0].Status)
	//time.Sleep(10 * time.Second)

	sr.SetStatus(Warn)
	svc, checks = consulService(t, name)
	assert.Equal(t, "warning", checks[0].Status)
	//time.Sleep(10 * time.Second)

	sr.SetStatus(Fail)
	svc, checks = consulService(t, name)
	assert.Equal(t, "critical", checks[0].Status)
	//time.Sleep(10 * time.Second)

	sr.SetStatus(Passing)
	svc, checks = consulService(t, name)
	assert.Equal(t, "passing", checks[0].Status)
	//time.Sleep(10 * time.Second)

	sr.Deregister()
	svc, checks = consulService(t, name)
	assert.Nil(t, svc)
}

//consulService nadji u consulu servis imena name, vrati za njega podatke i podatke za sve njegove check-ove
func consulService(t *testing.T, name string) (*api.AgentService, []*api.AgentCheck) {
	svcs, err := dcy.Agent().Services()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range svcs {
		if s.Service == name {
			checks, err := dcy.Agent().Checks()
			if err != nil {
				t.Fatal(err)
			}
			sa := []*api.AgentCheck{}
			for _, c := range checks {
				if c.ServiceID == s.ID {
					sa = append(sa, c)
				}
			}
			return s, sa
		}
	}
	return nil, nil
}
