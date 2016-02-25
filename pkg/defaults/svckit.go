package defaults

import (
	"fmt"
	"log/syslog"
	"os"
	"pkg/svckit/leader"
	"pkg/svckit/log"
	"pkg/svckit/metric"
	"pkg/svckit/nsq"
	"pkg/svckit/sd"
)

const (
	statsdDns      = "statsd.service.sd"
	lookupdHTTPDns = "nsqlookupd-http.service.sd"
)

func init() {
	// za docker containere preusmjeravam na $node_ip (ip adresa hosta)
	if e, ok := os.LookupEnv(EnvNodeIp); ok && e != "" {
		nodeIp := e
		log2Syslog(fmt.Sprintf("%s:514", nodeIp))
		sd.SetDns(fmt.Sprintf("%s:8600", nodeIp))
		leader.SetConsulHttpAddr(fmt.Sprintf("%s:8500", nodeIp))
		nsq.Set(nsq.NsqdTCPAddr(fmt.Sprintf("%s:4150", nodeIp)))
	}

	node := Hostname()
	if e, ok := os.LookupEnv(EnvNode); ok && e != "" {
		node = e
	}
	initMetric(node)

	nsq.Set(nsq.LookupdHTTPAddr(lookupdHTTPDns),
		nsq.Channel(fmt.Sprintf("%s-%s", AppName(), node)))
}

func initMetric(node string) {
	prefix := fmt.Sprintf("%s.%s", AppName(), node)
	metric.Connect(statsdDns, prefix)
}

func log2Syslog(addr string) {
	sys, err := syslog.Dial("udp", addr, syslog.LOG_LOCAL5, AppName())
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(sys)
}
