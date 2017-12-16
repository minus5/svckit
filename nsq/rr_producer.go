package nsq

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
)

var (
	DefaultTimeout = time.Hour
	ErrTimeout     = errors.New("timeout")
	ErrStopped     = errors.New("stopped")
	rrProducers    = make(map[string]*RrProducer)
)

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
	if topic == "" {
		topic = fmt.Sprintf("z...rsp.%s.%s", env.AppName(), dcy.NodeName())
	}
	if s, ok := rrProducers[topic]; ok {
		return s
	}
	s := &RrProducer{
		msgNo:     rand.Intn(math.MaxInt32),
		s:         make(map[string]chan *Envelope),
		producers: make(map[string]*Producer),
		topic:     topic,
	}
	rrProducers[topic] = s
	// Set default calc of correlation id-a
	s.apply(SetRrProducerCorrelation(s))
	// Set all options
	s.apply(opts...)
	go s.listen()
	return s
}

func defaultErrorParser(s string) error {
	if s == "" {
		return nil
	}
	return fmt.Errorf(s)
}

// SetRrProducerCorrelation sets new correleationID creation function
func SetRrProducerCorrelation(corr RrProducerCorrelation) func(*RrProducer) {
	return func(s *RrProducer) {
		s.corr = corr
	}
}

// apply calls all functions to setup options
func (s *RrProducer) apply(opts ...func(*RrProducer)) {
	for _, fn := range opts {
		fn(s)
	}
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

func typeToString(i interface{}) string {
	typ := reflect.TypeOf(i).String()
	if strings.HasPrefix(typ, "*") {
		typ = typ[1:]
	}
	return typ
}

// omogucuje mapiranje errora u aplikacijski specificne
// da ne moram imati referencu na ovaj paket
type ErrorsMapping struct {
	Parser     func(string) error
	ErrStopped error
	ErrTimeout error
	ErrFatal   error
}

// ReqRsp send request and wait for response
// topic - nsq topic on wich to send request
// typ   - message type for envelope
// req   - body of the request
// rsp   - stucture to unpuck response into
// sig   - timout singal to signal stop waiting for response
// ttl   - time to live of message for envelope
// errorParser - aplikacijsko specifican parser stringa greske u type
func (s *RrProducer) ReqRsp(topic, typ string, req interface{}, rsp interface{}, sig chan struct{}, ttl time.Duration, em *ErrorsMapping) error {
	if em == nil {
		em = &ErrorsMapping{
			Parser:     defaultErrorParser,
			ErrStopped: ErrStopped,
			ErrTimeout: ErrTimeout,
		}
	}
	if typ == "" {
		typ = typeToString(req)
	}
	if ttl < 0 {
		return em.ErrTimeout
	}
	if ttl == 0 {
		ttl = DefaultTimeout
	}
	buf, err := json.Marshal(req)
	if err != nil {
		if em.ErrFatal != nil {
			log.Error(err)
			return em.ErrFatal
		}
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
		if em.ErrFatal != nil {
			log.Error(err)
			return em.ErrFatal
		}
		return err
	}

	select {
	case re := <-c:
		if rsp != nil && len(re.Body) > 0 {
			if err := json.Unmarshal(re.Body, rsp); err != nil {
				if em.ErrFatal != nil {
					log.Error(err)
					return em.ErrFatal
				}
				return err
			}
		}
		return em.Parser(re.Error)
	case <-time.After(ttl):
		s.timeout(correlationId)
		return em.ErrTimeout
	case <-sig:
		s.timeout(correlationId)
		return em.ErrStopped
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

type RrClient struct {
	pub     *RrProducer
	topic   string
	nameFor func(interface{}) string
	sig     chan struct{}
	ttl     time.Duration
	em      *ErrorsMapping
}

func NewRrClient(topic string,
	nameFor func(interface{}) string,
	sig chan struct{},
	ttl time.Duration,
	em *ErrorsMapping) *RrClient {
	return &RrClient{
		pub:     RrPub(""),
		topic:   topic,
		nameFor: nameFor,
		sig:     sig,
		ttl:     ttl,
		em:      em,
	}
}

func (c *RrClient) Call(req, rsp interface{}) error {
	return c.pub.ReqRsp(c.topic,
		c.nameFor(req), req, rsp, c.sig, c.ttl, c.em,
	)
}

func (c *RrClient) Close() {
	c.pub.Close()
}
