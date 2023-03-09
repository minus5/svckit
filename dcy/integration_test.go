package dcy_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	// For now, skipping the final check after all service instances have been deregisterd.
	// Service cannot be completely removed from cache because of the code at the end of monitor function in dcy package `if !serviceExistedOnStart && len(ses) == 0 { continue }`.
	t.Skip("service cannot be completely removed from cache")
	check([]string{})

}

// I was running this with:
//   SVCKIT_DCY_CONSUL=10.50.1.57 SVCKIT_FEDERATED_DCS="s2 zi" go test -v --run=TestInProduction
func TestInProduction(t *testing.T) {
	t.Skip()
	// tecajna, sbk_api are from current dc s2
	addr, err := dcy.Service("tecajna")
	assert.Nil(t, err)
	fmt.Printf("tecajna: %s\n", addr)

	addrs, err := dcy.Services("tecajna")
	assert.Nil(t, err)
	assert.Len(t, addrs, 1)
	fmt.Printf("tecajna services: %s\n", addrs)

	addr, err = dcy.Service("sbk_api")
	assert.Nil(t, err)
	fmt.Printf("sbk_api service: %s\n", addr)

	addrs, err = dcy.Services("sbk_api")
	assert.Nil(t, err)
	assert.Len(t, addrs, 2)
	fmt.Printf("sbk_api services: %s\n", addrs)

	// this one is from federated dc zi
	addrs, err = dcy.Services("kladomat")
	assert.Nil(t, err)
	assert.Len(t, addrs, 2)
	fmt.Printf("kladomat services: %s\n", addrs)
	fmt.Printf("kladomat URL: %s\n", dcy.URL("kladomat"))

	// this one exists in both dc-s
	addrs, err = dcy.Services("nsqlookupd-tcp")
	assert.Nil(t, err)
	fmt.Printf("nsqlookupd services: %s\n", addrs)

	// Service should always return from local dc
	for i := 0; i < 100; i++ {
		addr, err = dcy.Service("nsqlookupd-tcp")
		assert.Nil(t, err)
		assert.True(t, strings.HasPrefix(addr.String(), "10.50."))
		//		fmt.Printf("nsqlookupd: %s\n", addr)
	}
}

func TestAddressesAppend(t *testing.T) {
	as1 := dcy.Addresses{
		dcy.Address{Address: "1", Port: 1},
		dcy.Address{Address: "2", Port: 2},
	}
	as2 := dcy.Addresses{
		dcy.Address{Address: "3", Port: 3},
		dcy.Address{Address: "2", Port: 2},
	}
	assert.Len(t, as1, 2)
	assert.Len(t, as2, 2)
	as1.Append(as2)
	assert.Len(t, as1, 3)
	//fmt.Println(as1)
}
