package jsonu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Integer        int
	IntegerPointer *int
	String         string
	StringPointer  *string
	StructPointer  *TestStruct
}

const (
	testStr = `{
    "Integer": 42,
    "IntegerPointer": 42,
    "String": "test",
    "StringPointer": "test",
    "StructPointer": {
      "Integer": 0,
      "IntegerPointer": null,
      "String": "",
      "StringPointer": null,
      "StructPointer": null
    }
  }`
	testStr2 = `{
  "Integer": 42,
  "IntegerPointer": 42,
  "String": "test",
  "StringPointer": "test",
  "StructPointer": {
    "Integer": 0,
    "IntegerPointer": null,
    "String": "",
    "StringPointer": null,
    "StructPointer": null
  }
}`
)

func newTestStruct() *TestStruct {
	integer := 42
	str := "test"
	return &TestStruct{
		Integer:        integer,
		IntegerPointer: &integer,
		String:         str,
		StringPointer:  &str,
		StructPointer:  &TestStruct{},
	}
}

func TestMarshalPrettyBuf(t *testing.T) {
	buf, err := MarshalPrettyBuf([]byte(testStr))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, testStr2, string(buf))
}

func TestSprint(t *testing.T) {
	ts := newTestStruct()
	assert.Equal(t, testStr, Sprint(ts))
	assert.Equal(t, "", Sprint(make(chan bool)))
}
