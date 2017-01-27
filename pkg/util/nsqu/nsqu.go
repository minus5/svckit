package nsqu

import (
	"encoding/json"
	"errors"
	"pkg/util"
	"sync"
	"time"

	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/nsq"
)

var ErrTimeout = errors.New("timeout")

type Sub struct {
	topic string
	sub   *nsq.Consumer
}

// NewSub ...
// handler u njega ulazi tip poruke i funkcija koja parsa body
//         ocekuje se da vrati reponse (nil ako nema odgovora) ili eventualni error
func NewSub(topic string, handler func(string, func(interface{}) error) (interface{}, error)) *Sub {
	h := func(m *nsq.Message) error {
		eReq, err := NewEnvelope(m.Body)
		if err != nil {
			return err
		}
		if eReq.Expired() {
			log.S("type", eReq.Type).S("correlationId", eReq.CorrelationId).Info("expired")
			return nil
		}
		rsp, err := handler(eReq.Type, func(o interface{}) error {
			return eReq.ParseBody(o)
		})
		if err != nil {
			return err
		}
		if eReq.ReplyTo == "" {
			return nil
		}
		eRsp, err := eReq.Reply(rsp)
		if err != nil {
			return err
		}
		pub := nsq.Pub(eReq.ReplyTo)
		if err := pub.Publish(eRsp.Bytes()); err != nil {
			return err
		}
		return nil
	}
	return &Sub{
		topic: topic,
		sub:   nsq.Sub(topic, h),
	}
}

func (s *Sub) Close() {
	if s.sub == nil {
		return
	}
	s.sub.Close()
}

type Nsqu struct {
	s     map[string]chan *Envelope
	p     map[string]*nsq.Producer
	topic string
	sub   *nsq.Consumer
	sync.Mutex
}

func (s *Nsqu) add(id string, c chan *Envelope) {
	s.Lock()
	defer s.Unlock()
	s.s[id] = c
}

func (s *Nsqu) get(id string) (chan *Envelope, bool) {
	s.Lock()
	defer s.Unlock()
	c, ok := s.s[id]
	if ok {
		delete(s.s, id)
	}
	return c, ok
}

func (s *Nsqu) timeout(id string) {
	s.Lock()
	defer s.Unlock()
	if _, found := s.s[id]; found {
		s.s[id] = nil
	}
}

func (s *Nsqu) pub(topic string) *nsq.Producer {
	s.Lock()
	defer s.Unlock()
	if p, ok := s.p[topic]; ok {
		return p
	}
	p := nsq.Pub(topic)
	s.p[topic] = p
	return p
}

// ReqRsp send request and wait for response
func (s *Nsqu) ReqRsp(topic, typ string, req interface{}, rsp interface{}, sig chan struct{}, ttl time.Duration) error {
	if ttl <= 0 {
		return ErrTimeout
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}
	cid := util.Uuid()
	eReq := &Envelope{
		Type:          typ,
		ReplyTo:       s.topic,
		CorrelationId: cid,
		Body:          buf,
	}
	if ttl > 0 {
		eReq.ExpiresAt = time.Now().Add(ttl).Unix()
	}
	p := s.pub(topic)
	if err := p.Publish(eReq.Bytes()); err != nil {
		return err
	}
	c := make(chan *Envelope)
	s.add(cid, c)

	select {
	case re := <-c:
		if err := json.Unmarshal(re.Body, rsp); err != nil {
			return err
		}
	case <-sig:
		s.timeout(cid)
		return ErrTimeout
	}
	return nil
}

// Req send request without vaiting for response
func (s *Nsqu) Req(topic, typ string, req interface{}) error {
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

func New(topic string) *Nsqu {
	s := &Nsqu{
		s:     make(map[string]chan *Envelope),
		p:     make(map[string]*nsq.Producer),
		topic: topic,
	}
	go s.listen()
	return s
}

func (s *Nsqu) listen() {
	handler := func(m *nsq.Message) error {
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
	s.sub = nsq.Sub(s.topic, handler)
}

func (s *Nsqu) Close() {
	if s.sub != nil {
		s.sub.Close()
	}
	for _, v := range s.p {
		v.Close()
	}
}
