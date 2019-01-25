package jsondiff

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

var (
	showJSONDiff   bool
	updateJSONDiff bool
)

func init() {
	flag.BoolVar(&showJSONDiff, "jd-show", false, "show json diff")
	flag.BoolVar(&updateJSONDiff, "jd-update", false, "update expected file with actual")
}

// Check checks whether provided object produces same JSON as previously saved file.
// It is intended for use in test so first param is testing.T.
// If differences exists there are options to show that diff (jd-show flag).
// And the option to overwrite expected file with actual object JSON (jd-update flag).
func Check(t *testing.T, filepath string, o interface{}) {
	actual, err := json.MarshalIndent(o, "  ", "  ")
	if err != nil {
		t.Fatal(err)
	}

	update := func() {
		err = ioutil.WriteFile(filepath, actual, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Logf("creating %s", filepath)
		update()
		return
	}
	expected, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	differ := diff.New()
	d, err := differ.Compare(expected, actual)
	if err != nil {
		log.Fatal(err)
	}
	if !d.Modified() {
		// they are same, no difference
		return
	}
	if updateJSONDiff {
		update()
		return
	}

	t.Errorf("failed to get same json as in %s", filepath)
	if !showJSONDiff {
		t.Errorf("use jd-show flag to show diff")
		t.Errorf("or jd-update to overwrite expected file")
		return
	}
	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
		Coloring:       true,
	}
	var m map[string]interface{}
	if err := json.Unmarshal(expected, &m); err != nil {
		log.Fatal(err)
	}
	diff, err := formatter.NewAsciiFormatter(m, config).Format(d)
	if err != nil {
		log.Fatal(err)
	}
	t.Errorf("diff: \n%s\n", diff)
}
