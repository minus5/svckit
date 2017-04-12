package util

import (
	"sync"
	"time"
)

// ExpireMapEntry required interface
type ExpireMapEntry interface {
	Id() string
	IsExpired() bool
}

// ExpireMapEntryCallback optional
// will be called when Entry is expired from map
type ExpireMapEntryCallback interface {
	Expire()
}

// ExpireMapRemoveHandler tip za add i remove handlere mape
type ExpireMapRemoveHandler func(ExpireMapEntry)

// ExpireMap mapa za elemete koji implemtiraju ExpireMapEntry interface
type ExpireMap struct {
	values        map[string]ExpireMapEntry
	valuesMutex   sync.RWMutex
	cleanupTicker *time.Ticker
	removeHandler ExpireMapRemoveHandler
	addHandler    ExpireMapRemoveHandler
}

// NewExpireMap kreira novu expire mapu
// - cleanupInterval == 0 NIJE aktivno automatsko ciscenje mape
// - removeHandler poziva se prilikom micanja entry-a iz mape
// - addHandler poziva se prilikom dodavanja entry-a u mapu
func NewExpireMap(cleanupInterval time.Duration,
	removeHandler ExpireMapRemoveHandler,
	addHandler ExpireMapRemoveHandler) *ExpireMap {
	m := &ExpireMap{
		values:        make(map[string]ExpireMapEntry),
		removeHandler: removeHandler,
		addHandler:    addHandler,
	}
	if cleanupInterval > 0 {
		m.cleanupTicker = time.NewTicker(cleanupInterval)
		go func() {
			for _ = range m.cleanupTicker.C {
				m.Cleanup()
			}
		}()
	}
	return m
}

// Find promalazi entry po Id-u
func (m *ExpireMap) Find(id string) (ExpireMapEntry, bool) {
	m.valuesMutex.RLock()
	defer m.valuesMutex.RUnlock()
	v, found := m.values[id]
	return v, found
}

// Each izvrsava handler za sve elemente mape
func (m *ExpireMap) Each(handler func(ExpireMapEntry)) {
	m.valuesMutex.RLock()
	defer m.valuesMutex.RUnlock()
	for _, e := range m.values {
		handler(e)
	}
}

// Add dodaje entry u mapu, ili zamijenjuje u mapi
// - za dodan entry biti ce pozvan addHandler
// - za zamijenjen entry-u (onaj koji se izbacije) biti ce pozvan removeHandler
func (m *ExpireMap) Add(entry ExpireMapEntry) {
	m.valuesMutex.Lock()
	removed, found := m.values[entry.Id()]
	if found {
		delete(m.values, entry.Id())
	}
	m.values[entry.Id()] = entry
	m.valuesMutex.Unlock()
	m.callRemoveHandler(removed)
	m.callAddHandler(entry)
}

func (m *ExpireMap) callRemoveHandler(entry ExpireMapEntry) {
	if m.removeHandler != nil && entry != nil {
		m.removeHandler(entry)
	}
}

func (m *ExpireMap) callAddHandler(entry ExpireMapEntry) {
	if m.addHandler != nil && entry != nil {
		m.addHandler(entry)
	}
}

func (m *ExpireMap) callExpireHandler(entry ExpireMapEntry) {
	if nil == entry {
		return
	}
	//if entry implements on expire callback
	if cb, ok := entry.(ExpireMapEntryCallback); ok {
		cb.Expire()
	}
}

// Remove mice entry iz mape
// - poziva removeHandler za izbacen entry
func (m *ExpireMap) Remove(entry ExpireMapEntry) {
	m.valuesMutex.Lock()
	delete(m.values, entry.Id())
	m.valuesMutex.Unlock()
	m.callRemoveHandler(entry)
}

// RemoveId mice entry iz mape po Id-u ako ga pronadje
// - poziva removeHandler za izbacen entry
func (m *ExpireMap) RemoveId(id string) {
	if e, found := m.Find(id); found {
		m.Remove(e)
	}
}

// Size vraca trenutan broj entry-a u mapi
func (m *ExpireMap) Size() int {
	return len(m.values)
}

// expired vraca mapu sa svim expired entry-ima kao priprema za cleanup
func (m *ExpireMap) expired() map[string]ExpireMapEntry {
	exp := make(map[string]ExpireMapEntry)
	m.valuesMutex.RLock()
	defer m.valuesMutex.RUnlock()
	for _, e := range m.values {
		if e.IsExpired() {
			exp[e.Id()] = e
		}
	}
	return exp
}

// Cleanup cisti iz mape entry-e koji su expired
func (m *ExpireMap) Cleanup() {
	exp := m.expired()
	for _, e := range exp {
		m.Remove(e)
		m.callExpireHandler(e)
	}
}

// Close zatvara koristenje mape
// - zaustavalja cleanup ticker ako je pokrenut
func (m *ExpireMap) Close() {
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}
}
