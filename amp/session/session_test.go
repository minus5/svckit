package session

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/minus5/svckit/amp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConn struct {
	in  chan []byte
	out chan []byte
}

func (c *mockConn) Read() ([]byte, error) {
	m, ok := <-c.in
	if !ok {
		return nil, io.EOF
	}
	if m == nil {
		fmt.Printf("nilllll\n")
	}
	return m, nil
}
func (c *mockConn) Write(payload []byte, deflated bool) error {
	c.out <- payload
	return nil
}
func (c *mockConn) DeflateSupported() bool     { return false }
func (c *mockConn) Headers() map[string]string { return nil }
func (c *mockConn) No() uint64                 { return 0 }
func (c *mockConn) Meta() map[string]string    { return nil }
func (c *mockConn) Close() error {
	close(c.in)
	return nil
}
func (c *mockConn) SetMeta(m map[string]string) {}

type mockBroker struct{}

func (b *mockBroker) Subscribe(amp.Sender, map[string]int64) {}
func (b *mockBroker) Unsubscribe(amp.Sender)                 {}
func (b *mockBroker) Wait()                                  {}

type mockRequester struct {
	t *testing.T

	WantSendCalls int
	gotSendCalls  int
}

func (r *mockRequester) Send(subscriber amp.Subscriber, msg *amp.Msg) {
	r.gotSendCalls++

	require.True(r.t, r.WantSendCalls >= r.gotSendCalls)
}

func (r *mockRequester) Unsubscribe(amp.Subscriber) {}

func (r *mockRequester) Wait() {}

func (r *mockRequester) Assert(t *testing.T) {
	require.Equal(t, r.WantSendCalls, r.gotSendCalls)
}

func testSession(outLen, inLen int) (chan []byte, chan []byte, func(), chan struct{}, func(*amp.Msg)) {
	out := make(chan []byte, outLen)
	in := make(chan []byte, inLen)

	ctx, cancel := context.WithCancel(context.Background())
	conn := &mockConn{out: out, in: in}
	done := make(chan struct{})

	s := &session{
		conn:        conn,
		outMessages: make(chan []*amp.Msg, 256),
		requester:   &mockRequester{},
		broker:      &mockBroker{},
	}
	go func() {
		s.loop(ctx)
		close(done)
	}()

	return out, in, cancel, done, s.Send
}

func TestAlive(t *testing.T) {
	aliveBefore := aliveInterval
	aliveInterval = time.Millisecond

	out, _, cancel, done, _ := testSession(1024, 0)
	time.Sleep(16 * time.Millisecond)
	cancel()
	<-done
	msgs := len(out)
	t.Logf("number of alive messages %d, example: %s", len(out), <-out)
	assert.True(t, msgs > 1)
	assert.True(t, msgs <= 16)
	assert.Equal(t, `{"t":6}`+"\n", string(<-out))

	aliveInterval = aliveBefore
}

func ping(cid uint64) *amp.Msg {
	return &amp.Msg{Type: amp.Ping, CorrelationID: cid}
}

func msgCID(buf []byte) int {
	m := amp.Parse(buf)
	return int(m.CorrelationID)
}

func TestOrderedMessages(t *testing.T) {
	out, _, cancel, done, send := testSession(3, 3)

	send(ping(1))
	send(ping(2))
	send(ping(3))
	assert.Equal(t, 1, msgCID(<-out))
	assert.Equal(t, 2, msgCID(<-out))
	assert.Equal(t, 3, msgCID(<-out))

	cancel()
	<-done
}

func TestPingPong(t *testing.T) {
	out, in, cancel, done, _ := testSession(3, 3)

	in <- ping(1).Marshal()
	in <- ping(2).Marshal()
	in <- ping(3).Marshal()
	assert.Equal(t, 1, msgCID(<-out))
	assert.Equal(t, 2, msgCID(<-out))
	assert.Equal(t, 3, msgCID(<-out))

	cancel()
	<-done
}

func TestQueueDrain(t *testing.T) {
	out := make(chan []byte, 1000)
	in := make(chan []byte, 1000)
	done := make(chan struct{})
	conn := &mockConn{out: out, in: in}
	s := &session{
		conn:        conn,
		outMessages: make(chan []*amp.Msg, 256),
		requester:   &mockRequester{},
		broker:      &mockBroker{},
	}
	for i := 0; i < 200; i++ {
		s.Send(&amp.Msg{})
	}
	assert.Len(t, s.outMessages, 200)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		s.loop(ctx)
		close(done)
	}()
	time.Sleep(500 * time.Millisecond)
	cancel()
	<-done
	assert.Len(t, s.outMessages, 0)
}

func Test_session_receive(t *testing.T) {
	type fields struct {
		conn           mockConn
		requester      mockRequester
		topicWhitelist []string
	}

	tests := []struct {
		name   string
		fields fields
		in     *amp.Msg
	}{
		{
			name: "it should call Send on requester if msg topic is whitelisted",
			fields: fields{
				requester: mockRequester{
					t:             t,
					WantSendCalls: 1,
				},
				topicWhitelist: []string{"whitelisted.req"},
			},
			in: &amp.Msg{
				Type: amp.Request,
				URI:  "whitelisted.req/method",
			},
		},
		{
			name: "it should not call Send on requester if msg topic is not whitelisted",
			fields: fields{
				requester: mockRequester{
					t:             t,
					WantSendCalls: 0,
				},
				topicWhitelist: []string{"whitelisted.req"},
			},
			in: &amp.Msg{
				Type: amp.Request,
				URI:  "blacklisted.req/method",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &session{
				conn:           &tt.fields.conn,
				requester:      &tt.fields.requester,
				topicWhitelist: tt.fields.topicWhitelist,
			}

			s.receive(tt.in)

			tt.fields.requester.Assert(t)
		})
	}
}
