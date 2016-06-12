package testu

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

//ako zelim snimiti fixture dodam true kraj
func AssertFixture(t *testing.T, expectedFile string, a []byte, params ...bool) {
	if len(params) > 0 && params[0] {
		t.Logf("saving fixture %s", expectedFile)
		SaveFixture(expectedFile, a)
		return
	}
	actual := string(a)
	expected := string(ReadFixture(t, expectedFile))
	same := actual == expected
	if !same {
		actualFile := expectedFile + "_actual"
		SaveFixture(actualFile, a)
		cmd := exec.Command("icdiff", "--highlight", expectedFile, actualFile)
		out, err := cmd.Output()
		if err != nil {
			t.Logf("diff status %s", err)
			t.Log("diff dependa na icdiff bin moguce da on nije instaliran http://www.jefftk.com/icdiff")
		}
		t.Logf("file differences between %s and %s:\n%s", expectedFile, actualFile, out)
	}
	assert.True(t, same, expectedFile)
}

func ReadFixture(t *testing.T, name string) []byte {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s", name))
	assert.Nil(t, err)
	return buf
}

func SaveFixture(name string, buf []byte) {
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
