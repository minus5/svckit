package nsq

import (
	"fmt"
	"time"

	"github.com/mnu5/svckit/dcy"
	"github.com/mnu5/svckit/log"

	gonsq "github.com/nsqio/go-nsq"
)

type Consumer struct {
	nsqConsumer *gonsq.Consumer
	logger      func() *log.Agregator
	lookups     dcy.Addresses
}

type nsqHandler struct {
	fn func(*Message) error
}

func (h *nsqHandler) HandleMessage(m *gonsq.Message) error {
	// javi periodicki nsqd-u da je procesiranje jos u tijeku
	stop := every(DefaultMsgTouchInterval, m.Touch)
	defer close(stop)
	// zovi handler
	return h.fn(newMessage(m))
}

func MustNewConsumer(topic string, handler func(*Message) error,
	opts ...func(*options)) *Consumer {
	c, err := NewConsumer(topic, handler, opts...)
	if err != nil {
		log.S("topic", topic).Fatal(err)
	}

	return c
}

func NewConsumer(topic string, handler func(*Message) error,
	opts ...func(*options)) (*Consumer, error) {

	o := getDefaults().clone()
	o.apply(opts...)

	cfg := gonsq.NewConfig()
	cfg.MaxInFlight = o.maxInFlight
	cfg.LookupdPollInterval = 10 * time.Second

	c, err := gonsq.NewConsumer(topic, o.channel, cfg)
	if err != nil {
		return nil, err
	}

	c.SetLogger(o.logger, o.logLevel)
	c.AddConcurrentHandlers(&nsqHandler{fn: handler}, o.concurrency)

	err = c.ConnectToNSQLookupds(o.lookupds.String())
	if err != nil {
		return nil, err
	}

	co := &Consumer{
		lookups:     o.lookupds,
		nsqConsumer: c,
		logger: func() *log.Agregator {
			return logger().S("topic", topic).S("channel", o.channel)
		},
	}

	co.logger().I("maxInFlight", o.maxInFlight).I("concurrency", o.concurrency).Debug("starting consumer")
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
	dcy.Unsubscribe(LookupdHTTPServiceName, c.onLookupChanges)
	c.nsqConsumer.Stop()
	<-c.nsqConsumer.StopChan
}

// StartClosing will initiate a graceful stop of the Consumer (permanent)
// Receive on returned chan to block until this process completes
func (c *Consumer) StartClosing() chan int {
	dcy.Unsubscribe(LookupdHTTPServiceName, c.onLookupChanges)
	c.nsqConsumer.Stop()
	return c.nsqConsumer.StopChan
}
