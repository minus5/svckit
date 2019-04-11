package broker

import "github.com/minus5/svckit/amp"

type appendCache struct {
	msgs  []*amp.Msg
	depth int
}

func newAppendCache() *appendCache {
	return &appendCache{
		depth: 64,
	}
}

// TODO cold start problem:
/*
mogu zamisliti da subscriber dodje nakon restarta
nakon sto je dosao jedan msg, prije nego smo dobili current iz backenda
i onda taj subsriber dobije samo jednu poruku, a ne dobije svih (depth)
*/

func (c *appendCache) Add(m *amp.Msg) {
	if m.IsReplay() && len(c.msgs) > 0 && c.msgs[len(c.msgs)-1].Ts == m.Ts {
		return
	}
	c.msgs = append(c.msgs, m)
	ln := len(c.msgs)
	if ln > 1 {
		if c.msgs[ln-2].Ts >= m.Ts {
			c.msgs = sortMsgs(c.msgs)
		}
	}
	if m.CacheDepth > 0 {
		c.depth = m.CacheDepth
	}
	if ln > c.depth {
		// shrink to depth
		c.msgs = c.msgs[ln-c.depth:]
	}
}

func (c *appendCache) Find(ts int64) []*amp.Msg {
	if len(c.msgs) > 0 && ts >= c.msgs[0].Ts && ts <= c.msgs[len(c.msgs)-1].Ts {
		return c.msgsAfter(ts)
	}
	return c.Current()
}

func (c *appendCache) msgsAfter(ts int64) []*amp.Msg {
	var d []*amp.Msg
	for _, m := range c.msgs {
		if m.Ts > ts {
			d = append(d, m)
		}
	}
	return d
}

func (c *appendCache) Current() []*amp.Msg {
	return c.msgs
}

func (c *appendCache) FindFor(consumerTs int64, m *amp.Msg) uint8 {
	if consumerTs == tsNone {
		return sendCurrent
	}
	if consumerTs == m.Ts {
		return sendNothing
	}
	if m.IsReplay() && consumerTs >= m.Ts { // nemoj ponavljati replay poruke onima koji ih vec imaju
		return sendNothing
	}
	return sendMsg
}
