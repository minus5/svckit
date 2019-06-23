package cgen

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

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
	e := &Event{}
	stcs := AnalyzeStruct(e)
	spew.Dump(stcs)
	buf, _ := json.MarshalIndent(stcs, "  ", "  ")
	fmt.Printf("%s\n", buf)
}

func TestDiff(t *testing.T) {
	e := &Event{}
	d := diff(AnalyzeStruct(e))
	fmt.Printf("%s\n", d)
}

func TestMerge(t *testing.T) {
	e := &Event{}
	d := merge(AnalyzeStruct(e))
	fmt.Printf("%s\n", d)
}

func TestDiffMethods(t *testing.T) {
	e := &Event{}
	d := diffMethods(AnalyzeStruct(e))
	fmt.Printf("%s\n", d)
}
