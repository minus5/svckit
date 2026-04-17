package nsqx

import (
	"fmt"

	"github.com/minus5/go-nsqx"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"
)

type Consumer struct {
	nsqConsumer *nsqx.Consumer
	logger      func() *log.Agregator
	lookups     dcy.Addresses
}

type nsqHandler struct {
	fn func(*Message) error
}

func (h *nsqHandler) HandleMessage(m *nsqx.Message) error {
	// javi periodicki nsqdu da je procesiranje jos u tijeku
	stop := every(DefaultMsgTouchInterval, m.Touch)
	defer close(stop)
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

	cfg := o.toNsqxConfig()

	c, err := nsqx.NewConsumer(topic, o.channel, cfg)
	if err != nil {
		return nil, err
	}

	c.SetLogger(o.logger, o.logLevel)
	c.AddConcurrentHandlers(&nsqHandler{fn: handler}, o.Concurrency())

	// Use shared Discovery so all consumers in the process share one lookupd poller
	disc := getDiscovery(o)
	c.SetDiscovery(disc)

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

	co.logger().I("maxInFlight", o.maxInFlight).I("concurrency", o.Concurrency()).Info("starting consumer")
	dcy.Subscribe(LookupdHTTPServiceName, co.onLookupChanges)
	dcy.SubscribeByTag(LookupdHTTPServiceNameByTag, LookupdHTTPServiceTag, co.onLookupChanges)
	return co, nil
}

func (c *Consumer) onLookupChanges(as dcy.Addresses) {
	// deduplicate addresses, dcy.Services and dcy.ServicesByTag can return same IP addr,
	// tcp-nsqlookupd and tcp.nsqlookupd
	seen := make(map[string]struct{})
	var unique []string
	for _, a := range as {
		s := a.String()
		if _, dup := seen[s]; !dup {
			seen[s] = struct{}{}
			unique = append(unique, s)
		}
	}

	// Update shared Discovery address list, it will use
	// these new addresses on the next cache miss/poll cycle
	disc := getDiscovery(getDefaults())
	disc.SetAddrs(unique)
	c.lookups = as
	c.logger().S("lookupds", fmt.Sprintf("%v", unique)).Info("lookupds update")
}

func (c *Consumer) Close() {
	dcy.Unsubscribe(LookupdHTTPServiceName, c.onLookupChanges)
	dcy.UnsubscribeByTag(LookupdHTTPServiceNameByTag, LookupdHTTPServiceTag, c.onLookupChanges)
	c.nsqConsumer.Stop()
}

// StartClosing will initiate a graceful stop of the Consumer (permanent),
// receive on returned chan to block until this process completes
func (c *Consumer) StartClosing() chan int {
	dcy.Unsubscribe(LookupdHTTPServiceName, c.onLookupChanges)
	dcy.UnsubscribeByTag(LookupdHTTPServiceNameByTag, LookupdHTTPServiceTag, c.onLookupChanges)
	c.nsqConsumer.Stop()
	// go-nsqx Stop() is synchronous — blocks until fully stopped.
	// Return a closed channel to maintain the same API shape.
	ch := make(chan int)
	close(ch)
	return ch
}
