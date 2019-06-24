package cgen_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/minus5/svckit/cgen"
	"github.com/minus5/svckit/pkg/testu"
)

// switch to true to update test fixtures
// by running test with flag save-fixtures
//   go test --save-fixtures
var saveFixtures = false

func init() {
	flag.BoolVar(&saveFixtures, "save-fixtures", false, "snimi fixture umjesto testiranja spram njih")
	flag.Parse()
}

type Event struct {
	Home     string         `json:"h"`
	Away     string         `json:"a"`
	Schedule time.Time      `json:"s"`
	Result   Result         `json:"r"`
	Markets  map[int]Market `json:"m,omitempty"`
}
type Market struct {
	Name     string
	Outcomes map[int]Outcome
}
type Outcome struct {
	Name string
	Odds float64
}
type Result struct {
	Home int
	Away int
}

func TestAnalyzeStruct(t *testing.T) {
	t.Skip("experiments")
	e := &Event{}
	stcs := cgen.Analyze(e)
	spew.Dump(stcs)
	buf, _ := json.MarshalIndent(stcs, "  ", "  ")
	fmt.Printf("%s\n", buf)
}

func TestDiff(t *testing.T) {
	e := &Event{}
	d := cgen.Analyze(e).Diff()
	testu.AssertFixture(t, "./fixtures/event_diff.gen", d.Bytes(), saveFixtures)
}
