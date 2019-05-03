package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/minus5/svckit/amp"
	"github.com/minus5/svckit/amp/session"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
)

func appHTTPServer(interupt context.Context, portLabel, appRoot string, sessions *session.Sessions) {
	srv := &http.Server{
		Addr: env.Address(portLabel),
		Handler: &appServer{
			sessions:   sessions,
			fileServer: http.FileServer(http.Dir(appRoot)),
		},
	}
	go func() {
		<-interupt.Done()
		srv.Shutdown(context.Background())
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error(err)
	}
}

type appServer struct {
	sessions   *session.Sessions
	fileServer http.Handler
}

func (s *appServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	switch path {
	case "pooling":
		s.pooling(w, r)
	case "health_check":
		w.WriteHeader(http.StatusOK)
	case "ping":
		w.WriteHeader(http.StatusOK)
	case "log/info":
		s.info(w, r)
	case "log/error":
		s.error(w, r)
	default:
		s.fileServer.ServeHTTP(w, r)
	}
}

func (s *appServer) pooling(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	m := amp.Parse(buf)
	if m == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rsp := s.sessions.Pool(m)
	for i, r := range rsp {
		w.Write(r.Marshal())
		if i < len(rsp)-1 {
			w.Write([]byte{10, 10})
		}
	}
}

func (s *appServer) info(w http.ResponseWriter, r *http.Request) {
	s.log(w, r, false)
}

func (s *appServer) error(w http.ResponseWriter, r *http.Request) {
	s.log(w, r, true)
}

func (s *appServer) log(w http.ResponseWriter, r *http.Request, isError bool) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		//log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	hm := make(map[string]interface{})
	for k, v := range r.Header {
		if len(v) == 1 {
			hm[k] = v[0]
		} else {
			hm[k] = v
		}
	}
	h, err := json.Marshal(hm)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	l := log.Jc("body", body).
		Jc("header", h)
	if isError {
		l.Error(nil)
	} else {
		l.Info("")
	}

	w.WriteHeader(http.StatusOK)
}
