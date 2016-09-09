package leader

import (
	"fmt"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"
	"sync"

	"github.com/hashicorp/consul/api"
)

func logger(o *options) *log.Agregator {
	return log.S("lib", "svckit/leader").S("keyPrefix", o.keyPrefix)
}

type options struct {
	keyPrefix string
	wg        *sync.WaitGroup
}

// KeyPrefix postavlja prefix za consul leadership key.
func KeyPrefix(k string) func(*options) {
	return func(o *options) {
		o.keyPrefix = k
	}
}

// WaitGroup postaviti za clean exit.
// Ako je postavljena unutar New ce se pozvati wg.Add(1) pa na kraju wg.Done().
// Izvana pozovemo wg.Wait() pa smo sigurni da je leader napravio clean exit (pustio key u consulu).
// Zgodno kada imamo vise leadera u aplikaciji.
func WaitGroup(wg *sync.WaitGroup) func(*options) {
	return func(o *options) {
		o.wg = wg
	}
}

//New - ako je ovaj proces postao leader zove worker-a.
//Parametar workera je kanal koji ce biti closan kad izgubi leadership.
//Postoje tri razloga zasto bi se leadership izgubio:
// * consul signalizira da vise nije leader
// * aplikacija dobije USR1 signal
// * aplikacija dobije interupt TERM ili INT signal
// U prva dva slucaja ponovo pokusavamo dobiti leadership i nastaviti raditi.
// U trecem izlazimo.
// USR1 je signal aplikaciji da pusti leadership.
func New(worker func(<-chan struct{}), opts ...func(*options)) error {
	o := &options{
		keyPrefix: env.AppName(),
		wg:        nil,
	}
	// apply options
	for _, fn := range opts {
		fn(o)
	}
	// consul leadership key
	key := fmt.Sprintf("%s-leadership_key", o.keyPrefix)

	// kontrolni kanali
	stopAcquiringLeadership := make(chan struct{})
	interupt := make(chan struct{})

	// cekam na vanjski interput
	go func() {
		signal.WaitForInterupt()
		logger(o).Debug("interupt received")
		close(stopAcquiringLeadership)
		close(interupt)
	}()

	if o.wg != nil {
		o.wg.Add(1)
	}
	defer func() {
		if o.wg != nil {
			o.wg.Done()
		}
	}()

	for {
		select {
		case <-interupt:
			logger(o).Debug("exit")
			return nil
		default:
			//		again:
			logger(o).Debug("acquiring leadership...")
			var leader *api.Lock
			var leaderShipLost <-chan struct{}

			connect := func() error {
				var err error
				leader, err = dcy.LockKey(key)
				if err != nil {
					return err
				}
				leaderShipLost, err = leader.Lock(stopAcquiringLeadership)
				if err != nil {
					logger(o).Error(err)
					return err
				}
				return nil
			}
			if err := signal.WithExponentialBackoff(connect); err != nil {
				logger(o).Error(err)
				return err
			}

			// ako sam postao leader
			if leaderShipLost != nil {
				logger(o).Debug("leadership acquired")
				// cekam na signale
				go func() {
					usr1 := signal.Usr1()
					for {
						select {
						case <-usr1:
							logger(o).Debug("usr1 signal received")
							leader.Unlock()
							return
						case <-interupt:
							leader.Unlock()
							return
						case <-leaderShipLost:
							return
						}
					}
				}()
				// zovem workera
				worker(leaderShipLost)
				leader.Unlock() //za slucaj da je work zavrsio sam od sebe
			}
		}
	}
}
