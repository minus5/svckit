package broker

type ring struct {
	head int
	tail int
	buff [][]byte
}

func newRingBuffer(size int) *ring {
	r := &ring{}
	r.setCapacity(size)
	return r
}

func (r *ring) setCapacity(size int) {
	r.checkInit()
	r.extend(size)
}

func (r ring) capacity() int {
	return len(r.buff)
}

func (r *ring) dequeue() []byte {
	r.checkInit()
	if r.head == -1 {
		return nil
	}
	v := r.getOne(r.tail)
	if r.tail == r.head {
		r.head = -1
		r.tail = 0
	} else {
		r.tail = r.mod(r.tail + 1)
	}
	return v
}

func (r *ring) peek() []byte {
	r.checkInit()
	if r.head == -1 {
		return nil
	}
	return r.getOne(r.tail)
}

func (r *ring) values() [][]byte {
	if r.head == -1 {
		return [][]byte{}
	}
	arr := make([][]byte, 0, r.capacity())
	for i := 0; i < r.capacity(); i++ {
		idx := r.mod(i + r.tail)
		arr = append(arr, r.getOne(idx))
		if idx == r.head {
			break
		}
	}
	return arr
}

func (r *ring) set(p int, v []byte) {
	r.buff[r.mod(p)] = v
}

func (r *ring) getOne(p int) []byte {
	return r.buff[r.mod(p)]
}

func (r *ring) mod(p int) int {
	return p % len(r.buff)
}

func (r *ring) checkInit() {
	if r.buff == nil {
		r.buff = make([][]byte, defaultSize)
		for i := range r.buff {
			r.buff[i] = nil
		}
		r.head, r.tail = -1, 0
	}
}

func (r *ring) extend(size int) {
	if size == len(r.buff) {
		return
	} else if size < len(r.buff) {
		r.buff = r.buff[0:size]
	}
	newb := make([][]byte, size-len(r.buff))
	for i := range newb {
		newb[i] = nil
	}
	r.buff = append(r.buff, newb...)
}

func (r *ring) put(msg []byte) {
	r.checkInit()
	r.set(r.head+1, msg)
	old := r.head
	r.head = r.mod(r.head + 1)
	if old != -1 && r.head == r.tail {
		r.tail = r.mod(r.tail + 1)
	}
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
