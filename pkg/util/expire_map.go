package util

import (
	"sync"
	"time"
)

//required interface
type ExpireMapEntry interface {
	Id() string
	IsExpired() bool	
}

//optional
//will be called when Entry is expired from map
type ExpireMapEntryCallback interface {
	Expire()
}

type ExpireMapRemoveHandler func(ExpireMapEntry)

type ExpireMap struct {
	values map[string]ExpireMapEntry
	valuesMutex sync.RWMutex
	cleanupTicker *time.Ticker
	removeHandler ExpireMapRemoveHandler
	addHandler ExpireMapRemoveHandler
}

func NewExpireMap(cleanupInterval time.Duration, 
	removeHandler ExpireMapRemoveHandler,
	addHandler ExpireMapRemoveHandler) *ExpireMap{
	m := &ExpireMap{
		values: make(map[string]ExpireMapEntry),
		removeHandler: removeHandler,
		addHandler: addHandler,
	}
	if (cleanupInterval > 0) {
		m.cleanupTicker = time.NewTicker(cleanupInterval)
		go func() {
			for _ = range m.cleanupTicker.C {
				m.Cleanup()
			}
		}()
	}
	return m
}

func (m *ExpireMap) Find(id string) (ExpireMapEntry, bool) {
	m.valuesMutex.RLock()
	defer m.valuesMutex.RUnlock()
	v, found := m.values[id]
	return v, found
}

func (m *ExpireMap) Each(handler func(ExpireMapEntry)) { 
	m.valuesMutex.RLock()
	defer m.valuesMutex.RUnlock()
	for _, e := range m.values {
		handler(e)
	}
}

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
	if m.addHandler != nil && entry != nil{
		m.addHandler(entry)
	}
}

func (m *ExpireMap) Remove(entry ExpireMapEntry) {
	m.valuesMutex.Lock()
	delete(m.values, entry.Id())
	m.valuesMutex.Unlock()
	m.callRemoveHandler(entry)
}

func (m *ExpireMap) RemoveId(id string) {
	if e, found := m.Find(id); found {
		m.Remove(e)
	}
}

func (m *ExpireMap) Size() int {
	return len(m.values)
}

func (m *ExpireMap) Cleanup() {
	for _, e := range m.values {
		if e.IsExpired() {
			m.Remove(e)
			//if entry implements on expire callback
			if cb, ok := e.(ExpireMapEntryCallback); ok {
				cb.Expire()
			}
		}
	}
}

func (m *ExpireMap) Close() {
	if (m.cleanupTicker != nil) {
		m.cleanupTicker.Stop()
	}
}
