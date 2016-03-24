package elect

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	CandidateCount = 5
	TestConsulAddr = "127.0.0.1:8500"
	TestConsulDc   = ""
	TestKey        = "testkey"
)

func TestElection(t *testing.T) {
	t.Skip()
	candidates := map[int]*LeaderElection{}
	for i := 0; i < CandidateCount; i++ {
		c, err := New(TestConsulAddr, TestConsulDc, TestKey)
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
}
