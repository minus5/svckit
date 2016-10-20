// Package merger sluzi obradi poruka full/diff tipa.
// Tip poruka koje dolaze zi tecajne servise.
// Brine se da poruke iz njega izadju u pravom redoslijedu.
// Uvijek izlazi diff pa full istog no-a.
// Iz izvora ocekujemo samo stream diff-ova.
// Full-ovi se proizvode merganjem (odatle i naziv paketa) diff-ova.
// Inicijalni full se dobija na zahtjev (Dopuna metoda).
package merger

import (
	"encoding/json"
	"pkg/msgs"
	"strings"
	"time"

	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
	"github.com/minus5/svckit/nsq"
)

// Router prima poruke za razlicite kanale i prosljedjuje ih
// fullDiffOrderer-ima na obradu.
type Router struct {
	fdos    map[string]*fullDiffOrderer
	handler func(*msgs.Backend)
	in      chan *msg
}

var pub = nsq.MustNewProducer("")

func dopunaTopicFromTyp(typ string) string {
	if strings.HasPrefix(typ, "vfl/") || strings.HasPrefix(typ, "vbl/") || strings.HasPrefix(typ, "vto/") {
		return "vsport.req"
	}
	return "tecajna.req"
}

// Dopuna salje poruku za dopunu live kanala.
func Dopuna(typ, channel string) {
	topic := dopunaTopicFromTyp(typ)
	log.S("typ", typ).S("topic", topic).S("channel", channel).Debug("merger/dopuna")
	msg := struct {
		Channel string `json:"channel"`
	}{Channel: channel}
	buf, _ := json.Marshal(msg)
	bBuf := msgs.CreateBackend("dopuna", 0, buf)
	if err := pub.PublishTo(topic, bBuf); err != nil {
		log.Error(err)
	}
	metric.Counter("merger.dopuna")
}

// New ulazna tocak u paket.
// Kreira novi router.
func New(handler func(*msgs.Backend)) *Router {
	r := &Router{
		fdos:    make(map[string]*fullDiffOrderer),
		handler: handler,
		in:      make(chan *msg),
	}
	go r.loop()
	go r.metrics()
	return r
}

func (r *Router) loop() {
	for m := range r.in {
		if m == nil {
			break
		}
		channel := m.channel
		fdo, ok := r.fdos[channel]
		if m.backend.IsDel {
			if ok {
				fdo.close()
				delete(r.fdos, channel)
				//log.Printf("[DEBUG] remove fdo %s", channel)
			}
			r.handler(m.backend)
			continue
		}
		if !ok {
			d := func() {
				Dopuna(m.typ, channel)
			}
			fdo = newFullDiffOrderer(d)
			r.fdos[channel] = fdo
			go func() {
				for m := range fdo.out {
					//log.Printf("[DEBUG] out %s %d", m.typ, m.no)
					r.handler(m.backend)
					metric.Counter("merger.out")
				}
			}()
			//log.Printf("[DEBUG] add fdo %s", channel)
		}
		//log.Printf("[DEBUG] in  %s %d", m.typ, m.no)
		fdo.in <- m
		metric.Counter("merger.in")
	}
	for _, fdo := range r.fdos {
		fdo.close()
	}
}

// Add dodaje novu poruku u router.
func (r *Router) Add(m *msgs.Backend) {
	if !m.IsFullDiff() {
		r.handler(m)
		return
	}
	fm := newMsg(m)
	r.in <- fm
}

// Close cleanup routera.
func (r *Router) Close() {
	r.in <- nil
}

// Size - broj kanala.
func (r *Router) Size() int {
	return len(r.fdos)
}

// QueueSize - ukupan broj poruka u queue-ima svih kanala.
func (r *Router) QueueSize() int {
	queueSize := 0
	for _, fdo := range r.fdos {
		queueSize += fdo.queueSize()
	}
	return queueSize
}

func (r *Router) metrics() {
	for {
		time.Sleep(10 * time.Second)
		metric.Gauge("merger.queueSize", r.QueueSize())
		metric.Gauge("merger.fdos", r.Size())
	}
}
