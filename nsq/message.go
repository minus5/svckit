package nsq

import (
	"time"

	nsqx "github.com/minus5/go-nsqx"
)

// Presipavam da klijent ne bi morao referencirati go-nsqx package.
type Message struct {
	nsqm        *nsqx.Message
	ID          nsqx.MessageID
	Body        []byte
	Timestamp   int64 // unix nanoseconds — preserved from original API
	Attempts    uint16
	NSQDAddress string
}

func newMessage(m *nsqx.Message) *Message {
	return &Message{
		nsqm:        m,
		ID:          m.ID,
		Body:        m.Body,
		Timestamp:   m.Timestamp.UnixNano(),
		Attempts:    m.Attempts,
		NSQDAddress: m.NSQDAddress,
	}
}

func (m *Message) RequeueWithoutBackoff(delay time.Duration) {
	m.nsqm.RequeueWithoutBackoff(delay)
}

func (m *Message) Touch() {
	m.nsqm.Touch()
}
