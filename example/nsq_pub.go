// +build nsq_pub

package main

import (
	"fmt"
	"github.com/mnu5/svckit/log"
	"github.com/mnu5/svckit/nsq"
	"github.com/mnu5/svckit/signal"
	"time"
)

func main() {
	log.Printf("starting")
	close := signal.Term()
	pub := nsq.Pub("nsq_example")

	no := 1
	for {
		select {
		case <-time.Tick(time.Second):
			msg := fmt.Sprintf("%5d %s", no, time.Now())
			log.S("msg", msg).Info("pub")
			pub.MustPublish([]byte(msg))
			no++
		case <-close:
			log.Debug("stopping")
			pub.Close()
			log.Debug("stopped")
			return
		}
	}

}
