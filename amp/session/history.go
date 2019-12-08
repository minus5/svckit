package session

import (
	"sync"
	"time"
)

const hlen = 32

const (
	actionAlive = iota
	actionSend
	actionExit
	actionUnsubscribe
	actionOut
	actionIn
)

var historyNames = map[uint8]string{
	actionAlive:       "alive",
	actionSend:        "send",
	actionExit:        "exit",
	actionUnsubscribe: "uns",
	actionOut:         "out",
	actionIn:          "in",
}

func newHistory() *history {
	return &history{
		items: [hlen]*hitem{},
	}
}

type history struct {
	items [hlen]*hitem
	index int
	sync.Mutex
}

type hitem struct {
	action    uint8
	typ       uint8
	updateTyp uint8
	value     int
	startedAt time.Time
	endedAt   *time.Time
}

func (h *history) put(action, typ, updateType uint8, value int) *hitem {
	h.Lock()
	defer h.Unlock()
	hi := &hitem{
		action:    action,
		typ:       typ,
		updateTyp: updateType,
		value:     value,
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

func (hi *hitem) name() string {
	return historyNames[hi.action]
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
