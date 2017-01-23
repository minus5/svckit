package nsq

import (
	"fmt"
	"sync"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"

	gonsq "github.com/nsqio/go-nsq"
)

type Consumer struct {
	nsqConsumer *gonsq.Consumer
	onceClose   sync.Once
	logger      func() *log.Agregator
	lookups     dcy.Addresses
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
	cfg.LookupdPollInterval = 2 * time.Second

	c, err := gonsq.NewConsumer(topic, defaults.channel, cfg)
	if err != nil {
		return nil, err
	}

	c.SetLogger(defaults.logger, defaults.logLevel)
	c.AddHandler(&nsqHandler{fn: handler})

	err = c.ConnectToNSQLookupds(defaults.lookupds.String())
	if err != nil {
		return nil, err
	}

	co := &Consumer{
		lookups:     defaults.lookupds,
		nsqConsumer: c,
		logger: func() *log.Agregator {
			return logger().S("topic", topic).S("channel", defaults.channel)
		},
	}

	co.logger().I("maxInFlight", defaults.maxInFlight).Info("starting consumer")
	dcy.Subscribe(LookupdHTTPServiceName, co.onLookupChanges)
	return co, nil
}

func (c *Consumer) onLookupChanges(as dcy.Addresses) {
	for _, a := range as {
		if err := c.nsqConsumer.ConnectToNSQLookupd(a.String()); err != nil {
			logger().Error(err)
		}
	}
	for _, a := range c.lookups {
		if !as.Contains(a) {
			if err := c.nsqConsumer.DisconnectFromNSQLookupd(a.String()); err != nil {
				logger().Error(err)
			}
		}
	}
	c.lookups = as
	c.logger().S("lookupds", fmt.Sprintf("%v", as)).Debug("lookupds update")
}

func (c *Consumer) Close() {
	c.onceClose.Do(func() {
		dcy.Unsubscribe(LookupdHTTPServiceName, c.onLookupChanges)
		c.nsqConsumer.Stop()
		<-c.nsqConsumer.StopChan
	})
}
