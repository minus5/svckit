package ws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheck(t *testing.T) {
	err := checkOriginHost("https://www.supersport.hr/sbk/", "www.supersport.hr")
	assert.NoError(t, err)
	err = checkOriginHost("http://localhost:3010", "localhost")
	assert.NoError(t, err)
	err = checkOriginHost("http://supersport.si:3010", "localhost")
	assert.Error(t, err)
	err = checkOriginHost("https://forum.hr", "supersport.si")
	assert.Error(t, err)
}
