// +build nsq_rr_sub

package main

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/mnu5/svckit/log"
	"github.com/mnu5/svckit/nsq"
	"github.com/mnu5/svckit/signal"
)

func main() {
	sub := nsq.RrSub("nsq_rr.req", handler)
	//clean exit
	signal.WaitForInterupt()
	log.Debug("stopping")
	sub.Close()
	log.Debug("stopped")
}

func handler(typ string, body []byte) (interface{}, error) {
	switch typ {
	case "req":
		req := &Msg{}
		if err := json.Unmarshal(body, req); err != nil {
			log.Error(err)
			return nil, err
		}
		// random delay
		n := rand.Intn(8)
		time.Sleep(time.Duration(n) * time.Second)
		log.I("req", req.Id).I("delay", n).Info("req")
		req.Id = -req.Id
		return req, nil
	default:
		log.S("type", typ).Notice("unknown type")
	}
	return nil, nil
}

type Msg struct {
	Id int
}
