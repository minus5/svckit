package broker

import "sync"

type ring struct {
	size int
	head int
	tail int
	buf  []*Message
	sync.RWMutex
}

func newRingBuffer(size int) *ring {
	r := &ring{
		size: size,
		head: 1,
		tail: size - 1,
		buf:  make([]*Message, size),
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
}

func (r *ring) get() *Message {
	out := &Message{
		Event: "state",
	}
	for _, line := range r.values() {
		if len(line.Data) > 0 {
			out.Data = append(out.Data, line.Data...)
			out.Data = append(out.Data, '\n')
		}
	}
	return out
}

func (r *ring) emit(ch chan *Message) {
	for _, line := range r.values() {
		if len(line.Data) > 0 {
			ch <- line
		}
	}
}
