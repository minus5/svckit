package broker

import (
	"testing"

	"github.com/minus5/svckit/amp"
	"github.com/stretchr/testify/assert"
)

func TestShrink(t *testing.T) {
	c := newAppendCache()
	c.depth = 2
	c.Add(&amp.Msg{Ts: 10, UpdateType: amp.Append})
	assert.Len(t, c.msgs, 1)
	c.Add(&amp.Msg{Ts: 11, UpdateType: amp.Append})
	assert.Len(t, c.msgs, 2)
	c.Add(&amp.Msg{Ts: 12, UpdateType: amp.Append})
	assert.Len(t, c.msgs, 2)
	assert.Equal(t, int64(11), c.msgs[0].Ts)
	assert.Equal(t, int64(12), c.msgs[1].Ts)

	c.Add(&amp.Msg{Ts: 13, UpdateType: amp.Append})
	assert.Len(t, c.msgs, 2)
	assert.Equal(t, int64(12), c.msgs[0].Ts)
	assert.Equal(t, int64(13), c.msgs[1].Ts)

	c.Add(&amp.Msg{Ts: 14, UpdateType: amp.Append, CacheDepth: 3})
	assert.Equal(t, c.depth, 3)
	assert.Len(t, c.msgs, 3)
}
