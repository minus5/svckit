package dcy_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/dcy/sr"
	"github.com/stretchr/testify/assert"
)

func startConsul(t *testing.T) *exec.Cmd {
	cmd := exec.Command("consul", "agent", "-server=true", "-bootstrap=true", "-bind=127.0.0.1", "-dc=dev", "-data-dir=./tmp/consul", "-domain=sd", "-node=node01", "-ui=true")
	randomBytes := &bytes.Buffer{}
	cmd.Stdout = randomBytes
	err := cmd.Start()
	assert.Nil(t, err)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	return cmd
}

func stopConsul(t *testing.T, cmd *exec.Cmd) {
	// stop consul
	err := cmd.Process.Signal(os.Kill)
	assert.Nil(t, err)
	cmd.Wait()
	//fmt.Printf("consul output:\n%s\n", string(randomBytes.Bytes()))
}

// This test requires working consul service
// Could be started somethin like this
//  consul agent -server=true -bootstrap=true -bind=127.0.0.1 -dc=dev -data-dir=./tmp/consul -domain=sd -node=node01 -ui=true
// Or using startConsul and stopConsul functions.
// I prefer first way because starting needs few seconds.
func TestIntegration(t *testing.T) {
	t.Skip("requires consul service")
	//cmd := startConsul(t)

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

	//stopConsul(t, cmd)
}
