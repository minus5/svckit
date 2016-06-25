package merger

import (
	"fmt"
	"math"
	"pkg/msgs"
	"pkg/svckit/log"
	"pkg/svckit/metric"
	"strings"
	"sync"
)

type msg struct {
	typ     string
	channel string
	isFull  bool
	isDiff  bool
	body    interface{}
	no      int
	backend *msgs.Backend
}

func newMsg(m *msgs.Backend) *msg {
	return &msg{
		typ:     m.Type,
		channel: m.RootType(),
		isFull:  m.IsFull(),
		isDiff:  m.IsDiff(),
		no:      m.No,
		backend: m,
	}
}

const undefinedNo = math.MinInt64

const (
	checkLater = iota
	checkSkip
	checkReset
	checkMerge
	checkReplace
	checkCurrent
	checkRequestFull
)

// fullDiffOrderer osigurava da poruke zavrse u pravom redoslijedu
// Na pocetku krene od full-a, zatrazit ce ga pozivom na dopunaHandler.
// Kada ima full na njega dodaje svaki pristigli diff. I izbaci van diff i full.
// Pri tome pazi na redoslijed. Ako neka poruka dodje van redoslijeda dodat ce ju u queue
// pa naknado obraditi, kada dodje ona koja fali.
type fullDiffOrderer struct {
	no            int
	queue         []*msg
	in            chan *msg
	out           chan *msg
	exitSignal    chan struct{}
	oneClose      sync.Once
	current       *msgs.Backend
	dopunaHandler func()
	dopunaAtNo    int
}

func newFullDiffOrderer(dopunaHandler func()) *fullDiffOrderer {
	o := &fullDiffOrderer{
		queue:         make([]*msg, 0),
		in:            make(chan *msg),
		out:           make(chan *msg),
		exitSignal:    make(chan struct{}),
		dopunaHandler: dopunaHandler,
	}
	o.init()
	go o.loop()
	return o
}

func (o *fullDiffOrderer) init() {
	o.no = undefinedNo
	o.dopunaAtNo = undefinedNo
}

func (o *fullDiffOrderer) close() {
	o.oneClose.Do(func() {
		close(o.exitSignal)
	})
}

func (o *fullDiffOrderer) loop() {
	for {
		select {
		case m := <-o.in:
			o.processMsg(m)
			o.processQueue()
		case <-o.exitSignal:
			close(o.out)
			close(o.in)
			return
		}
	}
}

func (o *fullDiffOrderer) processMsg(m *msg) {
	switch o.check(m) {
	case checkSkip:
		metric.Counter("merger.skip")
	case checkMerge:
		o.current.Merge(m.backend)
		o.no = m.no
		o.out <- m                 // diff
		o.out <- newMsg(o.current) // full
		metric.Counter("merger.merge")
	case checkReplace:
		o.current = m.backend
		if o.no == undefinedNo {
			o.no = m.no
			o.out <- m
		}
		metric.Counter("merger.replace")
	case checkCurrent:
		o.out <- m
		metric.Counter("merger.current")
	case checkRequestFull:
		o.queue = append(o.queue, m)
		if o.dopunaAtNo == undefinedNo {
			o.dopunaAtNo = m.no
			o.dopunaHandler()
			log.S("typ", m.typ).I("no", m.no).Debug("requestFull")
			metric.Counter("merger.requestFull")
		} else {
			metric.Counter("merger.requestFullSkip")
		}
	case checkLater:
		o.queue = append(o.queue, m)
		metric.Counter("merger.later")
	case checkReset:
		o.init()
		o.processMsg(m)
		o.processQueue()
		metric.Counter("merger.reset")
		log.S("type", m.typ).I("no", m.no).Notice("reset")
	}
}

func (o *fullDiffOrderer) processQueue() {
again:
	if len(o.queue) == 0 {
		return
	}

	proc := func(i int, m *msg) bool {
		// provjeri da li ju treba vaditi iz queue-a
		r := o.check(m)
		if r == checkLater || r == checkRequestFull {
			return false
		}
		// izvadi iz queue-a
		o.queue = append(o.queue[0:i], o.queue[i+1:]...)
		o.processMsg(m)
		return true
	}

	// najprije full-ovi
	for i, m := range o.queue {
		if m.isFull && proc(i, m) {
			goto again
		}
	}
	// pa onda diff-ovi
	for i, m := range o.queue {
		if m.isDiff && proc(i, m) {
			goto again
		}
	}
}

func (o *fullDiffOrderer) check(m *msg) int {
	if len(o.queue) > 99 {
		return checkReset
	}
	if m.isFull {
		if o.no == undefinedNo {
			return checkReplace
		}
		if o.no == m.no {
			return checkReplace
		}
		if m.no > o.no+2 {
			return checkReset
		}
		if m.no < o.no {
			return checkSkip
		}
	}
	if m.isDiff {
		if o.no == undefinedNo {
			return checkRequestFull
		}
		if o.no+1 == m.no {
			return checkMerge
		}
		if o.no+99 < m.no {
			return checkReset
		}
		if m.no == o.no {
			return checkCurrent
		}
		if m.no < o.no {
			return checkSkip
		}
	}
	return checkLater
}

func (o *fullDiffOrderer) queueSize() int {
	return len(o.queue)
}

func (o *fullDiffOrderer) inQueue() string {
	q := ""
	for _, m := range o.queue {
		if m.isFull {
			q += fmt.Sprintf(" f%d", m.no)
			continue
		}
		q += fmt.Sprintf(" d%d", m.no)
	}
	return strings.TrimSpace(q)
}
