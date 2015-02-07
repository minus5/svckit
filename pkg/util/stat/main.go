package stat

import (
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"log"
	"os"
)

type StatsdClient struct {
	client *statsd.Client
}

var client *statsd.Client

//Initialize connection to statsd server.
func Init(addr, prefix string, includeHostNameInPrefix bool) error {
	client = nil
	if addr == "" {
		return nil
	}
	if includeHostNameInPrefix {
		prefix = fmt.Sprintf("%s.%s", prefix, hostname())
	}
	c, err := statsd.Dial(addr, prefix)
	if err != nil {
		log.Printf("error connecting to statsd %s", err)
		return err
	}
	client = c
	return nil
}

func IncCounter(name string) {
	if client != nil {
		err := client.Inc(name, 1, 1)
		if err != nil {
			log.Printf("statsd error %s", err)
		}
	}
}

func Gauge(name string, value int64) {
	if client != nil {
		err := client.Gauge(name, value, 1)
		if err != nil {
			log.Printf("statsd error %s", err)
		}
	}
}

func Timing(name string, value int64) {
	if client != nil {
		err := client.Timing(name, value, 1)
		if err != nil {
			log.Printf("statsd error %s", err)
		}
	}
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

