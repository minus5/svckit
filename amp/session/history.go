package session

import (
	"sync"
	"time"
)

const hlen = 16

type history struct {
	items [hlen]*hitem
	index int
	sync.Mutex
}

type hitem struct {
	typ       string
	startedAt time.Time
	endedAt   *time.Time
}

func (h *history) put(typ string) *hitem {
	h.Lock()
	defer h.Unlock()
	hi := &hitem{
		typ:       typ,
		startedAt: time.Now(),
	}
	h.items[h.index%hlen] = hi
	h.index++
	return hi
}

func (hi *hitem) end() {
	t := time.Now()
	hi.endedAt = &t
}

func (h *history) dump() []*hitem {
	h.Lock()
	defer h.Unlock()
	res := []*hitem{}
	for i := 0; i < hlen; i++ {
		if hi := h.items[(h.index+i)%hlen]; hi != nil {
			res = append(res, hi)
		}
	}
	return res
}

func (h *hitem) duration() int {
	if h.endedAt == nil {
		return -1
	}
	return int(h.endedAt.Sub(h.startedAt).Milliseconds())
}
