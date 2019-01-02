package httpi

import (
	"fmt"
	"net/http"
	"github.com/mnu5/svckit/log"
	"time"

	"github.com/codegangsta/negroni"
)

type RequestLogger struct{}

func NewRequestLogger() *RequestLogger {
	return &RequestLogger{}
}

func (l *RequestLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	//logger().S("method", r.Method).S("url", r.URL.Path).Info("started")
	start := time.Now()
	next(rw, r)
	duration := time.Since(start)
	res := rw.(negroni.ResponseWriter)

	logger().S("method", r.Method).
		S("url", r.URL.Path).
		S("status", http.StatusText(res.Status())).
		I("code", res.Status()).
		I("duration", int(duration)).
		Info(fmt.Sprintf("completed in %v", duration))
}

func logger() *log.Agregator {
	return log.S("lib", "svckit.httpi")
}
