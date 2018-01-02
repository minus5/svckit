package nsq

import (
	"context"
	"sync"

	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/pkg/errors"
)

type rpcClient struct {
	pub   *RrProducer
	topic string
}

func RpcClient(topic string) *rpcClient {
	return &rpcClient{
		pub:   RrPub(""),
		topic: topic,
	}
}

func (t *rpcClient) Call(ctx context.Context, typ string, req []byte) ([]byte, string, error) {
	return t.pub.Rpc(ctx, t.topic, typ, req)
}

func (t *rpcClient) Close() {
	t.pub.Close()
}

type Closer interface {
	Close()
}

type server interface {
	Serve(ctx context.Context, typ string, req []byte) ([]byte, error)
}

func RpcServer(ctx context.Context, topic string, srv server) Closer {
	producers := make(map[string]*Producer)
	var l sync.Mutex
	pub := func(topic string) *Producer {
		l.Lock()
		defer l.Unlock()
		if p, ok := producers[topic]; ok {
			return p
		}
		p := Pub(topic)
		producers[topic] = p
		return p
	}

	nsqHandler := func(m *Message) error {
		// raspakiraj poruku u envelope
		eReq, err := NewEnvelope(m.Body)
		if err != nil {
			return errors.Wrap(err, "envelope unpack failed")
		}
		// provjeri da li je expired
		if eReq.Expired() {
			log.S("type", eReq.Type).S("correlationId", eReq.CorrelationId).Info("expired")
			return nil
		}
		// radi request
		rsp, handlerErr := srv.Serve(ctx, eReq.Type, eReq.Body)
		// ako je timeout ili cancel
		if ctx.Err() != nil {
			m.RequeueWithoutBackoff(RequeueDelay)
			return nil
		}
		// treba li odgovoriti
		if eReq.ReplyTo == "" {
			return nil
		}
		// zapakuj
		eRsp, err := eReq.Reply(rsp, handlerErr)
		if err != nil {
			return errors.Wrap(err, "envelope packing failed")
		}
		// posalji odgovor
		if err := pub(eReq.ReplyTo).Publish(eRsp.Bytes()); err != nil {
			return errors.Wrap(err, "nsq publish failed")
		}
		return nil
	}
	return Sub(topic, nsqHandler, Channel(env.AppName()))
}
