package nsqx

import (
	"github.com/minus5/go-nsqx"
	"github.com/minus5/svckit/log"
)

type Producer struct {
	topic       string
	nsqProducer *nsqx.Producer
}

func MustNewProducer(topic string, opts ...func(*options)) *Producer {
	p, err := NewProducer(topic, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return p
}

func NewProducer(topic string, opts ...func(*options)) (*Producer, error) {
	o := getDefaults().clone()
	o.apply(opts...)

	cfg := o.toNsqxConfig()
	p, err := nsqx.NewProducer(o.nsqdTCPAddr, cfg)
	if err != nil {
		return nil, err
	}
	p.SetLogger(o.logger, o.logLevel)
	// log on start
	logger().S("topic", topic).
		I("ProducerPipelines", o.producerPipelines).
		I("ZeroCopyThrshld", o.zeroCopyThreshold).
		I("BatchMaxSize", o.batchMaxSize).
		I("OutputBuffSize", o.outputBufferSize).
		S("Nsqd", o.nsqdTCPAddr).
		Info("starting producer")
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
