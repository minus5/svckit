package main

import (
	"fmt"
	"github.com/minus5/svckit/httpi"
	"net/http"
)

type Mdlw struct{}

func (m *Mdlw) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Println("pozvan je middleware handler")
	next(rw, r)
}

// demo za koristenje custom middlewarea na svim httpi routama
func main() {
	httpi.Route("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("pozvan je http handler")
		w.Write([]byte("neki response za browser"))
	})
	mws := []httpi.Middleware{&Mdlw{}}
	httpi.Start(":4321", mws...)
}
