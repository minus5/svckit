package sequence

import "sync"

var (
	sequences = map[string]*Sequence{}
	seqLock   sync.RWMutex
)

type Sequence struct {
	name string
	last int
	sync.RWMutex
}

func (s *Sequence) Name() string {
	s.RLock()
	defer s.RUnlock()
	return s.name
}

func (s *Sequence) Last() int {
	s.RLock()
	defer s.RUnlock()
	return s.last
}

func (s *Sequence) next() int {
	s.Lock()
	defer s.Unlock()
	n := s.last + 1
	s.last = n
	return n
}

func Get(name string) (*Sequence, bool) {
	seqLock.RLock()
	defer seqLock.RUnlock()
	s, ok := sequences[name]
	return s, ok
}

func set(name string, s *Sequence) {
	seqLock.Lock()
	defer seqLock.Unlock()
	sequences[name] = s
}

func Next(name string) int {
	s, ok := Get(name)
	if !ok {
		Set(name, 0)
		return 0
	}
	return s.next()
}

func Set(name string, last int) {
	s := &Sequence{
		name: name,
		last: last,
	}
	set(name, s)
}
