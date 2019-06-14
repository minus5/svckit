package broker

import (
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func concatenate(ch chan *Message, out *[]byte) <-chan bool {
	c := make(chan bool)
	go func() {
		for msg := range ch {
			*out = append(*out, msg.Data...)
		}
		c <- true
	}()
	return c
}

func TestFullDiff(t *testing.T) {
	Full("test", "testevent", []byte("12345"))

	var buf1, buf2 []byte

	b1 := GetFullDiffBroker("test")
	ch1 := b1.Subscribe()
	done1 := concatenate(ch1, &buf1)
	b2 := GetFullDiffBroker("test")
	ch2 := b2.Subscribe()
	done2 := concatenate(ch2, &buf2)

	time.Sleep(10 * time.Millisecond)

	Diff("test", "testevent", []byte("6"))
	Diff("test", "testevent", []byte("7"))
	Diff("test", "testevent", []byte("8"))

	time.Sleep(10 * time.Millisecond)

	b1.Unsubscribe(ch1)
	b2.Unsubscribe(ch2)

	<-done1
	<-done2

	assert.Equal(t, "12345678", string(buf1))
	assert.Equal(t, "12345678", string(buf2))
}

func TestBuffered(t *testing.T) {
	createBufferedBroker("teststream", 10)
	Stream("teststream", "testevent", []byte("1"))
	Stream("teststream", "testevent", []byte("2"))
	Stream("teststream", "testevent", []byte("3"))
	Stream("teststream", "testevent", []byte("4"))
	Stream("teststream", "testevent", []byte("5"))

	var buf1, buf2 []byte
	b1 := GetBufferedBroker("teststream")
	ch1 := b1.Subscribe()
	done1 := concatenate(ch1, &buf1)
	b2 := GetBufferedBroker("teststream")
	ch2 := b2.Subscribe()

	done2 := concatenate(ch2, &buf2)

	time.Sleep(10 * time.Millisecond)

	Stream("teststream", "testevent", []byte("6"))
	Stream("teststream", "testevent", []byte("7"))
	Stream("teststream", "testevent", []byte("8"))

	time.Sleep(10 * time.Millisecond)

	b1.Unsubscribe(ch1)
	b2.Unsubscribe(ch2)

	<-done1
	<-done2

	assert.Equal(t, "12345678", string(buf1))
	assert.Equal(t, "12345678", string(buf2))

	Stream("teststream", "testevent", []byte("9"))
	Stream("teststream", "testevent", []byte("10"))
	Stream("teststream", "testevent", []byte("11"))
	Stream("teststream", "testevent", []byte("12"))
	Stream("teststream", "testevent", []byte("13"))

	time.Sleep(10 * time.Millisecond)

	var buf3 []byte
	b3 := GetBufferedBroker("teststream")
	ch3 := b3.Subscribe()
	done3 := concatenate(ch3, &buf3)
	time.Sleep(10 * time.Millisecond)
	b3.Unsubscribe(ch3)
	<-done3
	assert.Equal(t, "45678910111213", string(buf3))
}

func TestTouch(t *testing.T) {
	log.Println("touch test")
	var buf1 []byte
	b1 := GetFullDiffBroker("touch")
	ch1 := b1.Subscribe()
	done1 := concatenate(ch1, &buf1)

	Diff("touch", "test", []byte("diff1"))
	Diff("touch", "test", []byte("diff2"))

	var buf2 []byte
	b2 := GetFullDiffBroker("touch")
	ch2 := b2.Subscribe()
	done2 := concatenate(ch2, &buf2)

	Diff("touch", "test", []byte("diff3"))
	Diff("touch", "test", []byte("diff4"))

	Full("touch", "test", []byte("full1"))
	time.Sleep(10 * time.Millisecond)

	var buf3 []byte
	b3 := GetFullDiffBroker("touch")
	ch3 := b3.Subscribe()
	done3 := concatenate(ch3, &buf3)

	time.Sleep(10 * time.Millisecond)
	Diff("touch", "test", []byte("diff5"))
	Diff("touch", "test", []byte("diff6"))

	b1.Unsubscribe(ch1)
	b2.Unsubscribe(ch2)
	b3.Unsubscribe(ch3)

	<-done1
	<-done2
	<-done3

	assert.Equal(t, "full1diff5diff6", string(buf1))
	assert.Equal(t, "full1diff5diff6", string(buf2))
	assert.Equal(t, "full1diff5diff6", string(buf3))
}

func TestCleanup(t *testing.T) {
	// Ocisti stare brokere (prethodni testovi)
	SetTTL(time.Nanosecond)
	time.Sleep(2 * time.Nanosecond) // Cekaj TTL
	CleanUpBrokers()
	assert.Len(t, brokers, 0)

	// Novi TTL da mogu testirati sa subscriberima
	SetTTL(10 * time.Millisecond)
	// napuni nekie podatke (kreira brokera)
	Stream("teststream", "testevent", []byte("1"))
	b := GetBufferedBroker("teststream") // dohvati brokera
	assert.NotNil(t, b)
	assert.Len(t, brokers, 1)

	// Subscribe i citanje prva 2 eventa
	msgCh := b.Subscribe()
	assert.Len(t, b.subscribers, 0) // Subscriber nije dodan nije dobio full
	m := <-msgCh
	assert.Equal(t, "1", string(m.Data))

	time.Sleep(5 * time.Millisecond) // cekaj pola vremena do expire
	CleanUpBrokers()
	assert.Len(t, brokers, 1)       // broker ziv (nije expired)
	assert.Len(t, b.subscribers, 1) // subscriber dobio sve fullove

	time.Sleep(11 * time.Millisecond) // Cekaj TTL
	CleanUpBrokers()
	assert.Len(t, brokers, 0)       // nema brokera
	assert.Len(t, b.subscribers, 0) // nema subscribera

	m = <-msgCh
	assert.Nil(t, m) // potvrdi da je closan channel
}

//
func TestUnsubscribeBuffered(t *testing.T) {
	topic := "unsubscribe_test"
	b := GetBufferedBroker(topic)
	assert.NotNil(t, b)
	assert.Len(t, b.subscribers, 0) // nema subscribera
	gr := runtime.NumGoroutine()    // sacuvaj broj gorutina

	msgChan := b.Subscribe()
	// nema subscribera jer nema poruka i nije zavrsio subscribe
	assert.NotNil(t, msgChan)
	assert.Len(t, b.subscribers, 0)
	assert.Equal(t, gr+1, runtime.NumGoroutine())

	// unsubscribe prije bilo koje poruke nece napraviti nista
	b.Unsubscribe(msgChan)

	// Poslana i procitana poruka dodaju subscribera
	Stream(topic, "testevent", []byte("1"))
	<-msgChan
	assert.Len(t, b.subscribers, 1)
	assert.Equal(t, gr, runtime.NumGoroutine()) // zavrsio subscribe

	// Unsubscribe drugi puta mice subscribera
	b.Unsubscribe(msgChan)
	assert.Len(t, b.subscribers, 0)
}
