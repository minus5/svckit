package statsd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
	"github.com/minus5/svckit/signal"

	api "github.com/alexcesaro/statsd"
)

const (
	StatsdServiceName = "statsd" //Default name in service discovery
)

type client interface {
	Count(stat string, value interface{})
	Gauge(stat string, value interface{})
	Timing(stat string, value interface{})
	Clone(opts ...api.Option) *api.Client
}

// Statsd metric driver.
// Implements metirc.Metric interface.
// Client is underlaying library.
type Statsd struct {
	prefix     string
	mainClient client
	client     client
	mapLock    sync.Mutex
	prefixes   map[string]*Statsd
}

// Same as Dial but raises Fatal on error.
func MustDial(opts ...string) {
	r := 0
again:
	if err := Dial(opts...); err != nil {
		if r > 10 {
			log.Fatal(err)
		}
		r++
		time.Sleep(time.Second)
		goto again
	}
}

// Connect to statsd server and set it as metric driver.
//   opts[0] - prefix
//   opts[1] - statsd address
// Examples:
//   Dial()
//   Dial("my_app")
//   Dial("my_appp", "127.0.0.1:8125")
func Dial(opts ...string) error {
	prefix := fmt.Sprintf("%s.%s", env.AppName(), env.InstanceId())
	addr := ""
	if len(opts) > 0 {
		prefix = opts[0]
	}
	if len(opts) > 1 {
		addr = opts[1]
	} else if os.Getenv("STATSD_LOGGER_ADDRESS") != "" {
		addr = os.Getenv("STATSD_LOGGER_ADDRESS")
	} else {
		// get statsd address from service discovery
		var a dcy.Address
		err := signal.WithExponentialBackoff(func() error {
			var err error
			a, err = dcy.Service(StatsdServiceName)
			return err
		})
		if err != nil {
			return err
		}
		addr = a.String()
	}
	mainClient, err := api.New(api.Address(addr))
	withPrefix := mainClient.Clone(api.Prefix(prefix))
	if err != nil {
		return err
	}
	//set statsd as metric dirver
	metric.Set(&Statsd{prefix: prefix, mainClient: mainClient, client: withPrefix, prefixes: make(map[string]*Statsd)})
	logger().S("addr", addr).S("prefix", prefix).Info("connected")
	return nil
}

// Increments counter name for sum(values).
// If called witohout values will increment for 1.
func (i *Statsd) Counter(name string, values ...int) {
	go func() {
		value := 1
		if len(values) > 0 {
			value = 0
			for _, v := range values {
				value += v
			}
		}
		i.client.Count(name, value)
	}()
}

// Submits/Updates a statsd gauge type.
func (i *Statsd) Gauge(name string, value int) {
	go i.client.Gauge(name, value)
}

// Measures execution time for f and submits it as statsd timing type.
func (i *Statsd) Timing(name string, f func()) {
	stopwatch := metric.NewStopwatch()
	f()
	duration := stopwatch.GetNs()
	go i.client.Timing(name, duration)
}

// Submits a statsd timing type.
func (i *Statsd) Time(name string, duration int) {
	go i.client.Timing(name, duration)
}

// Returns the clone of the original metric, but with a different prefix
func (i *Statsd) WithPrefix(prefix string) metric.Metric {
	i.mapLock.Lock()
	defer i.mapLock.Unlock()
	s, ok := i.prefixes[prefix]
	if ok && s != nil {
		return s
	}
	i.prefixes[prefix] = &Statsd{prefix: prefix, mainClient: i.client, client: i.mainClient.Clone(api.Prefix(prefix)), prefixes: make(map[string]*Statsd)}
	return i.prefixes[prefix]
}

// Returns the clone of the original metric, but with the
// suffix appended to the end of the original prefix
func (i *Statsd) AppendSuffix(suffix string) metric.Metric {
	return i.WithPrefix(fmt.Sprintf("%s.%s", i.prefix, suffix))
}

func logger() *log.Agregator {
	return log.S("lib", "svckit.metric")
}
