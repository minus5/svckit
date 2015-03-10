package testu

import (
	"fmt"
	"io/ioutil"
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
