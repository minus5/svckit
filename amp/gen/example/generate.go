// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates diff_gen.go
// It can be invoked by running:
// go generate
package main

import (
	"log"
	"reflect"

	"pkg/prematch/gen"
	"pkg/prematch/gen/example"
)

func main() {
	d := &example.Person{}
	v := reflect.ValueOf(d).Elem()
	err := gen.Diff(v, "./data_gen.go")
	if err != nil {
		log.Fatal(err)
	}
}
