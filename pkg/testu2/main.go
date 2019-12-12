package testu

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//ako zelim snimiti fixture dodam true kraj
func AssertFixture(t *testing.T, expectedFile string, a interface{}, params ...bool) {

	var actualBuf []byte
	if buf, ok := a.([]byte); ok {
		actualBuf = buf
	} else {
		buf, _ := json.MarshalIndent(a, "  ", "  ")
		actualBuf = buf
	}
	actual := string(actualBuf)

	if (len(params) > 0 && params[0]) || !Exists(expectedFile) {
		t.Logf("saving fixture %s", expectedFile)
		SaveFixture(expectedFile, actualBuf)
		return
	}

	expected := string(ReadFixture(t, expectedFile))
	same := actual == expected
	if !same {
		actualFile := expectedFile + "_actual"
		SaveFixture(actualFile, actualBuf)
		cmd := exec.Command("icdiff", "--highlight", "--cols=160", expectedFile, actualFile)
		//cmd := exec.Command("git", "diff", "--no-index", "--color=always", expectedFile, actualFile)
		out, err := cmd.Output()
		if err != nil {
			t.Logf("diff status %s", err)
			t.Log("diff dependa na icdiff bin moguce da on nije instaliran http://www.jefftk.com/icdiff")
		}
		//t.Logf("file differences between %s and %s:\n%s", expectedFile, actualFile, out)
		fmt.Printf("%s", out)
	}
	assert.True(t, same, expectedFile)
}

func AssertSameStrings(t *testing.T, expected string, actual string) {
	same := actual == expected
	if same {
		return
	}
	a, err := ioutil.TempFile(os.TempDir(), "actual")
	if err != nil {
		log.Fatal(err)
	}
	e, err := ioutil.TempFile(os.TempDir(), "expected")
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(a.Name(), []byte(actual), 0644)
	ioutil.WriteFile(e.Name(), []byte(expected), 0644)
	fmt.Printf("%s <-> %s\n", a.Name(), e.Name())
	defer os.Remove(a.Name())
	defer os.Remove(e.Name())

	cmd := exec.Command("icdiff", "--highlight", a.Name(), e.Name())
	out, err := cmd.Output()
	if err != nil {
		t.Logf("diff status %s", err)
		t.Log("diff dependa na icdiff bin moguce da on nije instaliran http://www.jefftk.com/icdiff")
	}
	fmt.Printf("%s", out)

	assert.True(t, same)
}

func JsonDiff(t *testing.T, expected, actual []byte) int {
	a, err := ioutil.TempFile(os.TempDir(), "actual")
	if err != nil {
		log.Fatal(err)
	}
	e, err := ioutil.TempFile(os.TempDir(), "expected")
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(a.Name(), actual, 0644)
	ioutil.WriteFile(e.Name(), expected, 0644)
	defer os.Remove(a.Name())
	defer os.Remove(e.Name())

	//fmt.Printf("json-diff %s %s\n", a.Name(), e.Name())
	cmd := exec.Command("json-diff", a.Name(), e.Name())
	out, err := cmd.Output()
	if err != nil {
		t.Logf("diff status %s", err)
	}
	if len(out) == 0 || strings.TrimSpace(string(out)) == "{\n }" {
		return 0
	}
	fmt.Printf("%s\n", out)
	// if len(out) > 0 {
	// 	ioutil.WriteFile("./actual", actual, 0644)
	// 	ioutil.WriteFile("./expected", expected, 0644)
	// }
	return len(out)
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// ReadFixture ucitava file
func ReadFixture(t *testing.T, name string) []byte {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s", name))
	assert.NoError(t, err)
	return buf
}

// ReadFixtureJSON ucitava JSON file
func ReadFixtureJSON(t *testing.T, val interface{}, name string) {
	buf := ReadFixture(t, name)
	if !assert.NotEmpty(t, buf) {
		t.FailNow()
	}
	if !assert.NoError(t, json.Unmarshal(buf, val)) {
		t.FailNow()
	}
}

func SaveFixture(filename string, buf []byte) error {
	dir, _ := path.Split(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf, 0644)
}

func SaveJSON(name string, o interface{}) {
	buf, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Printf("%s", err)
		return
	}
	ioutil.WriteFile(name, buf, 0644)
}

func SaveJSONUgly(name string, o interface{}) {
	buf, err := json.Marshal(o)
	if err != nil {
		log.Printf("%s", err)
		return
	}
	ioutil.WriteFile(name, buf, 0644)
}

func RunShellScript(t *testing.T, name string) {
	cmd := exec.Command("/bin/sh", name)
	err := cmd.Start()
	assert.Nil(t, err)
	err = cmd.Wait()
	assert.Nil(t, err)
}

func ShowDiff(file1, file2 string) {
	cmd := exec.Command("icdiff", "--highlight", "--line-numbers", "-U 1", "--cols=128", file1, file2)
	out, err := cmd.Output()
	if err == nil {
		fmt.Printf("%s", out)
	} else {
		log.Printf("diff status %s", err)
	}
}

// RunShellBatch izvrsi niz shell komandi
func RunShellBatch(t *testing.T, batch string) {
	r := 0
again:
	cmd := exec.Command("bash", "-c", batch)
	err := cmd.Start()
	assert.Nil(t, err)
	err = cmd.Wait()

	if err != nil {
		if r < 3 {
			r++
			goto again
		}
		assert.Nil(t, err)
		panic(err)
	}
}

// StartMongo pokrece cisti mongo za testiranje na portu 27018
func StartMongo(t *testing.T) {
	batch := `#!/bin/bash
mongo --port 27018  --eval "db.getSiblingDB('admin').shutdownServer()" --quiet
rm -rf /tmp/test_mongo
mkdir -p /tmp/test_mongo
mongod --port 27018 --nojournal --dbpath /tmp/test_mongo --logpath /tmp/test_mongo.log --fork --quiet -–smallfiles -–noprealloc 
`
	RunShellBatch(t, batch)
}

// StopMongo zaustavlja testni mongo
func StopMongo(t *testing.T) {
	batch := `#!/bin/bash
mongo --port 27018  --eval "db.getSiblingDB('admin').shutdownServer()" --quiet
rm -rf /tmp/test_mongo
rm -f /tmp/test_mongo.log
`

	RunShellBatch(t, batch)
}

// PP prety print object
func PP(o interface{}) {
	buf, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", buf)
}

// PP prety print object
func PpBuf(buf []byte) string {
	var m map[string]interface{}
	json.Unmarshal(buf, &m)
	return string(Pp(m))
}

func PPBuf(buf []byte) {
	fmt.Printf("pp:\n%s\n", PpBuf(buf))
}

// PP prety print object
func Pp(o interface{}) []byte {
	buf, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(err)
	}
	return buf
}

func Json(o interface{}) []byte {
	buf, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return buf
}

func FnFullName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func FnName(i interface{}) string {
	n := FnFullName(i)
	p := strings.Split(n, ".")
	if len(p) > 1 {
		return p[1]
	}
	return n
}

func Ppt(t *testing.T, o interface{}) {
	buf, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(err)
	}
	t.Logf("%s", buf)
}
