package httpi

import (
	_ "expvar"
	"net/http"

	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/health"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
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
	n := r.Handler()
	// Run i ListenAndServe (koji se zove unutra)
	// nikada ne nastavi na slijedecoj liniji
	// nisam skuzio kako no ako je u svojoj goroutini onda je malo bolje
	//n.Run(listen)
	go func() {
		log.S("lib", "svckit/httpi").S("listen", listen).Info("starting")
		err := http.ListenAndServe(listen, n)
		if err != nil {
			log.Fatal(err)
		}
	}()
	signal.WaitForInterupt()
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
