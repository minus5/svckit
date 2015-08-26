package elect

import (
	"log"
	"sync"

	consulapi "github.com/hashicorp/consul/api"
)

type LeaderElection struct {
	lock           *consulapi.Lock
	cleanupChannel chan struct{}
	stopChannel    chan struct{}
	leader         bool
	sync.RWMutex
}

func New(addr, dc, key string) (*LeaderElection, error) {
	config := consulapi.DefaultConfig()
	config.Address = addr
	if dc != "" {
		config.Datacenter = dc
	}
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	lock, err := client.LockKey(key)
	if err != nil {
		return nil, err
	}

	return &LeaderElection{
		lock:           lock,
		cleanupChannel: make(chan struct{}, 1),
		stopChannel:    make(chan struct{}, 1),
	}, nil
}

func (t *LeaderElection) Start() {
	clean := false
	for !clean {
		select {
		case <-t.cleanupChannel:
			clean = true
		default:
			log.Println("Running for leader election...")
			intChan, _ := t.lock.Lock(t.stopChannel)
			if intChan != nil {
				log.Println("Now acting as leader.")
				t.Lock()
				t.leader = true
				t.Unlock()
				<-intChan
				t.Lock()
				t.leader = false
				t.Unlock()
				log.Println("Lost leadership.")
				t.lock.Unlock()
				t.lock.Destroy()
			}
		}
	}
}

func (t *LeaderElection) Stop() {
	log.Println("cleaning up")
	t.cleanupChannel <- struct{}{}
	t.stopChannel <- struct{}{}
	t.lock.Unlock()
	t.lock.Destroy()
	t.Lock()
	t.leader = false
	t.Unlock()
	log.Println("cleanup done")
}

func (t *LeaderElection) Leader() bool {
	t.RLock()
	defer t.RUnlock()
	return t.leader
}
