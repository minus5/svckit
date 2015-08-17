package elect

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	CANDIDATE_COUNT  = 5
	TEST_CONSUL_ADDR = "127.0.0.1:8500"
	TEST_CONSUL_DC   = "local"
	TEST_KEY         = "testkey"
)

var consulCmd *exec.Cmd

func startConsul() error {
	gopath := os.Getenv("GOPATH")
	gopath = strings.Split(gopath, ":")[0]
	consulDir := gopath + "/src/github.com/hashicorp/consul"
	consulConfigDir, err := os.Getwd()
	if err != nil {
		return err
	}
	log.Printf(consulConfigDir)

	if _, err := os.Stat(consulDir); os.IsNotExist(err) {
		if err := exec.Command("git", "clone", "https://github.com/hashicorp/consul.git", consulDir).Run(); err != nil {
			return fmt.Errorf("clone error: %v", err)
		}
	} else {
		if err := os.Chdir(consulDir); err != nil {
			return fmt.Errorf("chdir error: %v", err)
		}
		if err := exec.Command("git", "pull", "origin", "master").Run(); err != nil {
			return fmt.Errorf("pull error: %v", err)
		}
	}

	if err := os.Chdir(consulDir); err != nil {
		return fmt.Errorf("chdir error: %v", err)
	}

	if err := exec.Command("go", "get", "./...").Run(); err != nil {
		return fmt.Errorf("go get error: %v", err)
	}

	if err := exec.Command("make").Run(); err != nil {
		return fmt.Errorf("make error: %v", err)
	}

	consulCmd = exec.Command(gopath+"/src/github.com/hashicorp/consul/bin/consul", "agent", "-bind", "127.0.0.1", "-config-dir", consulConfigDir)
	outPipe, err := consulCmd.StdoutPipe()
	if err != nil {
		return err
	}
	var errOut error
	go func() {
		if err := consulCmd.Run(); err != nil {
			errOut = err
		}
	}()

	for errOut == nil {
		buf := make([]byte, 1000)
		_, err := outPipe.Read(buf)
		if err != nil {
			return err
		}
		log.Printf("exec out: %s", buf)
		if strings.Contains(string(buf), "Disabling EnableSingleNode") {
			return nil
		}
		time.Sleep(time.Millisecond * 100)
	}
	return errOut
}

func stopConsul() error {
	return consulCmd.Process.Signal(os.Interrupt)
}

func TestElection(t *testing.T) {
	if err := os.Setenv("GOMAXPROCS", fmt.Sprint(runtime.NumCPU())); err != nil {
		log.Println("unable to set GOMAXPROCS env var")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := startConsul()
	assert.Nil(t, err)
	if err != nil {
		log.Println(err)
		return
	}

	candidates := map[int]*LeaderElection{}
	for i := 0; i < CANDIDATE_COUNT; i++ {
		c, err := New(TEST_CONSUL_ADDR, TEST_CONSUL_DC, "", TEST_KEY)
		assert.Nil(t, err)
		assert.NotNil(t, c)
		go c.Start()
		candidates[i] = c
	}
	time.Sleep(time.Second * 1)
	leaderCount := 0
	lastLeaderId := 0

	election := func() {
		leaderCount = 0
		lastLeaderId = 0
		for ix, c := range candidates {
			if c.Leader() {
				log.Printf("Candidate %d is the LEADER", ix)
				leaderCount++
				lastLeaderId = ix
			}
		}
	}

	for i := 0; i < 20; i++ {
		election()
		if leaderCount == 0 {
			log.Println("no leaders elected! maybe Consul not ready, waiting...")
			time.Sleep(time.Second * 5)
		} else {
			log.Println("leader elected!")
			break
		}
	}

	assert.Equal(t, 1, leaderCount)

	log.Printf("Stopping candidate %d", lastLeaderId)
	candidates[lastLeaderId].Stop()
	stoppedCandidate := lastLeaderId
	time.Sleep(time.Second * 1)
	election()
	assert.Equal(t, 1, leaderCount)

	for ix, c := range candidates {
		if ix != stoppedCandidate {
			c.Stop()
		}
	}

	stopConsul()
}
