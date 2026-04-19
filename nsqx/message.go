package nsqx

import (
	"time"

	"github.com/minus5/go-nsqx"
)

// Message wraps nsqx.Message so callers dont need to import go-nsqx directly
type Message struct {
	nsqm        *nsqx.Message
	ID          nsqx.MessageID
	Body        []byte
	Timestamp   int64 // unix nanoseconds
	Attempts    uint16
	NSQDAddress string
}

func newMessage(m *nsqx.Message) *Message {
	body := make([]byte, len(m.Body))
	copy(body, m.Body)
	return &Message{
		nsqm:        m,
		ID:          m.ID,
		Body:        body,
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
