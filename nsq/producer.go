package nsq

import (
	"github.com/mnu5/svckit/log"

	gonsq "github.com/nsqio/go-nsq"
)

type Producer struct {
	topic       string
	nsqProducer *gonsq.Producer
}

func MustNewProducer(topic string, opts ...func(*options)) *Producer {
	p, err := NewProducer(topic, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return p
}

func NewProducer(topic string, opts ...func(*options)) (*Producer, error) {
	Set(opts...)
	cfg := gonsq.NewConfig()
	p, err := gonsq.NewProducer(defaults.nsqdTCPAddr, cfg)
	if err != nil {
		return nil, err
	}
	p.SetLogger(defaults.logger, defaults.logLevel)
	return &Producer{nsqProducer: p, topic: topic}, nil
}

func (p *Producer) Close() {
	p.nsqProducer.Stop()
}

func (p *Producer) Publish(msg []byte) error {
	return p.nsqProducer.Publish(p.topic, msg)
}

func (p *Producer) PublishTo(topic string, msg []byte) error {
	return p.nsqProducer.Publish(topic, msg)
}

func (p *Producer) MustPublish(msg []byte) {
	if err := p.Publish(msg); err != nil {
		log.Fatal(err)
	}
}
