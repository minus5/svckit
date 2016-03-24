package broker

import "sync"

type ring struct {
	size int
	head int
	tail int
	buf  [][]byte
	sync.RWMutex
}

func newRingBuffer(size int) *ring {
	r := &ring{
		size: size,
		head: 1,
		tail: size - 1,
		buf:  make([][]byte, size),
	}
	return r
}

func (r *ring) mod(i int) int {
	return i % r.size
}

func (r *ring) values() [][]byte {
	out := make([][]byte, r.size)
	r.RLock()
	defer r.RUnlock()
	for i := 0; i < r.size; i++ {
		ix := r.mod(i + r.head)
		out[i] = r.buf[ix]
	}
	return out
}

func (r *ring) put(msg []byte) {
	r.Lock()
	defer r.Unlock()
	r.buf[r.head] = msg
	r.head = r.mod(r.head + 1)
	r.tail = r.mod(r.tail + 1)
}

func (r *ring) get() []byte {
	var out []byte
	for _, line := range r.values() {
		out = append(out, line...)
		out = append(out, '\n')
	}
	return out
}

func (r *ring) emit(ch chan []byte) {
	for _, line := range r.values() {
		ch <- line
	}
}
