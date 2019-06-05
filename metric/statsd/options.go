package statsd

import (
	"os"
	"strings"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/signal"
)

// options is set of configurable options
type options struct {
	addr              string
	prefix            string
	bufPoolCapacity   int
	sendQueueCapacity int
	sendLoopCount     int
	maxPacketSize     int
}

var staticStatsdEnvVars = []string{
	"SVCKIT_METRIC_STATSD",
	"STATSD_LOGGER_ADDRESS",
}

// Validate options before start
func (o *options) Validate() error {
	// for "github.com/smira/go-statsd" prefix must end with "."
	if "" != o.prefix && !strings.HasSuffix(o.prefix, ".") {
		o.prefix += "."
	}

	// check address
	if "" != o.addr {
		return nil
	}

	// get statsd address from env variable
	for _, envVar := range staticStatsdEnvVars {
		if os.Getenv(envVar) != "" {
			o.addr = os.Getenv(envVar)
			return nil
		}
	}

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
	o.addr = a.String()
	return nil
}

// Option is type for option implemtation
type Option func(o *options)

// StatsDAddr statsd server address in "host:port" format
func StatsDAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// MetricPrefix is prefix to prepend to every metric being sent
func MetricPrefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

// BufPoolCapacity controls size of pre-allocated buffer cache
func BufPoolCapacity(capacity int) Option {
	return func(o *options) {
		o.bufPoolCapacity = capacity
	}
}

// SendQueueCapacity controls length of the queue of packet ready to be sent
func SendQueueCapacity(capacity int) Option {
	return func(o *options) {
		o.sendQueueCapacity = capacity
	}
}

// SendLoopCount controls number of goroutines sending UDP packets
//
// Default value is 1, so packets are sent from single goroutine, this
// value might need to be bumped under high load
func SendLoopCount(threads int) Option {
	return func(o *options) {
		o.sendLoopCount = threads
	}
}
