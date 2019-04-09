package statsd

/*
import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/minus5/svckit/env"
)

// usage struct for counting statistic usage
type usage struct {
	c           client        // client for stats sending
	interval    time.Duration // interval to report usage
	stop        chan struct{} // stop sending usage
	prefix      string        // default stat prefix, appended to client prefix default statsd_sent.appname+instancename
	counters    int64         // number od sent counters in period
	timers      int64         // number of sent timers in period
	gauges      int64         // number of sent gauges in period
	packetsLost int64         // number of client packets lost
}

// usageReport starts reporting usage of statistics
func usageReport(interval time.Duration, client client) *usage {
	u := &usage{
		c:        client,
		interval: interval,
		stop:     make(chan struct{}, 1),
		prefix:   fmt.Sprintf("statsd_sent.%s.%s.", env.AppName(), env.InstanceId()),
	}
	//go u.run()
	return u
}

// // run usage reporting
// func (u *usage) run() {
// 	ri := time.NewTicker(u.interval)
// 	for {
// 		select {
// 		case <-ri.C:
// 			u.report()
// 		case <-u.stop:
// 			ri.Stop()
// 			u.report()
// 			return
// 		}
// 	}
// }

// // report sends report to statsd client
// func (u *usage) report() {
// 	pl := atomic.SwapInt64(&u.packetsLost, u.c.GetLostPackets()-u.packetsLost)
// 	u.c.Gauge(u.prefix+"lost_packets", pl)
// 	u.c.Gauge(u.prefix+"lost_bytes", u.c.GetLostBytes())

// 	u.c.Gauge(u.prefix+"sent_packets", u.c.GetSentPackets())
// 	u.c.Gauge(u.prefix+"sent_bytes", u.c.GetSentBytes())

// 	u.c.Gauge(u.prefix+"send_queue_len", u.c.GetSendQueueLen())
// 	u.c.Gauge(u.prefix+"buf_pool_len", u.c.GetBufPoolLen())

// 	u.c.Gauge(u.prefix+"counters", atomic.SwapInt64(&u.counters, 0))
// 	u.c.Gauge(u.prefix+"timers", atomic.SwapInt64(&u.timers, 0))
// 	u.c.Gauge(u.prefix+"gauges", atomic.SwapInt64(&u.gauges, 0))

// }

// Counter sent
func (u *usage) Counter() {
	atomic.AddInt64(&u.counters, 1)
}

// Timer sent
func (u *usage) Timer() {
	atomic.AddInt64(&u.timers, 1)
}

// Gauge sent
func (u *usage) Gauge() {
	atomic.AddInt64(&u.gauges, 1)
}

// Close stop sending usage reports
func (u *usage) Close() {
	close(u.stop)
}
*/
