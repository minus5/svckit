package statsd

import (
	"fmt"
	"strings"
	"sync"
	"time"

	golog "log"

	api "github.com/minus5/go-statsd"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
)

const (
	// StatsdServiceName default name in service discovery
	StatsdServiceName = "statsd"

	// DefaultBufPoolCapacity use ringbuffer of 100 buffers/packets 1000 * (1432 + 1024)=2456000 bytes = 2.342mb
	DefaultBufPoolCapacity = 1000

	// DefaultSendQueueCapacity max queue for sending
	DefaultSendQueueCapacity = 900

	// DefaultSendLoopCount max loops for sending data
	DefaultSendLoopCount = 1
)

type client interface {
	Incr(stat string, count int64, tags ...api.Tag)
	Decr(stat string, count int64, tags ...api.Tag)
	Timing(stat string, delta int64, tags ...api.Tag)
	PrecisionTiming(stat string, delta time.Duration, tags ...api.Tag)
	Gauge(stat string, value int64, tags ...api.Tag)
	GaugeDelta(stat string, value int64, tags ...api.Tag)
	FGauge(stat string, value float64, tags ...api.Tag)
	FGaugeDelta(stat string, value float64, tags ...api.Tag)
	SetAdd(stat string, value string, tags ...api.Tag)
	Close() error
	GetLostPackets() int64
	GetLostBytes() int64
	GetSendQueueLen() int64
	GetBufPoolLen() int64
	GetSentPackets() int64
	GetSentBytes() int64
}

var (
	currentClient client // currentClient global package client
	usageSent     *usage // usageSent global package usage send stat
)

// Statsd metric driver.
// Implements metirc.Metric interface.
// Client is underlaying library.
type Statsd struct {
	prefix   string
	client   client
	mapLock  sync.Mutex
	prefixes map[string]*Statsd
}

// MustDial same as Dial but raises Fatal on error on failure
func MustDial(opts ...Option) {
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

// Dial connects to statsd server and set it as metric driver.
// - default statsd address is read from dcy or env STATSD_LOGGER_ADDRESS
// - default prefix is AppName.InstanceId
// Examples:
//   - Dial()
//   - Dial(statsd.MetricPrefix("my_app"))
//   - Dial(statsd.MetricPrefix("my_app"),	statsd.StatsDAddr( "127.0.0.1:8125"))
func Dial(opts ...Option) error {
	// default options
	o := &options{
		prefix:            fmt.Sprintf("%s.%s.", env.AppName(), env.InstanceId()),
		addr:              "",
		bufPoolCapacity:   DefaultBufPoolCapacity,
		sendQueueCapacity: DefaultSendQueueCapacity,
		sendLoopCount:     DefaultSendLoopCount,
	}

	// apply sent options
	for _, optFn := range opts {
		optFn(o)
	}

	// validate options
	if err := o.Validate(); nil != err {
		logger().Error(err)
		return err
	}

	// client with options
	// - prefix to default client is not set to be able to replace prefix later
	// this is performance penalty, consider remove WithPrefix option
	currentClient = api.NewClient(o.addr,
		//api.MetricPrefix(o.prefix)
		api.BufPoolCapacity(o.bufPoolCapacity),
		api.SendQueueCapacity(o.sendQueueCapacity),
		api.SendLoopCount(o.sendLoopCount),
		api.Logger(golog.New(&logRedir{packetsLostNotice: false}, "", 0)),
	)

	// default statsd is without prefix
	// - prefix is empty here to supprot replacing and appending prefixes
	woPrefix := newStatsd("", currentClient)

	//set statsd as metric dirver with default prefix
	metric.Set(woPrefix.WithPrefix(o.prefix))

	logger().S("addr", o.addr).S("prefix", o.prefix).Info("started")

	// statSent send every 5 sec
	usageSent = usageReport(5*time.Second, currentClient)
	return nil
}

func newStatsd(prefix string, client client) *Statsd {
	return &Statsd{
		client:   client,
		prefix:   prefix,
		prefixes: make(map[string]*Statsd),
	}
}

// Counter increments counter name for sum(values).
// If called witohout values will increment for 1.
func (i *Statsd) Counter(name string, values ...int) {
	value := 1
	if len(values) > 0 {
		value = 0
		for _, v := range values {
			value += v
		}
	}
	i.client.Incr(i.handlePrefix(name), int64(value))
	usageSent.Counter()
}

// Close client connection and flush all metric
func Close() {
	if nil == currentClient {
		return
	}
	usageSent.Close()
	currentClient.Close()
}

// Gauge submits/updates a statsd gauge type.
func (i *Statsd) Gauge(name string, value int) {
	i.client.Gauge(i.handlePrefix(name), int64(value))
	usageSent.Gauge()
}

// Timing measures execution time for f and submits it as statsd timing type.
func (i *Statsd) Timing(name string, f func()) {
	stopwatch := metric.NewStopwatch()
	f()
	duration := stopwatch.GetNs()
	i.Time(name, duration)
}

// Time submits a statsd timing type.
func (i *Statsd) Time(name string, duration int) {
	i.client.Timing(i.handlePrefix(name), int64(duration))
	usageSent.Timer()
}

func (i *Statsd) handlePrefix(name string) string {
	if "" == i.prefix {
		return name
	}
	return i.prefix + name
}

// WithPrefix returns the clone of the original metric, but with a different prefix
func (i *Statsd) WithPrefix(prefix string) metric.Metric {
	i.mapLock.Lock()
	defer i.mapLock.Unlock()
	s, ok := i.prefixes[prefix]
	if ok && s != nil {
		return s
	}

	// new prefix for cloned
	mPrefix := prefix
	if !strings.HasSuffix(mPrefix, ".") {
		mPrefix += "."
	}

	i.prefixes[prefix] = newStatsd(mPrefix, i.client)
	return i.prefixes[prefix]
}

// AppendSuffix returns the clone of the original metric, but with the
// suffix appended to the end of the original prefix
func (i *Statsd) AppendSuffix(suffix string) metric.Metric {
	return i.WithPrefix(i.prefix + suffix)
}

func logger() *log.Agregator {
	return log.S("lib", "svckit.metric").S("statsd", "github.com/minus5/go-statsd")
}

// logRedir simple io.Writer for log redirection to svckit log
type logRedir struct {
	packetsLostNotice bool
}

// Write writes to log output
func (lr *logRedir) Write(p []byte) (int, error) {
	msg := string(p)
	if strings.Contains(msg, "Error") {
		logger().ErrorS(msg)
	} else if strings.Contains(msg, "packets lost") && !lr.packetsLostNotice {
		logger().Notice(msg)
		lr.packetsLostNotice = true
	} else {
		logger().Info(msg)
	}
	return len(p), nil
}
