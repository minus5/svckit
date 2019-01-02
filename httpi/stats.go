package httpi

import (
	"fmt"
	"net/http"
	"github.com/mnu5/svckit/metric"
	"strings"
	"time"
)

type Stats struct {
}

func NewStats() *Stats {
	return &Stats{}
}

func (s *Stats) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	method := "/"
	if parts := strings.Split(r.URL.Path, "/"); len(parts) > 1 {
		method = parts[1]
	}

	start := time.Now()
	next(rw, r)
	metric.Time(fmt.Sprintf("http.%s", method), int(time.Since(start).Nanoseconds()))
}
