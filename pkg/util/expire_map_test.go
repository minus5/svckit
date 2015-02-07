package util

import (
	"github.com/stretchrcom/testify/assert"
	"testing"
)

type testMapEntry struct {
	id string
	isExpired bool
	expireCalled bool
}

func (e *testMapEntry) Id() string {
	return e.id
}

func (e *testMapEntry) IsExpired() bool {
	return e.isExpired
}

func (e *testMapEntry) Expire() {
	e.expireCalled = true
}

func (e *testMapEntry) nonInterfaceFn() {
}

func newTestMapEntry() *testMapEntry {
	return &testMapEntry{
		id: Uuid(),
		isExpired: false,
		expireCalled: false, 
	}
}

func TestAddRemove(t *testing.T) {
	m := NewExpireMap(0, nil, nil)
	e := newTestMapEntry()
	m.Add(e)	
	assert.NotNil(t, m)
	assert.NotNil(t, e)
	assert.Equal(t, m.Size(), 1)
	m.Add(e)	
	assert.Equal(t, m.Size(), 1)
	e2 := newTestMapEntry()
	m.Add(e2)	
	assert.Equal(t, m.Size(), 2)
	m.Remove(e)
	assert.Equal(t, m.Size(), 1)
}

func TestRemoveOfMissing(t *testing.T) {
	m := NewExpireMap(0, nil, nil)
	e := newTestMapEntry()
	e2 := newTestMapEntry()
	m.Add(e)
	assert.Equal(t, m.Size(), 1)
	m.Remove(e2)
	assert.Equal(t, m.Size(), 1)
}

func TestCleanup(t *testing.T) {
	m := NewExpireMap(0, nil, nil)
	e := newTestMapEntry()
	e2 := newTestMapEntry()
	m.Add(e)
	m.Add(e2)
	assert.Equal(t, m.Size(), 2)
	m.Cleanup()
	assert.Equal(t, m.Size(), 2)
	e.isExpired = true;
	m.Cleanup()
	assert.Equal(t, m.Size(), 1)
	assert.True(t, e.expireCalled)
	assert.False(t, e2.expireCalled)
}

func TestClose(t *testing.T) {
	m := NewExpireMap(1, nil, nil)
	m.Close()
}

func TestFind(t *testing.T) {
	m := NewExpireMap(0, nil, nil)
	e := newTestMapEntry()
	m.Add(e)	
	e2, found := m.Find(e.Id())
	assert.True(t, found)
	assert.Equal(t, e2, e)
	e22, ok := e2.(*testMapEntry)
	assert.True(t, ok)
	e22.nonInterfaceFn()
}

func TestEach(t *testing.T) {
	m := NewExpireMap(0, nil, nil)
	m.Add(newTestMapEntry())
	testEachCount := func(expected int) {
		called := 0
		m.Each(func(e ExpireMapEntry) {
			called++
		})
		assert.Equal(t, expected, called)
	}
	testEachCount(1)
	m.Add(newTestMapEntry())
	testEachCount(2)
	e3 :=  newTestMapEntry()
	m.Add(e3)
	testEachCount(3)
	m.Remove(e3)
	testEachCount(2)
}

func TestHandlers(t *testing.T) {
	var removed, added ExpireMapEntry
	m := NewExpireMap(0, 
		func(e ExpireMapEntry) { removed = e}, 
		func(e ExpireMapEntry) { added = e})
	e1 := &testMapEntry{id: "1"}
	e2 := &testMapEntry{id: "2"}
	m.Add(e1)
	assert.Nil(t, removed)
	assert.Equal(t, added, e1)
	m.Add(e2)
	assert.Nil(t, removed)
	assert.Equal(t, added, e2)
	removed, added = nil, nil
	m.Remove(e2)
	assert.Equal(t, removed, e2)
	assert.Nil(t, added)
	removed, added = nil, nil
	m.Add(e1)
	assert.Equal(t, removed, e1)
	assert.Equal(t, added, e1) 
}
