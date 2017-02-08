// +build nsq_rr_pub

package main

import (
	"time"

	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/nsq"
	"github.com/minus5/svckit/signal"
)

func main() {
	log.Printf("starting")
	closeSig := signal.Term()
	pub := nsq.RrPub("nsq_rr.rsp")

	no := 1
	for {
		select {
		case <-time.Tick(2 * time.Second):
			no++
			go func(no int) {
				// simulate timeout, after 7 seconds
				sig := make(chan struct{})
				go func() {
					time.Sleep(7 * time.Second)
					close(sig)
				}()
				req := Msg{Id: no}
				rsp := &Msg{}
				// make request and wait for response
				err := pub.ReqRsp("nsq_rr.req", "req", req, rsp, sig, 0)
				if err != nil {
					log.I("req", no).I("rsp", rsp.Id).Error(err)
				} else {
					log.I("req", no).I("rsp", rsp.Id).Info("reponse")
				}
			}(no)
		case <-closeSig:
			log.Debug("stopping")
			pub.Close()
			log.Debug("stopped")
			return
		}
	}

}

type Msg struct {
	Id int
}
