package nsqx

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/minus5/go-nsqx"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"
)

type nsqLogger struct{}

// Ovdje ulaze logovi iz go-nsqx liba
// Logger interface is Output(calldepth int, s string) error — identical in go-nsq and go-nsqx
func (n *nsqLogger) Output(calldepth int, s string) error {
	a := log.NewAgregator(nil, calldepth)
	a.S("lib", "svckit.nsqx")
	if strings.HasPrefix(s, "INF") {
		a.Info(s)
		return nil
	}
	if strings.HasPrefix(s, "WRN") {
		if !strings.Contains(s, "there are 0 connections left alive") {
			a.Info(s)
		}
		return nil
	}
	if strings.HasPrefix(s, "ERR") {
		if !strings.Contains(s, "TOPIC_NOT_FOUND") {
			a.ErrorS(s)
		}
		return nil
	}
	a.Debug(s)
	return nil
}

type options struct {
	// Existing options (backward compatible)
	maxInFlight int
	concurrency int
	channel     string
	nsqdTCPAddr string
	logger      *nsqLogger
	logLevel    nsqx.LogLevel
	lookupds    dcy.Addresses

	// Timeouts
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration

	// Discovery
	lookupdPollInterval time.Duration
	lookupdPollJitter   float64
	lookupdCacheTTL     time.Duration
	lookupdPollTimeout  time.Duration

	// Adaptive RDY
	rdyEvalInterval      time.Duration
	rdySlowStartInitial  int64
	rdySlowStartInterval time.Duration
	idleConnectionRDY    int64
	emaAlpha             float64
	idleRateThreshold    float64
	idleTicksRequired    int

	// Retry & backoff
	maxAttempts             uint16
	backoffAlgorithm        nsqx.BackoffAlgo
	backoffBaseDelay        time.Duration
	backoffMaxDelay         time.Duration
	backoffMaxExponent      uint
	backoffFixedInterval    time.Duration
	backoffFixedMaxAttempts int

	// Circuit breaker
	circuitBreakerThreshold  int32
	circuitBreakerResetAfter time.Duration

	// Producer
	producerPipelines int
	batchMaxSize      int
	batchMaxDelay     time.Duration

	// Server side buffering
	heartbeatInterval   time.Duration
	outputBufferSize    int
	outputBufferTimeout time.Duration

	// Compression
	compression  nsqx.CompressionCodec
	deflateLevel int

	// Zero copy / ACK batching / adaptive buffer
	zeroCopyThreshold     int
	ackBatchSize          int
	ackBatchDelay         time.Duration
	adaptiveBufferEnabled *bool // pointer so we can detect "not set"
	adaptiveBufferInitial int
	adaptiveBufferMax     int

	// Identity
	clientID  string
	hostname  string
	userAgent string

	// Metrics
	metricsBackend       nsqx.MetricsBackend
	statsdAddr           string
	statsdPrefix         string
	prometheusNamespace  string
	metricsFlushInterval time.Duration

	// TLS
	tlsConfig *tls.Config
}

func (o *options) clone() *options {
	o2 := &options{}
	*o2 = *o
	return o2
}

func (o *options) set(opts ...func(*options)) {
	o.apply(opts...)
}

func (o *options) apply(opts ...func(*options)) *options {
	for _, fn := range opts {
		fn(o)
	}
	return o
}

func (o *options) Concurrency() int {
	if o.concurrency != 0 {
		return o.concurrency
	}
	return o.maxInFlight
}

// toNsqxConfig builds a go-nsqx Config from the wrapper options,
// applying only non zero overrides on top of go-nsqx defaults
func (o *options) toNsqxConfig() *nsqx.Config {
	cfg := nsqx.NewConfig()
	cfg.MaxInFlight = o.maxInFlight
	cfg.WorkerConcurrency = o.Concurrency()
	cfg.LogLevel = o.logLevel
	cfg.ZeroCopyThreshold = o.zeroCopyThreshold
	cfg.BackoffAlgorithm = o.backoffAlgorithm

	if o.dialTimeout > 0 {
		cfg.DialTimeout = o.dialTimeout
	}
	if o.readTimeout > 0 {
		cfg.ReadTimeout = o.readTimeout
	}
	if o.writeTimeout > 0 {
		cfg.WriteTimeout = o.writeTimeout
	}
	if o.lookupdPollInterval > 0 {
		cfg.LookupdPollInterval = o.lookupdPollInterval
	}
	if o.lookupdPollJitter > 0 {
		cfg.LookupdPollJitter = o.lookupdPollJitter
	}
	if o.lookupdCacheTTL > 0 {
		cfg.LookupdCacheTTL = o.lookupdCacheTTL
	}
	if o.lookupdPollTimeout > 0 {
		cfg.LookupdPollTimeout = o.lookupdPollTimeout
	}
	if o.rdyEvalInterval > 0 {
		cfg.RDYEvalInterval = o.rdyEvalInterval
	}
	if o.rdySlowStartInitial > 0 {
		cfg.RDYSlowStartInitial = o.rdySlowStartInitial
	}
	if o.rdySlowStartInterval > 0 {
		cfg.RDYSlowStartInterval = o.rdySlowStartInterval
	}
	if o.idleConnectionRDY > 0 {
		cfg.IdleConnectionRDY = o.idleConnectionRDY
	}
	if o.emaAlpha > 0 {
		cfg.EMAAlpha = o.emaAlpha
	}
	if o.idleRateThreshold > 0 {
		cfg.IdleRateThreshold = o.idleRateThreshold
	}
	if o.idleTicksRequired > 0 {
		cfg.IdleTicksRequired = o.idleTicksRequired
	}
	if o.maxAttempts > 0 {
		cfg.MaxAttempts = o.maxAttempts
	}
	if o.backoffBaseDelay > 0 {
		cfg.BackoffBaseDelay = o.backoffBaseDelay
	}
	if o.backoffMaxDelay > 0 {
		cfg.BackoffMaxDelay = o.backoffMaxDelay
	}
	if o.backoffMaxExponent > 0 {
		cfg.BackoffMaxExponent = o.backoffMaxExponent
	}
	if o.backoffFixedInterval > 0 {
		cfg.BackoffFixedInterval = o.backoffFixedInterval
	}
	if o.backoffFixedMaxAttempts > 0 {
		cfg.BackoffFixedMaxAttempts = o.backoffFixedMaxAttempts
	}
	if o.circuitBreakerThreshold > 0 {
		cfg.CircuitBreakerThreshold = o.circuitBreakerThreshold
	}
	if o.circuitBreakerResetAfter > 0 {
		cfg.CircuitBreakerResetAfter = o.circuitBreakerResetAfter
	}
	if o.producerPipelines > 0 {
		cfg.ProducerPipelines = o.producerPipelines
	}
	if o.batchMaxSize > 0 {
		cfg.BatchMaxSize = o.batchMaxSize
	}
	if o.batchMaxDelay > 0 {
		cfg.BatchMaxDelay = o.batchMaxDelay
	}
	if o.heartbeatInterval > 0 {
		cfg.HeartbeatInterval = o.heartbeatInterval
	}
	if o.outputBufferSize != 0 {
		cfg.OutputBufferSize = o.outputBufferSize
	}
	if o.outputBufferTimeout != 0 {
		cfg.OutputBufferTimeout = o.outputBufferTimeout
	}
	if o.compression != 0 {
		cfg.Compression = o.compression
	}
	if o.deflateLevel > 0 {
		cfg.DeflateLevel = o.deflateLevel
	}
	if o.ackBatchSize > 0 {
		cfg.AckBatchSize = o.ackBatchSize
	}
	if o.ackBatchDelay > 0 {
		cfg.AckBatchDelay = o.ackBatchDelay
	}
	if o.adaptiveBufferEnabled != nil {
		cfg.AdaptiveBufferEnabled = *o.adaptiveBufferEnabled
	}
	if o.adaptiveBufferInitial > 0 {
		cfg.AdaptiveBufferInitial = o.adaptiveBufferInitial
	}
	if o.adaptiveBufferMax > 0 {
		cfg.AdaptiveBufferMax = o.adaptiveBufferMax
	}
	if o.clientID != "" {
		cfg.ClientID = o.clientID
	}
	if o.hostname != "" {
		cfg.Hostname = o.hostname
	}
	if o.userAgent != "" {
		cfg.UserAgent = o.userAgent
	}
	if o.metricsBackend != 0 {
		cfg.MetricsBackend = o.metricsBackend
	}
	if o.statsdAddr != "" {
		cfg.StatsdAddr = o.statsdAddr
	}
	if o.statsdPrefix != "" {
		cfg.StatsdPrefix = o.statsdPrefix
	}
	if o.prometheusNamespace != "" {
		cfg.PrometheusNamespace = o.prometheusNamespace
	}
	if o.metricsFlushInterval > 0 {
		cfg.MetricsFlushInterval = o.metricsFlushInterval
	}
	if o.tlsConfig != nil {
		cfg.TLSConfig = o.tlsConfig
	}

	return cfg
}

// Existing options (backward compatible go-nsq/go-nsqx)

func MaxInFlight(m int) func(*options) { return func(o *options) { o.maxInFlight = m } }
func Channel(c string) func(*options)  { return func(o *options) { o.channel = c } }
func Concurrency(c int) func(*options) { return func(o *options) { o.concurrency = c } }
func Ordered() func(*options)          { return func(o *options) { o.concurrency = 1 } }
func LogLevelDebug() func(*options)    { return func(o *options) { o.logLevel = nsqx.LogLevelDebug } }

// New options

func DialTimeout(d time.Duration) func(*options)  { return func(o *options) { o.dialTimeout = d } }
func ReadTimeout(d time.Duration) func(*options)  { return func(o *options) { o.readTimeout = d } }
func WriteTimeout(d time.Duration) func(*options) { return func(o *options) { o.writeTimeout = d } }
func LookupdPollInterval(d time.Duration) func(*options) {
	return func(o *options) { o.lookupdPollInterval = d }
}
func LookupdPollJitter(f float64) func(*options) { return func(o *options) { o.lookupdPollJitter = f } }
func LookupdCacheTTL(d time.Duration) func(*options) {
	return func(o *options) { o.lookupdCacheTTL = d }
}
func LookupdPollTimeout(d time.Duration) func(*options) {
	return func(o *options) { o.lookupdPollTimeout = d }
}
func RDYEvalInterval(d time.Duration) func(*options) {
	return func(o *options) { o.rdyEvalInterval = d }
}
func RDYSlowStartInitial(n int64) func(*options) {
	return func(o *options) { o.rdySlowStartInitial = n }
}
func RDYSlowStartInterval(d time.Duration) func(*options) {
	return func(o *options) { o.rdySlowStartInterval = d }
}
func IdleConnectionRDY(n int64) func(*options)   { return func(o *options) { o.idleConnectionRDY = n } }
func EMAAlpha(f float64) func(*options)          { return func(o *options) { o.emaAlpha = f } }
func IdleRateThreshold(f float64) func(*options) { return func(o *options) { o.idleRateThreshold = f } }
func IdleTicksRequired(n int) func(*options)     { return func(o *options) { o.idleTicksRequired = n } }
func MaxAttempts(n uint16) func(*options)        { return func(o *options) { o.maxAttempts = n } }
func BackoffAlgorithm(a nsqx.BackoffAlgo) func(*options) {
	return func(o *options) { o.backoffAlgorithm = a }
}
func BackoffBaseDelay(d time.Duration) func(*options) {
	return func(o *options) { o.backoffBaseDelay = d }
}
func BackoffMaxDelay(d time.Duration) func(*options) {
	return func(o *options) { o.backoffMaxDelay = d }
}
func BackoffMaxExponent(n uint) func(*options) { return func(o *options) { o.backoffMaxExponent = n } }
func BackoffFixedInterval(d time.Duration) func(*options) {
	return func(o *options) { o.backoffFixedInterval = d }
}
func BackoffFixedMaxAttempts(n int) func(*options) {
	return func(o *options) { o.backoffFixedMaxAttempts = n }
}
func CircuitBreakerThreshold(n int32) func(*options) {
	return func(o *options) { o.circuitBreakerThreshold = n }
}
func CircuitBreakerResetAfter(d time.Duration) func(*options) {
	return func(o *options) { o.circuitBreakerResetAfter = d }
}
func ProducerPipelines(n int) func(*options)       { return func(o *options) { o.producerPipelines = n } }
func BatchMaxSize(n int) func(*options)            { return func(o *options) { o.batchMaxSize = n } }
func BatchMaxDelay(d time.Duration) func(*options) { return func(o *options) { o.batchMaxDelay = d } }
func HeartbeatInterval(d time.Duration) func(*options) {
	return func(o *options) { o.heartbeatInterval = d }
}
func OutputBufferSize(n int) func(*options) { return func(o *options) { o.outputBufferSize = n } }
func OutputBufferTimeout(d time.Duration) func(*options) {
	return func(o *options) { o.outputBufferTimeout = d }
}
func Compression(c nsqx.CompressionCodec) func(*options) {
	return func(o *options) { o.compression = c }
}
func DeflateLevel(n int) func(*options)            { return func(o *options) { o.deflateLevel = n } }
func ZeroCopyThreshold(n int) func(*options)       { return func(o *options) { o.zeroCopyThreshold = n } }
func AckBatchSize(n int) func(*options)            { return func(o *options) { o.ackBatchSize = n } }
func AckBatchDelay(d time.Duration) func(*options) { return func(o *options) { o.ackBatchDelay = d } }
func AdaptiveBufferEnabled(b bool) func(*options) {
	return func(o *options) { o.adaptiveBufferEnabled = &b }
}
func AdaptiveBufferInitial(n int) func(*options) {
	return func(o *options) { o.adaptiveBufferInitial = n }
}
func AdaptiveBufferMax(n int) func(*options) { return func(o *options) { o.adaptiveBufferMax = n } }
func ClientID(s string) func(*options)       { return func(o *options) { o.clientID = s } }
func Hostname(s string) func(*options)       { return func(o *options) { o.hostname = s } }
func UserAgent(s string) func(*options)      { return func(o *options) { o.userAgent = s } }
func MetricsPrometheus(namespace string) func(*options) {
	return func(o *options) { o.metricsBackend = nsqx.MetricsPrometheus; o.prometheusNamespace = namespace }
}
func MetricsStatsd(addr, prefix string) func(*options) {
	return func(o *options) { o.metricsBackend = nsqx.MetricsStatsd; o.statsdAddr = addr; o.statsdPrefix = prefix }
}
func MetricsFlushInterval(d time.Duration) func(*options) {
	return func(o *options) { o.metricsFlushInterval = d }
}
func TLSConfig(cfg *tls.Config) func(*options) { return func(o *options) { o.tlsConfig = cfg } }
