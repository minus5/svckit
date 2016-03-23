package broker

import (
	"log"
	"sync"
	"time"
)

var (
	brokers     map[string]*Broker
	brokersLock sync.RWMutex
	ttl         time.Duration = time.Hour
	defaultSize int           = 100
)

func init() {
	log.SetFlags(log.Llongfile)
	brokers = make(map[string]*Broker)
}

type state interface {
	put([]byte)
	get() []byte
	emit(chan []byte)
}

type Broker struct {
	topic       string
	state       state
	subscribers map[chan []byte]bool
	sync.RWMutex
	updated time.Time
}

func newBroker(topic string) *Broker {
	return &Broker{
		topic:       topic,
		subscribers: make(map[chan []byte]bool),
		updated:     time.Now(),
	}
}

func NewBufferedBroker(topic string, size int) *Broker {
	b := newBroker(topic)
	b.state = newRingBuffer(size)
	return b
}

func NewFullDiffBroker(topic string) *Broker {
	b := newBroker(topic)
	b.state = &singleBuffer{}
	return b
}

func (b *Broker) State() []byte {
	return b.state.get()
}

func (b *Broker) setSubscriber(ch chan []byte) {
	b.Lock()
	defer b.Unlock()
	b.subscribers[ch] = true
}

func (b *Broker) Subscribe() chan []byte {
	ch := make(chan []byte)
	b.setSubscriber(ch)
	if b.state != nil {
		go b.state.emit(ch)
	}
	return ch
}

func (b *Broker) Unsubscribe(ch chan []byte) {
	b.Lock()
	defer b.Unlock()
	delete(b.subscribers, ch)
	close(ch)
}

func (b *Broker) full(msg []byte) {
	b.Lock()
	defer b.Unlock()
	b.state.put(msg)
	b.updated = time.Now()
}

func (b *Broker) diff(msg []byte) {
	b.RLock()
	defer b.RUnlock()
	for c, _ := range b.subscribers {
		c <- msg
	}
}

func (b *Broker) expired() bool {
	b.RLock()
	defer b.RUnlock()
	return b.updated.Before(time.Now().Add(-ttl))
}

func Full(topic string, msg []byte) {
	GetFullDiffBroker(topic).full(msg)
}

func Diff(topic string, msg []byte) {
	GetFullDiffBroker(topic).diff(msg)
}

func Stream(topic string, msg []byte) {
	GetBufferedBroker(topic).full(msg)
	GetBufferedBroker(topic).diff(msg)
}

func FindBroker(topic string) (*Broker, bool) {
	brokersLock.RLock()
	brokersLock.RUnlock()
	b, ok := brokers[topic]
	return b, ok
}

func createFullDiffBroker(topic string, size int) *Broker {
	brokersLock.Lock()
	defer brokersLock.Unlock()
	log.Printf("creating full-diff broker %s", topic)
	b := NewFullDiffBroker(topic)
	brokers[topic] = b
	return b
}

func createBufferedBroker(topic string, size int) *Broker {
	brokersLock.Lock()
	defer brokersLock.Unlock()
	log.Printf("creating buffered broker %s", topic)
	b := NewBufferedBroker(topic, size)
	brokers[topic] = b
	return b
}

func GetFullDiffBroker(topic string) *Broker {
	b, ok := FindBroker(topic)
	if !ok {
		return createFullDiffBroker(topic, defaultSize)
	}
	return b
}

func GetBufferedBroker(topic string) *Broker {
	b, ok := FindBroker(topic)
	if !ok {
		return createBufferedBroker(topic, defaultSize)
	}
	return b
}

func CleanUpBrokers() {
	brokersLock.Lock()
	defer brokersLock.Unlock()
	for k, b := range brokers {
		if b.expired() {
			log.Printf("deleting broker %s", k)
			delete(brokers, k)
		}
	}
}
