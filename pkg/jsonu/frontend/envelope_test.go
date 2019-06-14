package frontend

import (
	"github.com/minus5/svckit/pkg/testu"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeader(t *testing.T) {
	m, err := NewEnvelope([]byte(`i/full:0:0:i_full_2028929.json:1492329991330:2028929
`))
	assert.Nil(t, err)
	testu.PP(m)

	m, err = NewEnvelope([]byte(`pong:24:0::1492330618370:
`))
	assert.Nil(t, err)
	testu.PP(m)
}
