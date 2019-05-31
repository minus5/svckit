package httpi

import (
	_ "expvar"
	"fmt"
	"net/http"

	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/health"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func NewRouter() *Router {
	return &Router{
		muxRouter: mux.NewRouter(),
		log:       false,
		noDebug:   false,
	}
}

type Router struct {
	muxRouter *mux.Router
	log       bool
	noDebug   bool
}

func (r *Router) Route(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route {
	return r.muxRouter.HandleFunc(path, f)
}

func (r *Router) Handle(path string, handler http.Handler) *mux.Route {
	return r.muxRouter.Handle(path, handler)
}

func (r *Router) HandlePath(path string, handler http.Handler) *mux.Route {
	return r.muxRouter.PathPrefix(path).Handler(handler)
}

func (r *Router) NewRoute() *mux.Route {
	return r.muxRouter.NewRoute()
}

func (r *Router) Subrouter(path string) *Router {
	return &Router{
		muxRouter: r.muxRouter.PathPrefix(path).Subrouter(),
		log:       false,
		noDebug:   false,
	}
}

//isto kao gore raspetlja url varijable
func (r *Router) RouteVars(path string,
	f func(http.ResponseWriter, *http.Request, map[string]string)) *mux.Route {
	return r.muxRouter.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		f(w, r, vars)
	})
}

// Start http server. Set port to listen and optionaly switches.
// Example:
//
//   httpi.Start(":8123", httpi.LogRequests())
func (r *Router) Start(listen string) {
	r.Go(listen)
	signal.WaitForInterupt()
}

func (r *Router) Go(listen string) {
	go func() {
		log.S("lib", "svckit/httpi").S("listen", listen).Info("starting")
		err := http.ListenAndServe(listen, r.Handler())
		if err != nil {
			log.Fatal(err)
		}
	}()
}

// Handler create http handler.
func (r *Router) Handler() *negroni.Negroni {
	if !r.noDebug {
		//ping
		r.muxRouter.HandleFunc("/ping", PingHttpResponse)
		//dodaj /health_check
		r.muxRouter.HandleFunc("/health_check", health.HttpHandler)
		//otvori expvar interface (na /debug/vars)
		r.muxRouter.Handle("/debug/vars", http.DefaultServeMux)
	}
	r.muxRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("501 url not implemented %s", r.URL.String()), http.StatusNotImplemented)
	})
	handlers := []negroni.Handler{negroni.NewRecovery(), NewStats()}
	if r.log {
		handlers = append(handlers, NewRequestLogger())
	}
	n := negroni.New(handlers...)
	if !r.noDebug {
		//dodaj pprof interface (na /debug/pprof)
		n.Use(Pprof())
	}
	n.UseHandler(r.muxRouter)
	return n
}

// NoDebug prevent adding debugging routers
func (r *Router) NoDebug() *Router {
	r.noDebug = true
	return r
}

// LogRequests if set will log all requests.
func (r *Router) Log() *Router {
	r.log = true
	return r
}

var defaultRouter = NewRouter()

func Route(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route {
	return defaultRouter.Route(path, f)
}

func Handle(path string, handler http.Handler) *mux.Route {
	return defaultRouter.Handle(path, handler)
}

func HandlePath(path string, handler http.Handler) *mux.Route {
	return defaultRouter.HandlePath(path, handler)
}

func NewRoute() *mux.Route {
	return defaultRouter.NewRoute()
}

func Subrouter(path string) *Router {
	return &Router{
		muxRouter: defaultRouter.muxRouter.PathPrefix(path).Subrouter(),
		log:       false,
		noDebug:   false,
	}
}

//isto kao gore raspetlja url varijable
func RouteVars(path string, f func(http.ResponseWriter, *http.Request, map[string]string)) *mux.Route {
	return defaultRouter.RouteVars(path, f)
}

// Start http server. Set port to listen and optionaly switches.
// Example:
//
//   httpi.Start(":8123", httpi.LogRequests())
func Start(listen string) {
	defaultRouter.Start(listen)
}

// Handler create http handler.
func Handler() *negroni.Negroni {
	return defaultRouter.Handler()
}

// PingHttpResponse helper for Ping http method.
// Returns status ok, an application name in header.
func PingHttpResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Application", env.AppName())
	w.WriteHeader(http.StatusOK)
}
