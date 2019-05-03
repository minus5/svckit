package main

import (
	"io/ioutil"
	"net/http"

	"github.com/minus5/svckit/amp"
	"github.com/minus5/svckit/amp/session"
	"github.com/minus5/svckit/log"
)

type pooling struct {
	sessions *session.Sessions
}

func (s *pooling) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if path == "health_check" || path == "ping" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
