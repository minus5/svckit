package nsq

import (
	"errors"
	"time"
)

type RpcTransport struct {
	pub   *RrProducer
	topic string
	ttl   time.Duration
	em    *ErrorsMapping
}

func NewRpcTransport(topic string, ttl time.Duration, em *ErrorsMapping) *RpcTransport {
	return &RpcTransport{
		pub:   RrPub(""),
		topic: topic,
		ttl:   ttl,
		em:    em,
	}
}

func (t *RpcTransport) Call(typ string, req []byte) ([]byte, error) {
	return t.pub.ReqRspBase(ReqRspBaseParams{
		Topic: t.topic,
		Ttl:   t.ttl,
		Typ:   typ,
		Req:   req,
		Em:    t.em,
	})
}

func (t *RpcTransport) Close() {
	t.pub.Close()
}

type rpcHandler func(typ string, body []byte) ([]byte, error)

func RpcServe(topic string, h rpcHandler) *RrConsumer {
	return RrSub(topic,
		func(typ string, body []byte) (interface{}, error) {
			return h(typ, body)
		},
		RequeueError(errors.New("newer")),
	)
}
