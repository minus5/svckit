package broker

import (
	"sort"

	"github.com/minus5/svckit/amp"
)

type fullDiffCache struct {
	full    *amp.Msg   // last full message
	diffs   []*amp.Msg // previous diff messages
	current []*amp.Msg // memoization of Current function
}

func newFullDiffCache() *fullDiffCache {
	return &fullDiffCache{
		diffs: make([]*amp.Msg, 0),
	}
}

// message for subsribers after he subscribes with ts
func (t *fullDiffCache) Find(ts int64) []*amp.Msg {
	if len(t.diffs) > 0 && ts >= t.diffs[0].Ts && ts <= t.diffs[len(t.diffs)-1].Ts {
		return t.diffsAfter(ts)
	}
	if t.full == nil {
		return nil
	}
	return t.Current()
}

// updateCache adds new message to the caches t.full or t.diffs
func (t *fullDiffCache) Add(m *amp.Msg) {
	t.current = nil

	if m.IsFull() {
		if m.IsReplay() && t.full != nil {
			return
		}
		if t.full != nil { // preserve all after previous full
			t.compactDiffs(t.full.Ts)
		}
		t.full = m
		return
	}

	t.diffs = append(t.diffs, m)
	if len(t.diffs) > 1 {
		prev := len(t.diffs) - 2
		if m.Ts <= t.diffs[prev].Ts {
			t.sortDiffs()
		}
	}
}

// compactDiffs preserves only diffs with Ts greater than input ts
func (t *fullDiffCache) compactDiffs(ts int64) {
	var n []*amp.Msg
	for _, m := range t.diffs {
		if m.Ts >= ts {
			n = append(n, m)
		}
	}
	t.diffs = n
}

// sortDiffs sorts and removes duplicates in t.diffs
func (t *fullDiffCache) sortDiffs() {
	sort.Slice(t.diffs, func(i, j int) bool {
		return t.diffs[i].Ts < t.diffs[j].Ts
	})
	// remove duplicates
	for i := 0; i < len(t.diffs)-1; i++ {
		m1 := t.diffs[i]
		m2 := t.diffs[i+1]
		if m1.Ts == m2.Ts {
			if m1.IsReplay() {
				t.diffs = append(t.diffs[:i], t.diffs[i+1:]...) //remove i
				continue
			}
			j := i + 1
			t.diffs = append(t.diffs[:j], t.diffs[j+1:]...) //remove i+1
		}
	}
}

func (t *fullDiffCache) diffsAfter(ts int64) []*amp.Msg {
	var d []*amp.Msg
	for _, m := range t.diffs {
		if m.Ts > ts {
			d = append(d, m)
		}
	}
	return d
}

func (t *fullDiffCache) Current() []*amp.Msg {
	if t.full == nil {
		return nil
	}
	if t.current == nil {
		t.current = append([]*amp.Msg{t.full}, t.diffsAfter(t.full.Ts)...)
	}
	return t.current
}

func (t *fullDiffCache) FindFor(cTs int64, m *amp.Msg) uint8 {
	if m.IsFull() {
		if cTs != tsNone {
			return sendNothing
		}
		return sendCurrent
	}

	if cTs == m.Ts || cTs == tsNone {
		return sendNothing
	}
	if m.IsReplay() && cTs >= m.Ts { // nemoj ponavljati replay poruke onima koji ih vec imaju
		return sendNothing
	}
	return sendMsg
}
