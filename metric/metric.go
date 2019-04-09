// Entry point for sending metrics.
//
// Depending on driver metric could be sent to various locations.
// Default is Noop driver which is doing nothing.
//
// For sending metrics to statsd application should do something like:
//      import 	"github.com/minus5/svckit/metric/statsd"
//  		err := statsd.Dial()
// That will sent statsd driver.
//
// For testing purpose it can be useful to Set own implementation of Metric.
// Implement Metric interface and set it through metric.Set(myImplementation).
//
// It is meant that all code uses global metric package methods:
//    metric.Counter("my_counter")
//    metric.Counter("my_counter", 12)
//    metric.Gauge("my_gauge", 123)
//    metric.Timing("my_gauge", func() { do_something } )
// Depending on the currently set driver metric will be sent ...somewhere.
package metric

var driver Metric

// Set the default driver.
func init() {
	driver = NewNoop()
}

// Interface definition for all metric implementations.
type Metric interface {
	Counter(name string, values ...int)
	Gauge(name string, value int)
	Timing(name string, f func())
	Time(name string, duration int)
	WithPrefix(prefix string) Metric
	AppendSuffix(suffix string) Metric
}

// Set driver, something wich implements Metric interface
func Set(d Metric) {
	driver = d
}

// Increments counter name for sum(values)
// If called witohout values will increment for 1
func Counter(name string, values ...int) {
	driver.Counter(name, values...)
}

// Submits/Updates a gauge type.
func Gauge(name string, value int) {
	driver.Gauge(name, value)
}

// Measures execution time for f and submits it as timing type.
func Timing(name string, f func()) {
	driver.Timing(name, f)
}

// Submits a statsd type.
func Time(name string, duration int) {
	driver.Time(name, duration)
}

// Returns a Metric with a different prefix
func WithPrefix(prefix string) Metric {
	return driver.WithPrefix(prefix)
}

// Returns a Metric with a suffix
// appended to the original prefix
func AppendSuffix(suffix string) Metric {
	return driver.AppendSuffix(suffix)
}
