// +build ignore

package main

import (
	"github.com/minus5/svckit/cgen"
	"github.com/minus5/svckit/cgen/example"
)

func main() {
	var e example.Event
	cgen.Analyze(e).Diff().Save("./event_diff_gen.go")
}
