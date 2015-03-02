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
	if f.nsqdAddress != "" {
		return nil, consumer.ConnectToNSQD(f.nsqdAddress)
	}
	return consumer, consumer.ConnectToNSQLookupds(f.nsqLookupdHttpAddress)
}

func (f *Connection) NewProducer() (*nsq.Producer, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer(f.nsqdAddress, config)
	if err != nil {
		return nil, err
	}
	return p, nil
}
