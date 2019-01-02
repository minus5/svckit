package nsq

import (
	"time"

	"github.com/mnu5/svckit/example/nsq_rr/api"
	"github.com/mnu5/svckit/nsq"
)

var topic = "nsq_rr.req"

func NewClient() *api.Client {
	rr := nsq.NewRrClient(topic,
		api.NameFor,
		nil,
		5*time.Second,
		&nsq.ErrorsMapping{
			Parser:     api.ParseError,
			ErrStopped: api.ErrTransport,
			ErrTimeout: api.ErrTransport,
			ErrFatal:   api.ErrTransport,
		})
	c := api.NewClient(rr)
	return c
}

type Closer interface {
	Close()
}

func NewServer(svc api.Service) Closer {
	srv := api.NewServer(svc)
	rr := nsq.NewRrServer(topic, srv, api.TypeFor, nil)
	return rr
}
