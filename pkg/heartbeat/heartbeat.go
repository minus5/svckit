// Package heartbeat obuhvaca baratanje heartbeatima. Vodi evidenciju o nizu heartbeatova (identificirani integerom) te odgovara na pitanje je li hearbeat OK.
package heartbeat

import (
	"sync"
	"time"
)

var (
	heartbeats = map[int]*heartbeat{}
	lock       sync.RWMutex
)

type heartbeat struct {
	last  time.Time
	limit time.Duration
}

func new(limit time.Duration) *heartbeat {
	return &heartbeat{
		last:  time.Now(),
		limit: limit,
	}
}

func (h *heartbeat) ok() bool {
	return time.Since(h.last) < h.limit
}

func get(id int) (*heartbeat, bool) {
	lock.RLock()
	h, ok := heartbeats[id]
	lock.RUnlock()
	return h, ok
}

// LastIn vraca true ako je zadnji heartbeat mladji od durationa d
func LastIn(id int, d time.Duration) bool {
	h, ok := get(id)
	if !ok {
		return false
	}
	return time.Since(h.last) < d
}

// New stvara novi heartbeat s danim id-em i limitom.
func New(id int, limit time.Duration) {
	h := new(limit)
	lock.Lock()
	heartbeats[id] = h
	lock.Unlock()
}

// OK vraca je li odredjeni hearbeat (po id-u) OK.
func OK(id int) bool {
	h, ok := get(id)
	if !ok {
		return false
	}
	return h.ok()
}

// Heartbeat setira zadnju pojavu odredjenog hearbeata (po id-u) na dano vrijeme.
func Heartbeat(id int, tm time.Time) {
	h, ok := get(id)
	if !ok {
		return
	}
	h.last = tm
}
