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
	s               map[string]chan *Envelope
	producers       map[string]*Producer
	topic           string
	sub             *Consumer
	consumerOptions []func(*options)
	msgNo           int
	corr            RrProducerCorrelation
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
		topic = fmt.Sprintf("z...rsp.%s.%s", env.AppName(), env.InstanceId())
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

// RrProducerConsumerOptions sets configuration options for the underlying Consumer of RrProducer.
func RrProducerConsumerOptions(opts ...func(*options)) func(*RrProducer) {
	return func(s *RrProducer) {
		s.consumerOptions = opts
	}
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
// em    - error mapping, mapping to application specific messages
func (s *RrProducer) ReqRsp(topic, typ string, req interface{}, rsp interface{}, sig chan struct{}, ttl time.Duration, em *ErrorsMapping) error {
	if typ == "" {
		typ = typeToString(req)
	}
	p := ReqRspBaseParams{
		Topic: topic,
		Typ:   typ,
		Ttl:   ttl,
		Sig:   sig,
		Em:    em,
	}
	p.defaults()
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return p.Fatal(err)
	}
	p.Req = reqBuf
	p.correlationId = s.corr.NewCorrelationID(topic, typ, req)
	rspBuf, err := s.ReqRspBase(p)
	if err != nil {
		return err
	}
	if rsp != nil && len(rspBuf) > 0 {
		if err := json.Unmarshal(rspBuf, rsp); err != nil {
			return p.Fatal(err)
		}
	}
	return nil
}

// ReqRspBuf sends request and waits for response, passes rsp as []byte
func (s *RrProducer) ReqRspBuf(topic, typ string, req interface{}, sig chan struct{}, ttl time.Duration, em *ErrorsMapping) ([]byte, error) {
	if typ == "" {
		typ = typeToString(req)
	}
	p := ReqRspBaseParams{
		Topic: topic,
		Typ:   typ,
		Ttl:   ttl,
		Sig:   sig,
		Em:    em,
	}
	p.defaults()
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return nil, p.Fatal(err)
	}
	p.Req = reqBuf
	p.correlationId = s.corr.NewCorrelationID(topic, typ, req)
	return s.ReqRspBase(p)
}

type ReqRspBaseParams struct {
	Topic         string
	Typ           string
	Req           []byte
	Ttl           time.Duration
	Sig           chan struct{}
	Em            *ErrorsMapping
	correlationId string
}

func (p *ReqRspBaseParams) defaults() {
	if p.Ttl <= 0 {
		p.Ttl = DefaultTimeout
	}
	if p.Em == nil {
		p.Em = &ErrorsMapping{
			Parser:     defaultErrorParser,
			ErrStopped: ErrStopped,
			ErrTimeout: ErrTimeout,
		}
	}
}

func (p *ReqRspBaseParams) Fatal(err error) error {
	if p.Em.ErrFatal != nil {
		return p.Em.ErrFatal
	}
	return err
}

func (p *ReqRspBaseParams) Timeout() error {
	if p.Em.ErrTimeout != nil {
		return p.Em.ErrTimeout
	}
	return ErrTimeout
}

func (p *ReqRspBaseParams) Stopped() error {
	if p.Em.ErrStopped != nil {
		return p.Em.ErrStopped
	}
	return ErrStopped
}

func (p *ReqRspBaseParams) Error(text string) error {
	if p.Em.Parser != nil {
		return p.Em.Parser(text)
	}
	return defaultErrorParser(text)
}

func (s *RrProducer) ReqRspBase(p ReqRspBaseParams) ([]byte, error) {
	p.defaults()
	if p.correlationId == "" {
		p.correlationId = s.NewCorrelationID("", "", nil)
	}

	eReq := &Envelope{
		Type:          p.Typ,
		ReplyTo:       s.topic,
		CorrelationId: p.correlationId,
		Body:          p.Req,
		ExpiresAt:     time.Now().Add(p.Ttl).Unix(),
	}
	c := make(chan *Envelope)
	s.add(p.correlationId, c)

	if err := s.pub(p.Topic).Publish(eReq.Bytes()); err != nil {
		return nil, p.Fatal(err)
	}

	timer := time.NewTimer(p.Ttl)
	defer timer.Stop()
	select {
	case re := <-c:
		return re.Body, p.Error(re.Error)
	case <-timer.C:
		s.timeout(p.correlationId)
		return nil, p.Timeout()
	case <-p.Sig:
		s.timeout(p.correlationId)
		return nil, p.Stopped()
	}
	return nil, nil
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
	s.sub = Sub(s.topic, handler, s.consumerOptions...)
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
