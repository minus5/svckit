/*
Package sorter sorts incoming messages in strictly increasing order.
If messages are not in order package waits ttw for missing messages.
If missing messages arrive during ttw sorter will output messages in order.
If ttw exceeds before missing messages arrive sorter will output queued messages.
Usage:
   	s := sorterNew(time.Second)

    // consuming sorter messages
    go func() {
      for m := range s.Output() {
         ...
      }
      // point of clean exit after close
    }()

    // adding messages to sorter
    ...
    s.Push(1, o)
    s.Push(2, o)
    ...
    s.Close()
*/
package sorter

import (
	"sort"
	"time"
)

type sorter struct {
	current int
	ttw     time.Duration
	queue   map[int]*Msg
	input   chan *Msg
	Output  chan *Msg
}

// Msg message to push to sorter.
// No is order number of the message.
// Body is unused in sorter.
type Msg struct {
	No   int
	Body interface{}
}

// New creates new sorter
// Ttw - time to wait for missing messages
func New(ttw time.Duration) *sorter {
	s := &sorter{
		current: 0,
		ttw:     ttw,
		queue:   make(map[int]*Msg),
		input:   make(chan *Msg),
		Output:  make(chan *Msg),
	}
	go s.loop()
	return s
}

func (s *sorter) loop() {
	var timer <-chan time.Time
	scheduleTimer := func() {
		timer = time.After(s.ttw)
	}
	for {
		select {
		case m := <-s.input:
			if m == nil {
				s.purge()
				close(s.Output)
				return
			}
			s.add(m)
			s.processQueue()
			if s.empty() {
				timer = nil
			} else {
				scheduleTimer()
			}
		case <-timer:
			s.purge()
		}
	}
}

// Push adds new message to sorter.
func (s *sorter) Push(m *Msg) {
	s.input <- m
}

// Close closes sorter.
// Will pruge any messages in queue.
func (s *sorter) Close() {
	close(s.input)
}

// Reset sorter.
// Will purge any messages in queue.
func (s *sorter) Reset() {
	s.purge()
	s.current = 0
}

func (s *sorter) add(m *Msg) {
	if m.No <= s.current+1 {
		s.out(m)
		return
	}
	s.addToQueue(m)
}

func (s *sorter) empty() bool {
	return len(s.queue) == 0
}

func (s *sorter) addToQueue(m *Msg) {
	s.queue[m.No] = m
}

func (s *sorter) processQueue() {
again:
	for k, v := range s.queue {
		if k <= s.current+1 {
			delete(s.queue, k)
			s.out(v)
			goto again
		}
	}
}

func (s *sorter) out(m *Msg) {
	if m.No > s.current {
		s.current = m.No
	}
	s.Output <- m
}

func (s *sorter) purge() {
	var nos []int
	for k := range s.queue {
		nos = append(nos, k)
	}
	sort.Ints(nos)
	for _, no := range nos {
		s.out(s.queue[no])
		delete(s.queue, no)
	}
}
