package nsqx

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/minus5/go-nsqx"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"
)

const (
	DefaultMaxInFlight          = 256
	DefaultConcurrency          = 16
	LookupdHTTPServiceName      = "nsqlookupd-http"
	LookupdHTTPServiceNameByTag = "nsqlookupd"
	LookupdHTTPServiceTag       = "http"
	EnvNsqd                     = "SVCKIT_NSQD"
	DefaultMsgTouchInterval     = time.Second * 30
	DefaultStatsdPrefix         = "nsq." // to keep the same as nsqd prefix, maybe change for diff?
)

var (
	Pub = MustNewProducer
	Sub = MustNewConsumer

	defaults  *options
	initMu    sync.Mutex
	discovery *nsqx.Discovery // shared across all consumers
)

func getDefaults() *options {
	initMu.Lock()
	defer initMu.Unlock()
	if defaults == nil {
		initDefaults()
	}
	return defaults
}

func Set(opts ...func(*options)) {
	initMu.Lock()
	defer initMu.Unlock()
	if defaults == nil {
		initDefaults()
	}
	defaults.apply(opts...)
}

func initDefaults() {
	defaults = &options{
		maxInFlight:          DefaultMaxInFlight,
		concurrency:          DefaultConcurrency,
		channel:              fmt.Sprintf("%s-%s", env.AppName(), env.InstanceId()),
		nsqdTCPAddr:          "127.0.0.1:4150",
		lookupds:             dcy.Addresses{dcy.Address{Address: "127.0.0.1", Port: 4161}},
		logLevel:             nsqx.LogLevelWarning,
		zeroCopyThreshold:    0,                 // always copy, so that messages are not truncated
		backoffAlgorithm:     nsqx.BackoffFixed, // Fixed backoff every FixedInterval
		backoffFixedInterval: 2 * time.Second,
		logger:               &nsqLogger{},
		lookupdPollInterval:  15 * time.Second, // see how will this behave in prod, current nsq pulls every 10s from consul
		lookupdCacheTTL:      10 * time.Second,
		lookupdPollTimeout:   5 * time.Second,
	}
	if e, ok := os.LookupEnv(EnvNsqd); ok && e != "" {
		defaults.nsqdTCPAddr = e
		logger().S("nsqd", defaults.nsqdTCPAddr).Debug("init nsqd")
	}
	connect := func() error {
		var addrs dcy.Addresses
		a, err := dcy.Services(LookupdHTTPServiceName)
		if err != nil && err != dcy.ErrNotFound {
			logger().Error(err)
		}
		if err == nil {
			addrs.Append(a)
		}
		a, err = dcy.ServicesByTag(LookupdHTTPServiceNameByTag, LookupdHTTPServiceTag)
		if err == nil {
			addrs.Append(a)
		}
		if err != nil && err != dcy.ErrNotFound {
			logger().Error(err)
		}
		if len(addrs) == 0 {
			return dcy.ErrNotFound
		}
		defaults.lookupds = addrs
		logger().S("lookupds", fmt.Sprintf("%v", defaults.lookupds.String())).Info("init lookupds")
		return nil
	}
	if err := signal.WithExponentialBackoff(connect); err != nil {
		logger().Fatal(err)
	}
	// statsd metrics for every topic/channel, new with nsqx
	if addr, ok := os.LookupEnv("STATSD_LOGGER_ADDRESS"); ok && addr != "" {
		defaults.statsdAddr = addr
		defaults.metricsBackend = nsqx.MetricsStatsd
		defaults.statsdPrefix = fmt.Sprintf("%s.", env.AppName())
		logger().S("statsd", defaults.statsdAddr).Info("init statsd")
	}
}

// getDiscovery returns the shared Discovery instance
func getDiscovery(o *options) *nsqx.Discovery {
	initMu.Lock()
	defer initMu.Unlock()
	if discovery == nil {
		cfg := o.toNsqxConfig()
		discovery = nsqx.NewDiscovery(o.lookupds.String(), cfg, o.logger)
	}
	return discovery
}

func logger() *log.Agregator {
	return log.S("lib", "svckit.nsqx")
}

// ChannelAppName sets default channel name to app name
func ChannelAppName() {
	Set(Channel(env.AppName()))
}

// ChannelEphemeral sets default channel name to app name suffixed with node name and #ephemeral
func ChannelEphemeral() {
	Set(Channel(fmt.Sprintf("%s-%s#ephemeral", env.AppName(), env.InstanceId())))
}

func DefaultChannel(c string) {
	Set(Channel(c))
}

func every(duration time.Duration, work func()) chan struct{} {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(duration)
		for {
			select {
			case <-ticker.C:
				work()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
	return stop
}
