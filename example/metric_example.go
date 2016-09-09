// +build metric_example

package main

import (
	"github.com/minus5/svckit/metric"
	"time"
)

const (
	counter = "counter"
	gauge   = "gauge"
	timing  = "timing"
)

func main() {
	metric.Counter(counter)
	metric.Gauge(gauge, 123)
	metric.Timing(timing, func() {
		time.Sleep(1e8)
	})
}
