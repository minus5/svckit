package geo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoIpCheck(t *testing.T) {
	file := os.Getenv("GEOIP_DB_PATH")
	if file == "" {
		t.Skip("skipping geo ip test, set GEOIP_DB_PATH env var to run this test")
	}
	g, err := NewIpCheck(file)
	assert.Nil(t, err)
	assert.NotNil(t, g)
	assert.False(t, g.Check("208.117.229.99")) //google.com
	assert.True(t, g.Check("212.92.207.181"))  //google.hr
	assert.True(t, g.Check("212.92.207.181"))  //google.hr
	assert.Equal(t, 2, len(g.cache))
	assert.Equal(t, true, g.cache["212.92.207.181"])

	assert.True(t, g.Check("127.0.0.1"))
	assert.True(t, g.Check("10.103.46.121"))
	assert.True(t, g.Check("212.15.168.195"))

	assert.False(t, g.Check("sto je ovo "))
}

func TestGeoIpOk(t *testing.T) {
	assert.True(t, IpOk("bilo sta"))
	Init(os.Getenv("GEOIP_DB_PATH"))
	assert.False(t, IpOk("208.117.229.99"))
	assert.True(t, IpOk("212.15.168.195"))
}

func TestGeoIsLocalAddress(t *testing.T) {
	assert.True(t, isLocalAddress("127.0.0.1"))
	assert.True(t, isLocalAddress("192.168.10.10"))
	assert.True(t, isLocalAddress("10.0.0.1"))
	assert.True(t, isLocalAddress("172.16.0.1"))
	assert.True(t, isLocalAddress("172.31.0.1"))
	assert.False(t, isLocalAddress("172.15.0.1"))
	assert.False(t, isLocalAddress("not an ip"))
	assert.False(t, isLocalAddress("212.92.207.181"))
}
