package health

import (
	"expvar"
	"net/http"
	"runtime"
	"time"

	"github.com/mnu5/svckit/env"
	"github.com/mnu5/svckit/log"
	"github.com/mnu5/svckit/metric"
)

// Allowed values for status
const (
	Passing = Status(0)
	Warn    = Status(1)
	Fail    = Status(2)
)

const (
	// when not passing status is longer then this interval send notification
	notificationAfter = 30 * time.Second
)

func init() {
	checkCh = make(chan bool)
	handler = notImplemented
	lastPassingCheck = time.Now()
	go loop()
	expvar.Publish("svckit.health", expvar.Func(func() interface{} {
		stats := struct {
			Status string    `json:"status"`
			Note   string    `json:"note"`
			Time   time.Time `json:"time"`
		}{
			Status: status.String(),
			Note:   string(note),
			Time:   checkTime,
		}
		return stats
	}))
}

// Status represents service status
type Status int

// Add connects twu statsu values
func (s *Status) Add(s2 Status) {
	if s2 > *s {
		*s = s2
	}
}

var (
	handler          func() (Status, []byte)
	status           Status
	note             []byte
	checkCh          chan bool
	checkTime        time.Time
	lastPassingCheck time.Time
	notificationSent bool
)

// ToHtmlStatus converts status to Consul frendly http status.
// Reference: https://www.consul.io/docs/agent/checks.html
//   any 2xx code is considered passing,
//   a 429 Too Many Requests is a warning,
//   and anything else is a failure
func (s Status) ToHtmlStatus() int {
	switch s {
	case Passing:
		return http.StatusOK
	case Warn:
		return http.StatusTooManyRequests
	}
	return http.StatusInternalServerError
}

func (s Status) String() string {
	switch s {
	case Passing:
		return "passing"
	case Warn:
		return "warn"
	}
	return "fail"
}

// HttpHandler exposes status to http
func HttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Application", env.AppName())
	w.WriteHeader(status.ToHtmlStatus())
	w.Write([]byte(note))
}

func notImplemented() (Status, []byte) {
	return Fail, []byte("health check handler not implemented")
}

// Set the health check handler
func Set(h func() (Status, []byte)) {
	handler = h
	check()
}

// Get the current health status
func Get() (Status, []byte) {
	return status, note
}

// Run starts health check.
func Run() {
	checkCh <- true
}

func loop() {
	for {
		select {
		case <-checkCh:
			check()
		case <-time.After(10 * time.Second):
			check()
		}
	}
}

func check() {
	status, note = handler()
	checkTime = time.Now()
	sendNotification()
	sendMetric()
}

func sendNotification() {
	if status != Passing {
		if lastPassingCheck.Before(time.Now().Add(-notificationAfter)) && !notificationSent {
			logger().S("status", status.String()).Jc("note", note).Notice("health check failed")
			notificationSent = true
		} else {
			logger().S("status", status.String()).Jc("note", note).Info("health check failed")
		}
		return
	}
	if notificationSent {
		logger().S("status", status.String()).Jc("note", note).Notice("health check OK")
		notificationSent = false
	}
	lastPassingCheck = time.Now()
}

var ms = &runtime.MemStats{}

func sendMetric() {
	switch status {
	case Passing:
		metric.Gauge("health.passing", 1)
		metric.Gauge("health.warn", 0)
		metric.Gauge("health.fail", 0)
	case Warn:
		metric.Gauge("health.passing", 0)
		metric.Gauge("health.warn", 1)
		metric.Gauge("health.fail", 0)
	default:
		metric.Gauge("health.passing", 0)
		metric.Gauge("health.warn", 0)
		metric.Gauge("health.fail", 1)
	}
	runtime.ReadMemStats(ms)
	metric.Gauge("runtime.Sys", int(ms.Sys))
	metric.Gauge("runtime.Alloc", int(ms.Alloc))
	metric.Gauge("runtime.HeapSys", int(ms.HeapSys))
	metric.Gauge("runtime.HeapInuse", int(ms.HeapInuse))
	metric.Gauge("runtime.StackSys", int(ms.StackSys))
	metric.Gauge("runtime.StackInuse", int(ms.StackInuse))
	metric.Gauge("runtime.NumGC", int(ms.NumGC))
	metric.Gauge("runtime.NumGoroutine", runtime.NumGoroutine())
	//metric.Gauge("runtime.NumCgoCall", runtime.NumCgoCall())
}

func logger() *log.Agregator {
	return log.S("lib", "svckit.health")
}
