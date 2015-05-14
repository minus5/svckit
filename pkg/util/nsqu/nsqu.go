package nsqu

import (
	"fmt"
	"log"
	"pkg/util"
	"strings"

	"github.com/bitly/go-nsq"
)

type Connection struct {
	nsqLookupdHttpAddress []string
	nsqdAddress           string
}

func NewDefaultPortConnection() *Connection {
	return NewConnection(nil, "")
}

func NewConnection(nsqLookupdHttpAddress []string, nsqdAddress string) *Connection {
	if nsqLookupdHttpAddress == nil && nsqdAddress == "" {
		nsqdAddress = "localhost:4150"
	}
	return &Connection{
		nsqLookupdHttpAddress: nsqLookupdHttpAddress,
		nsqdAddress:           nsqdAddress,
	}
}

type MessageHandler func([]byte) error

func (f *Connection) MustNewConsumer(topic, channel string, concurency int, handler MessageHandler) *nsq.Consumer {
	c, err := f.NewConsumer(topic, channel, concurency, handler)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func (f *Connection) NewConsumer(topic, channel string, concurency int, handler MessageHandler) (*nsq.Consumer, error) {
	//replace %s in channel with hostname
	if strings.Contains(channel, "%") {
		channel = fmt.Sprintf(channel, util.Hostname())
	}
	nsqCfg := nsq.NewConfig()
	nsqCfg.Set("max_in_flight", concurency)
	consumer, err := nsq.NewConsumer(topic, channel, nsqCfg)
	if err != nil {
		return nil, err
	}
	log.Printf("nsq consumer for topic %s channel %s concurency: %d", topic, channel, concurency)
	consumer.AddConcurrentHandlers(nsq.HandlerFunc(func(m *nsq.Message) error {
		return handler(m.Body)
	}), concurency)
	if len(f.nsqLookupdHttpAddress) == 0 {
		return nil, consumer.ConnectToNSQD(f.nsqdAddress)
	}
	return consumer, consumer.ConnectToNSQLookupds(f.nsqLookupdHttpAddress)
}

//Consume - zakaci za na topic na channel
// sve sto dobije salje u handler
// rjesava clean stop tako da ceka da se closa exitChan
// opcionalno zove register/unregisterProcess na pocetku i na kraju
func (f *Connection) Consume(topic, channel string, concurency int, handler MessageHandler,
	exitChan chan struct{}, registerProcess, unregisterProcess func()) {
	consumer, err := f.NewConsumer(topic, channel, concurency, func(buf []byte) error {
		return handler(buf)
	})
	if err != nil {
		log.Fatal(err)
	}

	if registerProcess != nil {
		registerProcess()
	}
	<-exitChan
	if consumer != nil {
		consumer.Stop()
		<-consumer.StopChan
		log.Printf("nsqConsumer zaustavio %s", topic)
	}
	if unregisterProcess != nil {
		unregisterProcess()
	}
}

func (f *Connection) NewProducer() (*nsq.Producer, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer(f.nsqdAddress, config)
	if err != nil {
		return nil, err
	}
	return p, nil
}
