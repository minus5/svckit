package httpi

import (
	"net/http"
	"net/http/pprof"
	runtime_pprof "runtime/pprof"
	"strings"
)

type handler struct {
}

var profileHandlers map[string]http.Handler

func init() {
	profileHandlers = make(map[string]http.Handler)
	for _, profile := range runtime_pprof.Profiles() {
		profileHandlers[profile.Name()] = pprof.Handler(profile.Name())
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	// pathove kopiram iz init libexec/src/net/http/pprof/pprof.go
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/cmdline") {
		pprof.Cmdline(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/profile") {
		pprof.Profile(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/symbol") {
		pprof.Symbol(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/trace") {
		pprof.Trace(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/debug/pprof") {
		pprof.Index(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/debug/block") ||
		strings.HasPrefix(r.URL.Path, "/debug/goroutine") ||
		strings.HasPrefix(r.URL.Path, "/debug/heap") ||
		strings.HasPrefix(r.URL.Path, "/debug/threadcreate") {
		parts := strings.Split(r.URL.Path, "/")
		handler := profileHandlers[parts[2]]
		if handler != nil {
			handler.ServeHTTP(w, r)
			return
		}
	}

	next(w, r)
}

// Pprof returns a handler which will serve pprof data for the path /debug/pprof
func Pprof() *handler {
	return &handler{}
}
