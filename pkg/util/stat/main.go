package stat

import (
	"fmt"
	"log"
	"os"
	"pkg/util"

	"github.com/cactus/go-statsd-client/statsd"
)

type StatsdClient struct {
	client statsd.Statter
}

var client statsd.Statter

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

func Counter(name string) {
	IncCounter(name)
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

func Time(name string, f func()) {
	stopwatch := util.NewStopwatch()
	f()
	duration := stopwatch.GetMs()
	if client == nil {
		log.Printf("timer %s %8.4f ms", name, duration)
	} else {
		Timing(name, int64(duration))
	}
}

func TimeNs(name string, f func()) {
	stopwatch := util.NewStopwatch()
	f()
	duration := stopwatch.GetNs()
	if client == nil {
		log.Printf("timer %s %15.4f ns", name, duration)
	} else {
		Timing(name, int64(duration))
	}
}

func Inc(name string) {
	IncCounter(name)
}

type inf struct{}

func (*inf) Time(name string, f func()) {
	TimeNs(name, f)
}
func (*inf) TimeNs(name string, f func()) {
	TimeNs(name, f)
}
func (*inf) Gauge(name string, value int) {
	Gauge(name, int64(value))
}
func (*inf) Counter(name string) {
	IncCounter(name)
}

func Interface() *inf {
	return &inf{}
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}
