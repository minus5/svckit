package cgen

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

type Event struct {
	Home     string
	Away     string
	Schedule time.Time
	Result   Result
	Markets  map[int]Market
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
	e := &Event{}
	stcs := analyzeStruct(e)
	spew.Dump(stcs)
	//spew.Printf("myVar1: %#v\n", stcs)
	buf, _ := json.MarshalIndent(stcs, "  ", "  ")
	fmt.Printf("%s\n", buf)
}
