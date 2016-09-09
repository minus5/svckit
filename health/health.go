package health

import (
	"expvar"
	"net/http"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
	"time"
)

const (
	Passing = Status(0)
	Warn    = Status(1)
	Fail    = Status(2)
)

func init() {
	checkCh = make(chan bool)
	handler = notImplemented
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

type Status int

func (s *Status) Add(s2 Status) {
	if s2 > *s {
		*s = s2
	}
}

var (
	handler   func() (Status, []byte)
	status    Status
	note      []byte
	checkCh   chan bool
	checkTime time.Time
)

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

func HttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Application", env.AppName())
	w.WriteHeader(status.ToHtmlStatus())
	w.Write([]byte(note))
}

func notImplemented() (Status, []byte) {
	return Fail, []byte("health check handler not implemented")
}

func Set(h func() (Status, []byte)) {
	handler = h
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
	if status != Passing {
		logger().S("status", status.String()).Jc("note", note).Notice("health check failed")
	}
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
}

func logger() *log.Agregator {
	return log.S("lib", "svckit.health")
}
