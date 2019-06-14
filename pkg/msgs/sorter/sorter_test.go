package sorter

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoredak(t *testing.T) {
	s := New(time.Second)

	var wg sync.WaitGroup
	var out []int
	wg.Add(1)
	go func() {
		for m := range s.Output {
			out = append(out, m.No)
		}
		wg.Done()
	}()

	s.Push(&Msg{No: 1})
	s.Push(&Msg{No: 2})
	s.Push(&Msg{No: 4})
	s.Push(&Msg{No: 1})

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 2, s.current)
	assert.Equal(t, []int{1, 2, 1}, out)
	t.Logf("%#v\n", out)

	s.Push(&Msg{No: 5})
	s.Push(&Msg{No: 2})
	s.Push(&Msg{No: 3})

	s.Close()
	wg.Wait()
	assert.Equal(t, out, []int{1, 2, 1, 2, 3, 4, 5})
	t.Logf("%#v\n", out)
}

func TestFixDrugiPurgeSeNijeDogodio(t *testing.T) {
	s := New(100 * time.Millisecond)

	var wg sync.WaitGroup
	var out []int
	wg.Add(1)
	go func() {
		for m := range s.Output {
			out = append(out, m.No)
		}
		wg.Done()
	}()

	s.Push(&Msg{No: 1})
	s.Push(&Msg{No: 2})
	s.Push(&Msg{No: 5})
	s.Push(&Msg{No: 6})
	s.Push(&Msg{No: 9})

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 2, s.current)
	assert.Equal(t, []int{1, 2}, out)

	// prvi purge
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 9, s.current)
	assert.Equal(t, []int{1, 2, 5, 6, 9}, out)
	//fmt.Printf("%#v\n", out)

	// drugi purge
	s.Push(&Msg{No: 11})
	s.Push(&Msg{No: 12})
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 12, s.current)
	assert.Equal(t, []int{1, 2, 5, 6, 9, 11, 12}, out)
	//fmt.Printf("%#v\n", out)

	s.Close()
	wg.Wait()
	assert.Equal(t, []int{1, 2, 5, 6, 9, 11, 12}, out)
	//t.Logf("%#v\n", out)
	//fmt.Printf("%#v\n", out)
}

func Test0ResetiraCurrent(t *testing.T) {
	s := New(100 * time.Millisecond)

	var wg sync.WaitGroup
	var out []int
	wg.Add(1)
	go func() {
		for m := range s.Output {
			out = append(out, m.No)
		}
		wg.Done()
	}()

	s.Push(&Msg{No: 1})
	s.Push(&Msg{No: 2})
	s.Push(&Msg{No: 4})
	s.Push(&Msg{No: 0})
	s.Push(&Msg{No: 6})
	s.Push(&Msg{No: 9})

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 6, s.current)
	assert.Equal(t, []int{1, 2, 0, 6}, out)

	// prvi purge
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 9, s.current)
	assert.Equal(t, []int{1, 2, 0, 6, 9}, out)
	//fmt.Printf("%#v\n", out)

	s.Close()
	wg.Wait()
	assert.Equal(t, []int{1, 2, 0, 6, 9}, out)
}

func TestClose(t *testing.T) {
	s := New(time.Second)

	var wg sync.WaitGroup
	var out []int
	wg.Add(1)
	go func() {
		for m := range s.Output {
			out = append(out, m.No)
		}
		wg.Done()
	}()

	s.Push(&Msg{No: 1})
	s.Push(&Msg{No: 4})
	s.Push(&Msg{No: 5})
	s.Push(&Msg{No: 3})

	s.Close()
	wg.Wait()
	assert.Equal(t, out, []int{1, 3, 4, 5})
	t.Logf("%#v\n", out)
}

func TestPurge(t *testing.T) {
	s := New(100 * time.Millisecond)

	var wg sync.WaitGroup
	var out []int
	wg.Add(1)
	go func() {
		for m := range s.Output {
			out = append(out, m.No)
		}
		wg.Done()
	}()

	s.Push(&Msg{No: 1})
	s.Push(&Msg{No: 4})
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, s.current)

	s.Push(&Msg{No: 5})
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, s.current)

	s.Push(&Msg{No: 3})
	assert.Equal(t, 1, s.current)
	assert.False(t, s.empty())

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 5, s.current)
	assert.Equal(t, []int{1, 3, 4, 5}, out)
	assert.True(t, s.empty())

	s.Push(&Msg{No: 5})
	s.Push(&Msg{No: 8})
	s.Push(&Msg{No: 7})
	s.Push(&Msg{No: 6})
	time.Sleep(90 * time.Millisecond)
	assert.Equal(t, []int{1, 3, 4, 5, 5, 6, 7, 8}, out)

	s.Close()
	wg.Wait()
	t.Logf("%#v\n", out)
}
