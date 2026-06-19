// Prometheus metric driver.
//
// Usage:
//
//	import "github.com/minus5/svckit/metric/prometheus"
//
//	prometheus.Init()
//	http.Handle("/metrics", prometheus.Handler())
package prometheus

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/minus5/svckit/metric"
)

var defaultRegistry *prom.Registry

// Prometheus metric driver.
// Implements metric.Metric interface.
type Prometheus struct {
	prefix string
	shared *shared
}

type shared struct {
	registry      *prom.Registry
	mu            sync.Mutex
	counters      map[string]prom.Counter
	gauges        map[string]prom.Gauge
	histograms    map[string]prom.Histogram
	counterVecs   map[string]*prom.CounterVec
	gaugeVecs     map[string]*prom.GaugeVec
	histogramVecs map[string]*prom.HistogramVec
}

// Init creates the Prometheus driver and sets it as the global metric driver.
func Init() {
	defaultRegistry = prom.NewRegistry()
	s := &shared{
		registry:      defaultRegistry,
		counters:      make(map[string]prom.Counter),
		gauges:        make(map[string]prom.Gauge),
		histograms:    make(map[string]prom.Histogram),
		counterVecs:   make(map[string]*prom.CounterVec),
		gaugeVecs:     make(map[string]*prom.GaugeVec),
		histogramVecs: make(map[string]*prom.HistogramVec),
	}
	metric.Set(&Prometheus{shared: s})
}

// Handler returns an http.Handler exposing Prometheus metrics for scraping.
func Handler() http.Handler {
	return promhttp.HandlerFor(defaultRegistry, promhttp.HandlerOpts{})
}

func sanitizeName(s string) string {
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

func (p *Prometheus) key(name string) string {
	return sanitizeName(p.prefix + name)
}

func (p *Prometheus) getCounter(name string) prom.Counter {
	k := p.key(name)
	p.shared.mu.Lock()
	defer p.shared.mu.Unlock()
	if c, ok := p.shared.counters[k]; ok {
		return c
	}
	c := prom.NewCounter(prom.CounterOpts{Name: k, Help: k})
	p.shared.registry.MustRegister(c)
	p.shared.counters[k] = c
	return c
}

// Counter increments the named counter by sum(values), defaulting to 1.
func (p *Prometheus) Counter(name string, values ...int) {
	value := 1
	if len(values) > 0 {
		value = 0
		for _, v := range values {
			value += v
		}
	}
	p.getCounter(name).Add(float64(value))
}

func (p *Prometheus) getGauge(name string) prom.Gauge {
	k := p.key(name)
	p.shared.mu.Lock()
	defer p.shared.mu.Unlock()
	if g, ok := p.shared.gauges[k]; ok {
		return g
	}
	g := prom.NewGauge(prom.GaugeOpts{Name: k, Help: k})
	p.shared.registry.MustRegister(g)
	p.shared.gauges[k] = g
	return g
}

// Gauge sets the named gauge to value.
func (p *Prometheus) Gauge(name string, value int) {
	p.getGauge(name).Set(float64(value))
}

func (p *Prometheus) getHistogram(name string) prom.Histogram {
	k := p.key(name)
	p.shared.mu.Lock()
	defer p.shared.mu.Unlock()
	if h, ok := p.shared.histograms[k]; ok {
		return h
	}
	h := prom.NewHistogram(prom.HistogramOpts{
		Name:    k,
		Help:    k,
		Buckets: prom.DefBuckets,
	})
	p.shared.registry.MustRegister(h)
	p.shared.histograms[k] = h
	return h
}

// Timing measures execution time of f and records it as a histogram in seconds.
func (p *Prometheus) Timing(name string, f func()) {
	sw := metric.NewStopwatch()
	f()
	p.Time(name, sw.GetNs())
}

// Time records duration (nanoseconds) as a histogram observation in seconds.
func (p *Prometheus) Time(name string, duration int) {
	p.getHistogram(name).Observe(float64(duration) / 1e9)
}

func sortedLabelNames(labels map[string]string) []string {
	names := make([]string, 0, len(labels))
	for k := range labels {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (p *Prometheus) getCounterVec(name string, labelNames []string) *prom.CounterVec {
	k := p.key(name)
	p.shared.mu.Lock()
	defer p.shared.mu.Unlock()
	if cv, ok := p.shared.counterVecs[k]; ok {
		return cv
	}
	cv := prom.NewCounterVec(prom.CounterOpts{Name: k, Help: k}, labelNames)
	p.shared.registry.MustRegister(cv)
	p.shared.counterVecs[k] = cv
	return cv
}

// CounterL increments the named counter with labels by sum(values), defaulting to 1.
func (p *Prometheus) CounterL(name string, labels map[string]string, values ...int) {
	value := 1
	if len(values) > 0 {
		value = 0
		for _, v := range values {
			value += v
		}
	}
	p.getCounterVec(name, sortedLabelNames(labels)).With(prom.Labels(labels)).Add(float64(value))
}

func (p *Prometheus) getGaugeVec(name string, labelNames []string) *prom.GaugeVec {
	k := p.key(name)
	p.shared.mu.Lock()
	defer p.shared.mu.Unlock()
	if gv, ok := p.shared.gaugeVecs[k]; ok {
		return gv
	}
	gv := prom.NewGaugeVec(prom.GaugeOpts{Name: k, Help: k}, labelNames)
	p.shared.registry.MustRegister(gv)
	p.shared.gaugeVecs[k] = gv
	return gv
}

// GaugeL sets the named gauge with labels to value.
func (p *Prometheus) GaugeL(name string, labels map[string]string, value int) {
	p.getGaugeVec(name, sortedLabelNames(labels)).With(prom.Labels(labels)).Set(float64(value))
}

func (p *Prometheus) getHistogramVec(name string, labelNames []string) *prom.HistogramVec {
	k := p.key(name)
	p.shared.mu.Lock()
	defer p.shared.mu.Unlock()
	if hv, ok := p.shared.histogramVecs[k]; ok {
		return hv
	}
	hv := prom.NewHistogramVec(prom.HistogramOpts{
		Name:    k,
		Help:    k,
		Buckets: prom.DefBuckets,
	}, labelNames)
	p.shared.registry.MustRegister(hv)
	p.shared.histogramVecs[k] = hv
	return hv
}

// TimeL records duration (nanoseconds) with labels as a histogram observation in seconds.
func (p *Prometheus) TimeL(name string, labels map[string]string, duration int) {
	p.getHistogramVec(name, sortedLabelNames(labels)).With(prom.Labels(labels)).Observe(float64(duration) / 1e9)
}

// WithPrefix returns a Metric with the given prefix.
func (p *Prometheus) WithPrefix(prefix string) metric.Metric {
	if !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}
	return &Prometheus{prefix: prefix, shared: p.shared}
}

// AppendSuffix returns a Metric with suffix appended to the current prefix.
func (p *Prometheus) AppendSuffix(suffix string) metric.Metric {
	return p.WithPrefix(p.prefix + suffix)
}
