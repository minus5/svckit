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

var ErrTimeout = errors.New("timeout")

// RrProducer request response producer.
// Implemets reuest response communication over nsq.
type RrProducer struct {
	s         map[string]chan *Envelope
	producers map[string]*Producer
	topic     string
	sub       *Consumer
	msgNo     int
	sync.Mutex
}

// RrPub creates new RrProducer.
// Topic is nsq topic where remote services send responses.
func RrPub(topic string) *RrProducer {
	s := &RrProducer{
		msgNo:     rand.Intn(math.MaxInt32),
		s:         make(map[string]chan *Envelope),
		producers: make(map[string]*Producer),
		topic:     topic,
	}
	go s.listen()
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
	correlationId := s.correlationId()
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

// creates unique request identifier
func (s *RrProducer) correlationId() string {
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
			// ako je s == nil to znaci da se dogodio timeout pa vise nitko ne ceka na response
			// u tom slucaju nista ne radim
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
