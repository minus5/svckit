// TODO - objasni dependency na consul u kojem moraju biti navedeni lookup-i
// i to pod ServiceDiscoveryLookupdHttp key-om
// to je konvencija i ne moze se kofigurirati !!!
package nsq

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"

	gonsq "github.com/nsqio/go-nsq"
)

const (
	DefaultMaxInFlight      = 256
	DefaultConcurrency      = 16
	LookupdHTTPServiceName  = "nsqlookupd-http"
	NsqdTCPServiceName      = "nsqd-tcp"
	EnvNsqd                 = "SVCKIT_NSQD"
	DefaultMsgTouchInterval = time.Second * 30
)

var (
	//Aliasi za lijepsi api
	Pub = MustNewProducer
	Sub = MustNewConsumer

	defaults *options
	initMu   sync.Mutex
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
		logLevel:    gonsq.LogLevelWarning,
		logger:      &nsqLogger{},
	}
	if e, ok := os.LookupEnv(EnvNsqd); ok && e != "" {
		defaults.nsqdTCPAddr = e
		logger().S("nsqd", defaults.nsqdTCPAddr).Debug("init nsqd")
	}
	connect := func() error {
		addrs, err := dcy.Services(LookupdHTTPServiceName)
		if err != nil {
			logger().Error(err)
			return err
		}
		defaults.lookupds = addrs
		logger().S("lookupds", fmt.Sprintf("%v", defaults.lookupds.String())).Debug("init lookupds")
		return nil
	}
	if err := signal.WithExponentialBackoff(connect); err != nil {
		logger().Fatal(err)
	}
}

func logger() *log.Agregator {
	return log.S("lib", "svckit.nsq")
}

// ChannelAppName sets default channel name to app name.
// Default is app name suffixed with node name.
func ChannelAppName() {
	Set(Channel(env.AppName()))
}

// ChannelEphemeral sets default channel name to app name suffixed with node name and #ephemeral.
// Default is app name suffixed with node name.
func ChannelEphemeral() {
	Set(Channel(fmt.Sprintf("%s-%s#ephemeral", env.AppName(), env.InstanceId())))
}

func DefaultChannel(c string) {
	Set(Channel(c))
}

// Run the function work every duration
// Returns stop chan close which client needs to close.
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
