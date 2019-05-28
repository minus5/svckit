package dcy_test

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/dcy/sr"
	"github.com/minus5/svckit/log"
	"github.com/stretchr/testify/assert"
)

func TestDcy(t *testing.T) {
	cmd := exec.Command("consul", "agent", "-dev", "-datacenter=dev", "-domain=sd", "-node=node01", "-ui=true", "-bind=127.0.0.1")
	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Discard()

	err := cmd.Start()
	assert.Nil(t, err)

	tests := map[string]func(*testing.T){
		"localIntegration":   testLocalIntegration,
		"federatedIngration": testFederatedIntegration,
	}
	for name, handler := range tests {
		t.Run(name, handler)
	}

	cmd.Process.Kill()
	_ = cmd.Wait()
}

func testLocalIntegration(t *testing.T) {
	s1Port := 12345
	s2Port := 23456
	name := "test-service"
	var srvsBySubscribe dcy.Addresses

	check := func(as []string) {
		a, err := dcy.LocalServices(name)
		if len(as) == 0 {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Len(t, a, len(as))
		assert.Equal(t, as, a.String())
		assert.Equal(t, as, srvsBySubscribe.String())
	}
	wait := func() {
		time.Sleep(100 * time.Millisecond)
	}

	dcy.MustConnect()
	// subscribe to service changes
	dcy.Subscribe(name, func(srvs dcy.Addresses) {
		srvsBySubscribe = srvs
	})
	assert.Nil(t, srvsBySubscribe)

	// register first service
	s1, err := sr.New(s1Port, sr.Name(name))
	assert.Nil(t, err)
	wait()
	check([]string{"127.0.0.1:12345"})

	// register second service
	s2, err := sr.New(s2Port, sr.Name(name))
	assert.Nil(t, err)
	wait()
	check([]string{"127.0.0.1:12345", "127.0.0.1:23456"})

	// remove one
	s2.Fail()
	wait()
	check([]string{"127.0.0.1:12345"})

	// recover
	s2.Passing()
	wait()
	check([]string{"127.0.0.1:12345", "127.0.0.1:23456"})

	// deregister first
	s1.Deregister()
	wait()
	check([]string{"127.0.0.1:23456"})

	// deregister second
	s2.Deregister()
	wait()
	check([]string{})
}

// TODO: try to introduce multiple datacenters locally so this can be tested more thoroughly
func testFederatedIntegration(t *testing.T) {
	s1Port := 12345
	s2Port := 23456
	name := "test-service"
	var srvsBySubscribe dcy.Addresses

	check := func(as []string) {
		a, err := dcy.Services(name)
		if len(as) == 0 {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Len(t, a, len(as))
		assert.Equal(t, as, a.String())
		assert.Equal(t, as, srvsBySubscribe.String())
	}
	wait := func() {
		time.Sleep(100 * time.Millisecond)
	}

	dcy.MustConnect()
	// subscribe to service changes
	dcy.Subscribe(name, func(srvs dcy.Addresses) {
		srvsBySubscribe = srvs
	})
	assert.Nil(t, srvsBySubscribe)

	// register first service
	s1, err := sr.New(s1Port, sr.Name(name))
	assert.Nil(t, err)
	wait()
	// register second service
	s2, err := sr.New(s2Port, sr.Name(name))
	assert.Nil(t, err)
	wait()
	check([]string{"127.0.0.1:12345", "127.0.0.1:23456"})

	// deregister first
	s1.Deregister()
	wait()
	check([]string{"127.0.0.1:23456"})

	// deregister second
	s2.Deregister()
	wait()
	check([]string{})

}
