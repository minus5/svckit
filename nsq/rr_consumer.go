package nsq

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
)

var (
	RequeueDelay = 2 * time.Second
)

// RrConsumer request reponse consumer.
// Implements consumer side of the request response communication over nsq.
type RrConsumer struct {
	topic           string
	sub             *Consumer
	producers       map[string]*Producer
	consumerOptions []func(*options)
	requeueError    error // set this error to requeue only on this
	// if nil requeues on all errors
	sync.Mutex
}

// RrSub creates RrConsumer
// topic   - nsq topic where reuqest arrive
// handler - gets message type and body and creates response (or error)
func RrSub(topic string, handler func(string, []byte) (interface{}, error), opts ...func(*RrConsumer)) *RrConsumer {
	s := &RrConsumer{
		topic:     topic,
		producers: make(map[string]*Producer),
	}
	s.apply(opts...)
	h := func(m *Message) error {
		// zapakiraj poruku u envelope
		eReq, err := NewEnvelope(m.Body)
		if err != nil {
			return err
		}
		// provjeri da li je expired
		if eReq.Expired() {
			log.S("type", eReq.Type).S("correlationId", eReq.CorrelationId).I("now", int(time.Now().Unix())).I("expires_at", int(eReq.ExpiresAt)).Info("expired")
			return nil
		}
		// radi request
		rsp, handlerErr := handler(eReq.Type, eReq.Body)
		// ako je puklo vrati poruku u nsq
		if handlerErr != nil && (s.requeueError == nil || handlerErr == s.requeueError) {
			m.RequeueWithoutBackoff(RequeueDelay)
			log.S("type", eReq.Type).S("correlationId", eReq.CorrelationId).Error(handlerErr)
			return nil
		}
		// treba li odgovoriti
		if eReq.ReplyTo == "" {
			return nil
		}
		// odgovori
		eRsp, err := eReq.Reply(rsp, handlerErr)
		if err != nil {
			log.Error(err)
			return err
		}
		pub := s.pub(eReq.ReplyTo)
		if err := pub.Publish(eRsp.Bytes()); err != nil {
			log.Error(err)
			return err
		}
		return nil
	}
	s.consumerOptions = append(s.consumerOptions, Channel(env.AppName()))
	s.sub = Sub(topic, h, s.consumerOptions...)
	return s
}

// apply calls all functions to setup options
func (s *RrConsumer) apply(opts ...func(*RrConsumer)) {
	for _, fn := range opts {
		fn(s)
	}
}

func RequeueError(err error) func(*RrConsumer) {
	return func(s *RrConsumer) {
		s.requeueError = err
	}
}

// SetConsumerOptions sets configuration options for the underlying Consumer.
func SetConsumerOptions(opts ...func(*options)) func(*RrConsumer) {
	return func(s *RrConsumer) {
		s.consumerOptions = opts
	}
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

// StartClosing will initiate a graceful stop of the Consumer (permanent)
// Receive on returned chan to block until this process completes
func (s *RrConsumer) StartClosing() chan int {
	if s.sub == nil {
		return nil
	}
	return s.sub.StartClosing()
}

type Server interface {
	Serve(req interface{}) (interface{}, error)
}

func NewRrServer(topic string,
	srv Server,
	typeFor func(string) reflect.Type,
	requeueError error) *RrConsumer {
	handler := func(typ string, body []byte) (interface{}, error) {
		t := typeFor(typ)
		req := reflect.New(t).Interface()
		if err := json.Unmarshal(body, req); err != nil {
			return nil, err
		}
		return srv.Serve(req)
	}
	if requeueError == nil {
		requeueError = errors.New("newer")
	}
	return RrSub(topic, handler, RequeueError(requeueError))
}
