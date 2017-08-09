package nsq

import (
	"time"

	gonsq "github.com/nsqio/go-nsq"
)

// Presipavam da klijent ne bi morao referencirati go-nsq package.
type Message struct {
	nsqm        *gonsq.Message
	ID          gonsq.MessageID
	Body        []byte
	Timestamp   int64
	Attempts    uint16
	NSQDAddress string
}

func newMessage(m *gonsq.Message) *Message {
	return &Message{
		nsqm:        m,
		ID:          m.ID,
		Body:        m.Body,
		Timestamp:   m.Timestamp,
		Attempts:    m.Attempts,
		NSQDAddress: m.NSQDAddress,
	}
}

func (m *Message) RequeueWithoutBackoff(delay time.Duration) {
	m.nsqm.RequeueWithoutBackoff(delay)
}
