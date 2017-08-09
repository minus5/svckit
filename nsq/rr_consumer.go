package nsq

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/minus5/svckit/log"
)

var (
	RequeueDelay = 2 * time.Second
)

// RrConsumer request reponse consumer.
// Implements consumer side of the request response communication over nsq.
type RrConsumer struct {
	topic     string
	sub       *Consumer
	producers map[string]*Producer
	sync.Mutex
}

// RrSub creates RrConsumer
// topic   - nsq topic where reuqest arrive
// handler - gets message type and body and creates response (or error)
func RrSub(topic string, handler func(string, []byte) (interface{}, error)) *RrConsumer {
	s := &RrConsumer{
		topic:     topic,
		producers: make(map[string]*Producer),
	}
	h := func(m *Message) error {
		eReq, err := NewEnvelope(m.Body)
		if err != nil {
			return err
		}
		if eReq.Expired() {
			log.S("type", eReq.Type).S("correlationId", eReq.CorrelationId).Info("expired")
			return nil
		}
		rsp, err := handler(eReq.Type, eReq.Body)
		if err != nil {
			m.RequeueWithoutBackoff(RequeueDelay)
			log.S("type", eReq.Type).S("correlationId", eReq.CorrelationId).Error(err)
			return nil
		}
		if eReq.ReplyTo == "" {
			return nil
		}
		eRsp, err := eReq.Reply(rsp)
		if err != nil {
			return err
		}
		pub := s.pub(eReq.ReplyTo)
		if err := pub.Publish(eRsp.Bytes()); err != nil {
			return err
		}
		return nil
	}
	s.sub = Sub(topic, h)
	return s
}

// RrAsyncSub creates RrConsumer in async mode
// Hendler gets type, correlationId, and body.
// It is users reposibility to call Pub with that correlationId and response.
func RrAsyncSub(topic string, handler func(string, string, []byte) error) *RrConsumer {
	h := func(m *Message) error {
		eReq, err := NewEnvelope(m.Body)
		if err != nil {
			return err
		}
		if eReq.Expired() {
			log.S("type", eReq.Type).S("correlationId", eReq.CorrelationId).Info("expired")
			return nil
		}
		correlationId := fmt.Sprintf("%s|%s|%s", eReq.Type, eReq.CorrelationId, eReq.ReplyTo)
		if err := handler(eReq.Type, correlationId, eReq.Body); err != nil {
			return err
		}
		return nil
	}
	return &RrConsumer{
		topic:     topic,
		sub:       Sub(topic, h),
		producers: make(map[string]*Producer),
	}
}

// Pub replay for asycn sub.
// correlationId is the one recived in handler send to RrAsyncSub.
func (s *RrConsumer) Pub(correlationId string, body []byte) error {
	parts := strings.SplitN(correlationId, "|", 3)
	if len(parts) != 3 {
		return fmt.Errorf("wrong correlationId: %s", correlationId)
	}
	eRsp := &Envelope{
		Type:          parts[0],
		CorrelationId: parts[1],
		ReplyTo:       parts[2],
		Body:          body,
	}
	pub := s.pub(eRsp.ReplyTo)
	if err := pub.Publish(eRsp.Bytes()); err != nil {
		return err
	}
	return nil
}

func (s *RrConsumer) pub(topic string) *Producer {
	s.Lock()
	defer s.Unlock()
	if p, ok := s.producers[topic]; ok {
		return p
	}
	p := Pub(topic)
	s.producers[topic] = p
	return p
}

// Close implements gracefully stop.
func (s *RrConsumer) Close() {
	if s.sub == nil {
		return
	}
	s.sub.Close()
}
