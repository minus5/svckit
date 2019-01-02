package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mnu5/svckit/amp"
	"github.com/mnu5/svckit/amp/broker"
	"github.com/mnu5/svckit/nsq"
	"github.com/mnu5/svckit/signal"
)

const (
	methodAdd    = "add"
	methodReplay = "replay"
)

type params struct {
	X int64 `json:"x,omitempty"`
	Y int64 `json:"y,omitempty"`
}

func (p *params) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

func (p *params) FromJSON(buf []byte) error {
	return json.Unmarshal(buf, p)
}

func (p *params) FromMsgp(buf []byte) error {
	panic("not supported")
}

type rsp struct {
	Z int64 `json:"z"`
}

func (r *rsp) ToLang(string) amp.BodyMarshaler {
	return r
}

func (r *rsp) ToMsgp() ([]byte, error) {
	panic("not supported")
}

func (r *rsp) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func main() {
	msgs := make(chan *amp.Msg, 1)
	broker := broker.New()
	broker.Consume(msgs)

	pub := nsq.Pub("math.rsp")

	reply := func(m *amp.Msg, r *rsp) error {
		buf := m.Response(r).Marshal()
		return pub.PublishTo(m.ReplyTo, buf)
	}

	nsq.Sub("math.req", func(nm *nsq.Message) error {
		m := amp.Parse(nm.Body)
		switch m.Method {
		case methodAdd:
			p := &params{}
			if err := m.BodyTo(p); err != nil {
				return err
			}
			z := p.X + p.Y
			fmt.Printf("z: %d\n", z)
			if err := reply(m, &rsp{Z: z}); err != nil {
				return err
			}
		case methodReplay:
			for _, m := range broker.Replay("") {
				_ = pub.PublishTo("math.topics", m.Marshal())
			}
		default:
			return fmt.Errorf("unsupported method %s", m.Method)
		}
		return nil
	})

	go func() {

		i := int64(1)
		x := i
		y := x

		publish := func() {
			p := &params{
				X: x,
				Y: y,
			}
			updateType := amp.Diff
			if y != 0 {
				updateType = amp.Full
			}
			m := amp.NewPublish("math.topic.1", time.Now().UnixNano(), updateType, p)
			msgs <- m
			_ = pub.PublishTo("math.topics", m.Marshal())
		}
		publish()

		diff := time.Tick(time.Second)
		full := time.Tick(30 * time.Second)
		for {
			select {
			case <-diff:
				i++
				x = i
				y = 0
				publish()
			case <-full:
				x = i
				y = x
				publish()
			}
		}
	}()

	signal.WaitForInterupt()
}
