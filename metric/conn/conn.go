package conn

import (
	"github.com/minus5/svckit/httpi"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric/prometheus"
	"github.com/minus5/svckit/metric/statsd"
	"os"
)

func Connect(prefix string) func() {
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
	opts := []statsd.Option{}
	if prefix != "" {
		opts = append(opts, statsd.MetricPrefix(prefix))
	}
	statsd.MustDial(opts...)
	return statsd.Close
}
