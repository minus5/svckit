package merger

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

const undefinedNo = math.MinInt64
const maxQueueSize = 256
const defaultOOOLimit = 16

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
	no            int64
	queue         []*msg
	in            chan *msg
	out           chan *msg
	exitSignal    chan struct{}
	oneClose      sync.Once
	current       *msg
	dopunaHandler func()
	dopunaAtNo    int64
	oooLimit      int64 //out of order limit, odredjuje koliko diffova ceka prije trazenja fulla
	changedAt     time.Time
}

func newFullDiffOrderer(dopunaHandler func()) *fullDiffOrderer {
	o := &fullDiffOrderer{
		queue:         make([]*msg, 0),
		in:            make(chan *msg),
		out:           make(chan *msg),
		exitSignal:    make(chan struct{}),
		dopunaHandler: dopunaHandler,
		oooLimit:      defaultOOOLimit,
		changedAt:     time.Now(),
	}
	o.init()
	go o.loop()
	return o
}

func (o *fullDiffOrderer) init() {
	o.no = undefinedNo
	o.dopunaAtNo = undefinedNo
	o.queue = make([]*msg, 0)
}

//noinspection GoReservedWordUsedAsName
func (o *fullDiffOrderer) close() {
	o.oneClose.Do(func() {
		close(o.exitSignal)
	})
}

func (o *fullDiffOrderer) loop() {
	for {
		select {
		case m := <-o.in:
			o.changedAt = time.Now()
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
	r := o.check(m)
	o.processMsgWithResult(m, r)
}

func (o *fullDiffOrderer) processMsgWithResult(m *msg, r int) {
	switch r {
	case checkSkip:
		mtrc.Counter("skip")
	case checkMerge:
		o.current.Merge(m)
		o.no = m.no
		o.out <- m         // diff
		o.out <- o.current // full
		mtrc.Counter("merge")
	case checkReplace:
		o.current = m
		if o.no == undefinedNo {
			o.no = m.no
			o.out <- m
		}
		mtrc.Counter("replace")
	case checkCurrent:
		o.out <- m
		mtrc.Counter("current")
	case checkRequestFull:
		o.queue = append(o.queue, m)
		if o.dopunaAtNo == undefinedNo {
			o.dopunaAtNo = m.no
			o.dopunaHandler()
			//TODO log.S("typ", m.typ).I("no", m.no).Debug("requestFull")
			mtrc.Counter("requestFull")
		} else {
			mtrc.Counter("requestFullSkip")
		}
	case checkLater:
		o.queue = append(o.queue, m)
		mtrc.Counter("later")
	case checkReset:
		o.init()
		o.processMsg(m)
		mtrc.Counter("reset")
		//TODO log.S("type", m.typ).I("no", m.no).ErrorS("reset")
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
		o.processMsgWithResult(m, r)
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
	if len(o.queue) >= maxQueueSize {
		return checkReset
	}
	if m.isFull {
		if o.no == undefinedNo {
			return checkReplace
		}
		if o.no == m.no {
			return checkReplace
		}
		if m.no > o.no+2 || m.no <= 1 {
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
		if m.no == o.no {
			return checkCurrent
		}
		if m.no < o.no {
			return checkSkip
		}
		if m.no > o.no+o.oooLimit {
			return checkReset
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

func (o *fullDiffOrderer) expired() bool {
	return o.changedAt.Before(time.Now().Add(-3 * time.Hour))
}
