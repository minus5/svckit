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
		maxInFlight: DefaultMaxInFlight,
		concurrency: DefaultConcurrency,
		channel:     fmt.Sprintf("%s-%s", env.AppName(), env.InstanceId()),
		nsqdTCPAddr: "127.0.0.1:4150",
		lookupds:    dcy.Addresses{dcy.Address{Address: "127.0.0.1", Port: 4161}},
		logLevel:    nsqx.LogLevelWarning,
		logger:      &nsqLogger{},
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
		logger().S("lookupds", fmt.Sprintf("%v", defaults.lookupds.String())).Debug("init lookupds")
		return nil
	}
	if err := signal.WithExponentialBackoff(connect); err != nil {
		logger().Fatal(err)
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
