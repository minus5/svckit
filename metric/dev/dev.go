package dev

import (
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
)

func Init() {
	metric.Set(NewDev())
}

type Dev struct{}

func (*Dev) Counter(name string, values ...int) {
	return
}

func (*Dev) Gauge(name string, value int) {
	return
}

func (*Dev) Timing(name string, f func()) {
	stopwatch := metric.NewStopwatch()
	f()
	duration := stopwatch.GetNs()
	log.I("ns", duration).I("ms", duration/1e6).Debug(name)
}

func (*Dev) Time(name string, duration int) {
	return
}

func (*Dev) WithPrefix(prefix string) metric.Metric {
	return &Dev{}
}

func (*Dev) AppendSuffix(suffix string) metric.Metric {
	return &Dev{}
}

func NewDev() *Dev {
	return &Dev{}
}
