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
