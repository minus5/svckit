package broker

import "sync"

type ring struct {
	size int
	head int
	tail int
	buf  []*Message
	sync.RWMutex
	touched     bool
	touchSignal chan struct{}
	touchOnce   sync.Once
}

func newRingBuffer(size int) *ring {
	r := &ring{
		size:        size,
		head:        1 % size,
		tail:        size - 1,
		buf:         make([]*Message, size),
		touchSignal: make(chan struct{}),
	}
	return r
}

func (r *ring) mod(i int) int {
	return i % r.size
}

func (r *ring) values() []*Message {
	out := make([]*Message, r.size)
	r.RLock()
	defer r.RUnlock()
	for i := 0; i < r.size; i++ {
		ix := r.mod(i + r.head)
		out[i] = r.buf[ix]
	}
	return out
}

func (r *ring) put(msg *Message) {
	r.Lock()
	defer r.Unlock()
	r.buf[r.head] = msg
	r.head = r.mod(r.head + 1)
	r.tail = r.mod(r.tail + 1)
	r.touchOnce.Do(func() {
		r.touched = true
		close(r.touchSignal)
	})
}

func (r *ring) get() *Message {
	for _, line := range r.values() {
		return line
	}
	return nil
}

func (r *ring) emit(ch chan *Message) {
	for _, line := range r.values() {
		if line != nil && len(line.GetData()) > 0 {
			ch <- line
		}
	}
}

func (r *ring) waitTouch() {
	r.RLock()
	touched := r.touched
	r.RUnlock()
	if touched {
		return
	}
	<-r.touchSignal
}
