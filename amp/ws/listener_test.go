package ws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheck(t *testing.T) {
	// exact match
	err := checkOriginHost("https://www.supersport.hr/sbk/", "www.supersport.hr")
	assert.NoError(t, err)
	// localhost
	err = checkOriginHost("http://localhost:3010", "localhost")
	assert.NoError(t, err)
	// subdomain of same SLD matches
	err = checkOriginHost("https://api.supersport.hr", "www.supersport.hr")
	assert.NoError(t, err)
	// empty origin allowed
	err = checkOriginHost("", "www.supersport.hr")
	assert.NoError(t, err)
	// different SLD
	err = checkOriginHost("http://supersport.si:3010", "localhost")
	assert.Error(t, err)
	err = checkOriginHost("https://forum.hr", "supersport.si")
	assert.Error(t, err)
}
