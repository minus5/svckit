package broker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func concatenate(ch chan []byte, out *[]byte) {
	for msg := range ch {
		*out = append(*out, msg...)
	}
}

func TestFullDiff(t *testing.T) {
	Full("test", []byte("12345"))

	var buf1, buf2 []byte
	b1 := GetFullDiffBroker("test")
	ch1 := b1.Subscribe()
	b2 := GetFullDiffBroker("test")
	ch2 := b2.Subscribe()
	go concatenate(ch1, &buf1)
	go concatenate(ch2, &buf2)

	time.Sleep(10 * time.Millisecond)

	Diff("test", []byte("6"))
	Diff("test", []byte("7"))
	Diff("test", []byte("8"))

	time.Sleep(10 * time.Millisecond)

	b1.Unsubscribe(ch1)
	b2.Unsubscribe(ch2)

	assert.Equal(t, "12345678", string(buf1))
	assert.Equal(t, "12345678", string(buf2))
}

func TestBuffered(t *testing.T) {
	createBufferedBroker("teststream", 10)
	Stream("teststream", []byte("1"))
	Stream("teststream", []byte("2"))
	Stream("teststream", []byte("3"))
	Stream("teststream", []byte("4"))
	Stream("teststream", []byte("5"))

	var buf1, buf2 []byte
	b1 := GetBufferedBroker("teststream")
	ch1 := b1.Subscribe()
	b2 := GetBufferedBroker("teststream")
	ch2 := b2.Subscribe()
	go concatenate(ch1, &buf1)
	go concatenate(ch2, &buf2)

	time.Sleep(10 * time.Millisecond)

	Stream("teststream", []byte("6"))
	Stream("teststream", []byte("7"))
	Stream("teststream", []byte("8"))

	time.Sleep(10 * time.Millisecond)

	b1.Unsubscribe(ch1)
	b2.Unsubscribe(ch2)

	assert.Equal(t, "12345678", string(buf1))
	assert.Equal(t, "12345678", string(buf2))

	Stream("teststream", []byte("9"))
	Stream("teststream", []byte("10"))
	Stream("teststream", []byte("11"))
	Stream("teststream", []byte("12"))
	Stream("teststream", []byte("13"))

	time.Sleep(10 * time.Millisecond)

	var buf3 []byte
	b3 := GetBufferedBroker("teststream")
	ch3 := b3.Subscribe()
	go concatenate(ch3, &buf3)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "45678910111213", string(buf3))
}
