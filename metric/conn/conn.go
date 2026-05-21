package conn

import (
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/httpi"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric/prometheus"
	"github.com/minus5/svckit/metric/statsd"
	"os"
)

func Connect() func() {
	defer func() {
		log.Info("metrics initialized")
	}()
	if _, ok := os.LookupEnv("SVCKIT_METRIC_PROMETHEUS"); ok {
		log.S("target", "prometheus").Info("metrics init")
		prometheus.Init()
		httpi.Handle("/metrics", prometheus.Handler())
		return func() {}
	}
	log.S("target", "statsd").Info("metrics init")
	statsd.Dial(statsd.MetricPrefix(env.AppName()))
	return statsd.Close
}
