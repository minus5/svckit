package nsq

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/minus5/svckit/log"
)

// ErrTimeout deafult timeout error
var ErrTimeout = errors.New("timeout")

// RrProducer request response producer.
// Implements request response communication over nsq.
type RrProducer struct {
	s         map[string]chan *Envelope
	producers map[string]*Producer
	topic     string
	sub       *Consumer
	msgNo     int
	corr      RrProducerCorrelation
	sync.Mutex
}

// RrProducerCorrelation interface for custom generation of correlation IDs
type RrProducerCorrelation interface {
	// NewCorrelationID creates new correlation ID for message
	NewCorrelationID(topic, typ string, req interface{}) string
}

// RrPub creates new RrProducer.
// - topic is nsq topic where remote services send responses.
// - opts are functions to set additional options
func RrPub(topic string, opts ...func(*RrProducer)) *RrProducer {
	s := &RrProducer{
		msgNo:     rand.Intn(math.MaxInt32),
		s:         make(map[string]chan *Envelope),
		producers: make(map[string]*Producer),
		topic:     topic,
	}
	// Set default calc of correlation id-a
	s.apply(SetRrProducerCorrelation(s))
	// Set all options
	s.apply(opts...)
	go s.listen()
	return s
}

// SetRrProducerCorrelation sets new correleationID creation function
func SetRrProducerCorrelation(corr RrProducerCorrelation) func(*RrProducer) {
	return func(s *RrProducer) {
		s.corr = corr
	}
}

// apply calls all functions to setup options
func (s *RrProducer) apply(opts ...func(*RrProducer)) *RrProducer {
	for _, fn := range opts {
		fn(s)
	}
	return s
}

func (s *RrProducer) add(id string, c chan *Envelope) {
	s.Lock()
	defer s.Unlock()
	s.s[id] = c
}

func (s *RrProducer) get(id string) (chan *Envelope, bool) {
	s.Lock()
	defer s.Unlock()
	c, ok := s.s[id]
	if ok {
		delete(s.s, id)
	}
	return c, ok
}

func (s *RrProducer) timeout(id string) {
	s.Lock()
	defer s.Unlock()
	if _, found := s.s[id]; found {
		s.s[id] = nil
	}
}

func (s *RrProducer) pub(topic string) *Producer {
	s.Lock()
	defer s.Unlock()
	if p, ok := s.producers[topic]; ok {
		return p
	}
	p := Pub(topic)
	s.producers[topic] = p
	return p
}

// ReqRsp send request and wait for response
// topic - nsq topic on wich to send request
// typ   - message type for envelope
// req   - body of the request
// rsp   - stucture to unpuck response into
// sig   - timout singal to signal stop waiting for response
// ttl   - time to live of message for envelope
func (s *RrProducer) ReqRsp(topic, typ string, req interface{}, rsp interface{}, sig chan struct{}, ttl time.Duration) error {
	if ttl < 0 {
		return ErrTimeout
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}
	correlationId := s.corr.NewCorrelationID(topic, typ, req)
	eReq := &Envelope{
		Type:          typ,
		ReplyTo:       s.topic,
		CorrelationId: correlationId,
		Body:          buf,
	}
	if ttl > 0 {
		eReq.ExpiresAt = time.Now().Add(ttl).Unix()
	}
	c := make(chan *Envelope)
	s.add(correlationId, c)

	p := s.pub(topic)
	if err := p.Publish(eReq.Bytes()); err != nil {
		return err
	}

	select {
	case re := <-c:
		if rsp != nil {
			if err := json.Unmarshal(re.Body, rsp); err != nil {
				return err
			}
		}
	case <-sig:
		s.timeout(correlationId)
		return ErrTimeout
	}
	return nil
}

// NewCorrelationID creates unique correlationID as request identifier
func (s *RrProducer) NewCorrelationID(topic, typ string, req interface{}) string {
	s.Lock()
	defer s.Unlock()
	s.msgNo++
	return strconv.Itoa(s.msgNo)
}

// Pub send request without waiting for response.
// topic - nsq topic on wich to send request
// typ   - message type for envelope
// req   - body of the request
func (s *RrProducer) Pub(topic, typ string, req interface{}) error {
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}
	eReq := &Envelope{
		Type: typ,
		Body: buf,
	}
	p := s.pub(topic)
	if err := p.Publish(eReq.Bytes()); err != nil {
		return err
	}
	return nil
}

func (s *RrProducer) listen() {
	handler := func(m *Message) error {
		e, err := NewEnvelope(m.Body)
		if err != nil {
			log.Error(err)
			return err
		}
		if s, found := s.get(e.CorrelationId); found {
			// when s == nil, means that request timed out, nobody is waiting for response
			// nothing to do in that case
			if s != nil {
				s <- e
			}
			return nil
		}
		log.S("id", e.CorrelationId).Info("subscriber not found")
		return nil
	}
	s.sub = Sub(s.topic, handler)
}

// Close implements gracefully stop.
func (s *RrProducer) Close() {
	if s.sub != nil {
		s.sub.Close()
	}
	for _, p := range s.producers {
		p.Close()
	}
}
