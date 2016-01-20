package logger

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestSubexp(t *testing.T) {

	msgs := []string{
		"2016/01/18 23:14:18.295980 file.go:100: [NOTICE] msg1",
		"2016/01/18 23:14:18.295980 file.go:100: [ERROR] msg1",
		"2016/01/18 23:14:18.295980 [INFO] msg1",
		"2016/01/18 23:14:18.295980 msg1",
	}

	//vizualno testiranje
	// for _, msg := range msgs {
	// 	fmt.Printf("\n%s\n", msg)
	// 	m := subexps([]byte(msg))
	// 	for k, v := range m {
	// 		if k != "" {
	// 			fmt.Printf("\t%s => %s\n", k, v)
	// 		}
	// 	}
	// }

	m := subexps([]byte(msgs[0]))
	assert.Equal(t, m["date"], "2016/01/18")
	assert.Equal(t, m["time"], "23:14:18.295980")
	assert.Equal(t, m["file"], "file.go:100")
	assert.Equal(t, m["level"], "NOTICE")
	assert.Equal(t, m["msg"], " msg1")

	m = subexps([]byte(msgs[1]))
	assert.Equal(t, m["date"], "2016/01/18")
	assert.Equal(t, m["time"], "23:14:18.295980")
	assert.Equal(t, m["file"], "file.go:100")
	assert.Equal(t, m["level"], "ERROR")
	assert.Equal(t, m["msg"], " msg1")

	m = subexps([]byte(msgs[2]))
	assert.Equal(t, m["date"], "2016/01/18")
	assert.Equal(t, m["time"], "23:14:18.295980")
	assert.Equal(t, m["file"], "")
	assert.Equal(t, m["level"], "INFO")
	assert.Equal(t, m["msg"], " msg1")

	m = subexps([]byte(msgs[3]))
	assert.Equal(t, m["date"], "2016/01/18")
	assert.Equal(t, m["time"], "23:14:18.295980")
	assert.Equal(t, m["file"], "")
	assert.Equal(t, m["level"], "")
	assert.Equal(t, m["msg"], "msg1")

}

func TestToMessage(t *testing.T) {
	m := toMessage([]byte("2016/01/18 23:14:18.295980 file.go:100: [NOTICE] msg1"), 5)
	assert.Equal(t, m.Level, "notice")
	assert.Equal(t, m.File, "file.go:100")
	assert.Equal(t, m.Message, "msg1")
	assert.Equal(t, m.No, 5)
}
