package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/minus5/svckit/log"
)

type logger struct{}

func (logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if path == "health_check" || path == "ping" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	if path == "error" {
		l.Error(nil)
	} else {
		l.Info("")
	}

	w.WriteHeader(http.StatusOK)
}
