// Package merger sluzi obradi poruka full/diff tipa.
// Tip poruka koje dolaze zi tecajne servise.
// Brine se da poruke iz njega izadju u pravom redoslijedu.
// Uvijek izlazi diff pa full istog no-a.
// Iz izvora ocekujemo samo stream diff-ova.
// Full-ovi se proizvode merganjem (odatle i naziv paketa) diff-ova.
// Inicijalni full se dobija na zahtjev (Dopuna metoda).
package merger

// Router prima poruke za razlicite kanale i prosljedjuje ih
// fullDiffOrderer-ima na obradu.
type Router struct {
	fdos          map[string]*fullDiffOrderer
	in            chan *msg
	Output        chan *OutMsg
	dopunaHandler func(string, string)
}

// New ulazna tocak u paket.
// Kreira novi router.
func New(dopunaHandler func(string, string)) *Router {
	r := &Router{
		fdos:          make(map[string]*fullDiffOrderer),
		in:            make(chan *msg),
		Output:        make(chan *OutMsg, 1024),
		dopunaHandler: dopunaHandler,
	}
	go r.loop()
	return r
}

func (r *Router) handler(m *msg) {
	//fmt.Printf("Router handler: Type=%s, No=%d, channel=%s\n", m.typ, m.no,m.channel)
	body := m.JsonBody()
	r.Output <- &OutMsg{
		Type:     m.typ,
		No:       m.no,
		jsonBody: body,
		body:     m.body,
	}
}

func (r *Router) loop() {
	for m := range r.in {
		channel := m.channel
		fdo, ok := r.fdos[channel]
		if m.isDel {
			if ok {
				fdo.close()
				delete(r.fdos, channel)
				//log.Printf("[DEBUG] remove fdo %s", channel)
			}
			r.handler(m)
			continue
		}
		if !ok {
			typ := m.typ
			d := func() {
				if r.dopunaHandler != nil {
					r.dopunaHandler(typ, channel)
				}
			}
			fdo = newFullDiffOrderer(d)
			if limit, ok := oooLimits[channel]; ok {
				fdo.oooLimit = limit //ako imamo custom limit za ovaj kanal
			}
			r.fdos[channel] = fdo
			go func() {
				for m := range fdo.out {
					r.handler(m)
					mtrc.Counter("out")
				}
			}()
		}
		fdo.in <- m
		mtrc.Counter("in")
	}
	for _, fdo := range r.fdos {
		fdo.close()
	}
	close(r.Output)
}

// Add dodaje novu poruku u router.
func (r *Router) Add(typ string, no int64, body []byte, isDel bool) {
	r.in <- newMsg(typ, no, body, isDel)
}

// Close cleanup routera.
func (r *Router) Close() {
	close(r.in)
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
