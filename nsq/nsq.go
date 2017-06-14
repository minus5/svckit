// TODO - objasni dependency na consul u kojem moraju biti navedeni lookup-i
// i to pod ServiceDiscoveryLookupdHttp key-om
// to je konvencija i ne moze se kofigurirati !!!
package nsq

import (
	"fmt"
	"os"
	"sync"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"

	gonsq "github.com/nsqio/go-nsq"
)

const (
	DefaultMaxInFlight     = 256
	LookupdHTTPServiceName = "nsqlookupd-http"
	NsqdTCPServiceName     = "nsqd-tcp"
	EnvNsqd                = "SVCKIT_NSQD"
)

var (
	//Aliasi za lijepsi api
	Pub = MustNewProducer
	Sub = MustNewConsumer

	defaults *options
	initMu   sync.Mutex
)

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
		channel:     fmt.Sprintf("%s-%s", env.AppName(), env.NodeName()),
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
	Set(Channel(fmt.Sprintf("%s-%s#ephemeral", env.AppName(), env.NodeName())))
}

func DefaultChannel(c string) {
	Set(Channel(c))
}
