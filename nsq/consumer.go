package nsq

import (
	"github.com/minus5/svckit/log"
	"sync"

	gonsq "github.com/nsqio/go-nsq"
)

type Consumer struct {
	nsqConsumer *gonsq.Consumer
	onceClose   sync.Once
}

type nsqHandler struct {
	fn func(*Message) error
}

func (h *nsqHandler) HandleMessage(m *gonsq.Message) error {
	return h.fn(newMessage(m))
}

func MustNewConsumer(topic string, handler func(*Message) error,
	opts ...func(*options)) *Consumer {
	c, err := NewConsumer(topic, handler, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func NewConsumer(topic string, handler func(*Message) error,
	opts ...func(*options)) (*Consumer, error) {
	Set(opts...)

	cfg := gonsq.NewConfig()
	cfg.MaxInFlight = defaults.maxInFlight

	c, err := gonsq.NewConsumer(topic, defaults.channel, cfg)
	if err != nil {
		return nil, err
	}

	c.SetLogger(defaults.logger, defaults.logLevel)
	c.AddHandler(&nsqHandler{fn: handler})

	err = c.ConnectToNSQLookupds(defaults.lookupdHTTPAddrs)
	if err != nil {
		return nil, err
	}

	log.S("lib", "svckit.nsq").S("topic", topic).S("channel", defaults.channel).I("maxInFlight", defaults.maxInFlight).Info("starting consumer")
	return &Consumer{nsqConsumer: c}, nil
}

func (c *Consumer) Close() {
	c.onceClose.Do(func() {
		c.nsqConsumer.Stop()
		<-c.nsqConsumer.StopChan
	})
}
