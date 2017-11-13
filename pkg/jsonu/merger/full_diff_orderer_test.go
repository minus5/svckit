package merger

import (
	"fmt"
	"testing"
	"time"

	"github.com/minus5/svckit/log"

	"github.com/stretchr/testify/assert"
)

func full(no int) *msg {
	return &msg{typ: "d_123/full", isFull: true, no: int64(no)}
}

func diff(no int) *msg {
	return &msg{typ: "d_123/diff", isDiff: true, no: int64(no)}
}

func TestCheck(t *testing.T) {
	type testData struct {
		full   bool
		no     int
		result int
	}
	var data []testData

	o := newFullDiffOrderer(func() {})

	assertData := func() {
		for _, d := range data {
			var m *msg
			if d.full {
				m = full(d.no)
			} else {
				m = diff(d.no)
			}
			assert.Equal(t, o.check(m), d.result,
				fmt.Sprintf("no %d, m.no: %d, m.isDiff: %v", o.no, m.no, m.isDiff))
		}
	}

	//na poceku moze bilo koji full i niti jedan diff
	data = []testData{
		{true, 5, checkReplace},
		{true, 15, checkReplace},
		{false, 5, checkRequestFull},
		{false, 15, checkRequestFull},
	}
	assertData()

	o.no = 5
	data = []testData{
		{true, 4, checkSkip},
		{true, 5, checkReplace},
		{true, 6, checkLater},
		{true, 7, checkLater},
		{false, 4, checkSkip},
		{false, 5, checkCurrent},
		{false, 6, checkMerge},
		{false, 7, checkLater},
	}
	assertData()

	//reset ide kada full naraste za vise od 2 ili diff za vise od 99
	o.no = 5
	data = []testData{
		{false, 7, checkLater},
		{false, 10, checkLater},
		{false, 21, checkLater},
		{false, 22, checkReset}, //diff je narastao za vise od 16
		{true, 7, checkLater},
		{true, 8, checkReset}, //full je narastao za vise od 2
	}
	assertData()

	//bug fix, zakasnjeli diff na startu treba bit skip-nut a ne later
	o.no = 96
	data = []testData{
		{false, 88, checkSkip},
	}
	assertData()
}

func TestFullDiffOrderer(t *testing.T) {
	log.Discard()
	fullRequests := 0
	o := newFullDiffOrderer(func() {
		fullRequests++
	})
	out := func() *msg {
		select {
		case m := <-o.out:
			return m
		case <-time.After(time.Millisecond * 100): // vrati nil ako nema vise poruka za van
			return nil
		}
	}

	// dok nema full diffovi rade request full
	o.in <- diff(9)
	o.in <- diff(10)
	o.in <- diff(11)
	o.in <- diff(12)
	m := out()
	assert.Nil(t, m)
	assert.Nil(t, o.current)
	assert.Equal(t, 4, len(o.queue))
	assert.Equal(t, 1, fullRequests)
	assert.Equal(t, "d9 d10 d11 d12", o.inQueue())

	// processMsg(f10) i processQueeu izbaci d10
	o.in <- full(10)
	m = out()
	assert.NotNil(t, o.current)
	assert.True(t, m.isFull)
	assert.EqualValues(t, 10, m.no)
	assert.EqualValues(t, 10, o.no)
	assert.Equal(t, 2, len(o.queue))
	assert.Equal(t, 1, fullRequests)
	assert.Equal(t, "d11 d12", o.inQueue())

	// i oslobodi diff-ove
	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isDiff)
	assert.EqualValues(t, 10, m.no)
	assert.Equal(t, "d11 d12", o.inQueue())

	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isDiff)
	assert.EqualValues(t, 11, m.no)
	assert.Equal(t, "d12", o.inQueue())

	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isFull)
	assert.EqualValues(t, 11, m.no)

	m = out()
	assert.Equal(t, "", o.inQueue())
	assert.NotNil(t, m)
	assert.True(t, m.isDiff)
	assert.EqualValues(t, 12, m.no)

	m = out()
	assert.Equal(t, "", o.inQueue())
	assert.NotNil(t, m)
	assert.True(t, m.isFull)
	assert.EqualValues(t, 12, m.no)

	m = out()
	assert.Nil(t, m)
	assert.EqualValues(t, 12, o.no)
	assert.Equal(t, 0, len(o.queue))
	assert.Equal(t, 1, fullRequests)
	assert.Equal(t, "", o.inQueue())

	// diff koji preskacemo
	o.in <- diff(11)
	m = out()
	assert.Nil(t, m)

	// i full koji preskacemo
	o.in <- full(11)
	m = out()
	assert.Nil(t, m)

	// i full koji samo zamjeni trenutni
	o.in <- full(12)
	m = out()
	assert.Nil(t, m)

	// full-ovi cekaju diff-ove
	o.in <- full(13)
	m = out()
	assert.Nil(t, m)
	assert.Equal(t, 1, len(o.queue))
	assert.Equal(t, "f13", o.inQueue())

	o.in <- full(14)
	m = out()
	assert.Nil(t, m)
	assert.Equal(t, 2, len(o.queue))
	assert.Equal(t, "f13 f14", o.inQueue())

	o.in <- diff(13)
	m = out()
	assert.NotNil(t, m)
	m = out()
	assert.NotNil(t, m)
	m = out()
	assert.Nil(t, m)
	o.in <- diff(14)
	m = out()
	assert.NotNil(t, m)
	m = out()
	assert.NotNil(t, m)
	m = out()
	assert.Nil(t, m)

	// fali diff, cekamo ga
	o.in <- diff(16)
	m = out()
	assert.Nil(t, m)
	o.in <- diff(17)
	m = out()
	assert.Nil(t, m)
	assert.Equal(t, 2, len(o.queue))
	assert.Equal(t, "d16 d17", o.inQueue())

	m = out()
	assert.Nil(t, m)

	// kada dodje izadju van i diff-ovi i full-ovi
	o.in <- diff(15)
	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isDiff)
	assert.EqualValues(t, 15, m.no)
	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isFull)
	assert.EqualValues(t, 15, m.no)
	assert.Equal(t, "d16 d17", o.inQueue())

	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isDiff)
	assert.EqualValues(t, 16, m.no)
	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isFull)
	assert.EqualValues(t, 16, m.no)

	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isDiff)
	assert.EqualValues(t, 17, m.no)
	m = out()
	assert.NotNil(t, m)
	assert.True(t, m.isFull)
	assert.EqualValues(t, 17, m.no)

	m = out()
	assert.Nil(t, m)
}

// imali smo bug kada se ova varijanta zavrtila u beskonacnoj petlji
func TestBugFix(t *testing.T) {
	fullRequests := 0
	o := newFullDiffOrderer(func() {
		fullRequests++
	})
	for i := 1; i <= maxQueueSize; i++ {
		m := diff(i)
		o.queue = append(o.queue, m)
	}
	assert.Equal(t, 0, fullRequests)
	assert.Equal(t, maxQueueSize, len(o.queue))
	//t.Logf("queueSize: %d", len(o.queue))
	o.processQueue()
	assert.Equal(t, 1, fullRequests)
	assert.Equal(t, 1, len(o.queue))
}
