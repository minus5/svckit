package nsq

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/mnu5/svckit/amp"
	"github.com/mnu5/svckit/env"
	"github.com/mnu5/svckit/log"
	"github.com/mnu5/svckit/nsq"
	"github.com/pkg/errors"
)

type Requester struct {
	topic           string
	p               *nsq.Producer
	c               *nsq.Consumer
	queue           map[string]*request // requests in process
	correlationNo   int
	topicForMsgType func(string) string
	sync.Mutex
}

type request struct {
	msg    *amp.Msg
	source amp.Subscriber
}

func MustRequester(topicForMsgType func(string) string) *Requester {
	r, err := NewRequester(topicForMsgType)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func NewRequester(topicForMsgType func(string) string) (*Requester, error) {
	p, err := nsq.NewProducer("")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	r := &Requester{
		p:               p,
		queue:           make(map[string]*request),
		topic:           resposesTopicName(),
		topicForMsgType: topicForMsgType,
	}
	c, err := nsq.NewConsumer(r.topic, r.responses)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	r.c = c
	return r, nil
}

func resposesTopicName() string {
	return fmt.Sprintf("z...rsp-%s-%s", env.AppName(), env.NodeName())
}

func (r *Requester) responses(nm *nsq.Message) error {
	m := amp.Parse(nm.Body)
	r.reply(m.CorrelationID, m)
	return nil
}

func (r *Requester) reply(correlationID string, m *amp.Msg) {
	r.Lock()
	req, ok := r.queue[correlationID]
	if ok {
		delete(r.queue, correlationID)
	}
	r.Unlock()

	if !ok {
		return
	}
	m.CorrelationID = req.msg.CorrelationID
	req.source.Send(m)
	return
}

func (r *Requester) Send(e amp.Subscriber, m *amp.Msg) {
	r.Lock()
	r.correlationNo++
	correlationID := strconv.Itoa(r.correlationNo)
	r.queue[correlationID] = &request{msg: m, source: e}
	r.Unlock()

	rm := m.Request()
	rm.CorrelationID = correlationID
	rm.ReplyTo = r.topic
	buf := rm.Marshal()

	go func() {
		err := r.p.PublishTo(r.topicForMsgType(m.Method), buf)
		if err != nil {
			r.reply(correlationID, m.ResponseTransportError())
		}
	}()
}

func (r *Requester) Unsubscribe(e amp.Subscriber) {
	r.Lock()
	defer r.Unlock()
	for key, req := range r.queue {
		if req.source == e {
			delete(r.queue, key)
		}
	}
}

func (r *Requester) Wait(ctx context.Context) {
	<-ctx.Done()
	r.p.Close()
	r.c.Close()
	r.Lock()
	defer r.Unlock()
	r.queue = make(map[string]*request)
}
