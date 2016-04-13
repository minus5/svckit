package broker

import (
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
	brokers = make(map[string]*Broker)
}

type Message struct {
	Event string
	Data  []byte
}

func NewMessage(event string, data []byte) *Message {
	return &Message{
		Event: event,
		Data:  data,
	}
}

type state interface {
	put(*Message)
	get() *Message
	emit(chan *Message)
}

type Broker struct {
	topic       string
	state       state
	subscribers map[chan *Message]struct{}
	sync.RWMutex
	updated time.Time
}

func newBroker(topic string) *Broker {
	// log.S("topic", topic).Debug("new broker")
	return &Broker{
		topic:       topic,
		subscribers: make(map[chan *Message]struct{}),
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
	b.state = newRingBuffer(1)
	return b
}

func (b *Broker) State() *Message {
	return b.state.get()
}

func (b *Broker) Remove() {
	b.Lock()
	for ch, _ := range b.subscribers {
		delete(b.subscribers, ch)
		close(ch)
	}
	b.Unlock()
	brokersLock.Lock()
	delete(brokers, b.topic)
	brokersLock.Unlock()
}

func (b *Broker) setSubscriber(ch chan *Message) {
	b.Lock()
	defer b.Unlock()
	b.subscribers[ch] = struct{}{}
}

func (b *Broker) Subscribe() chan *Message {
	// log.S("topic", b.topic).Debug("subscribe")
	ch := make(chan *Message)
	b.setSubscriber(ch)
	if b.state != nil {
		go b.state.emit(ch)
	}
	return ch
}

func (b *Broker) Unsubscribe(ch chan *Message) {
	b.Lock()
	defer b.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

func (b *Broker) full(msg *Message) {
	b.Lock()
	defer b.Unlock()
	b.state.put(msg)
	b.updated = time.Now()
}

func (b *Broker) diff(msg *Message) {
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

func Full(topic, event string, data []byte) {
	msg := NewMessage(event, data)
	GetFullDiffBroker(topic).full(msg)
}

func Diff(topic, event string, data []byte) {
	msg := NewMessage(event, data)
	GetFullDiffBroker(topic).diff(msg)
}

func Stream(topic, event string, data []byte) {
	msg := NewMessage(event, data)
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
	b := NewFullDiffBroker(topic)
	brokers[topic] = b
	return b
}

func createBufferedBroker(topic string, size int) *Broker {
	brokersLock.Lock()
	defer brokersLock.Unlock()
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
			delete(brokers, k)
		}
	}
}
