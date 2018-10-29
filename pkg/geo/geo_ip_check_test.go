package geo

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	TestGeoIPFile = "testGeoIP.dat"
)

func TestGeoIPCheck(t *testing.T) {
	g, err := NewIPCheck(TestGeoIPFile, []string{})
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

func TestGeoIPOk(t *testing.T) {
	assert.True(t, IpOk("bilo sta"))
	Init(TestGeoIPFile)
	assert.False(t, IpOk("208.117.229.99"))
	assert.True(t, IpOk("212.15.168.195"))
	assert.True(t, IpOk("127.0.0.1"))
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

func TestGetGeoIPFile(t *testing.T) {
	fi, err := os.Stat(TestGeoIPFile + ".gz")
	assert.Nil(t, err)
	http.HandleFunc("/geoip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", fi.ModTime().Format(time.RFC1123))
		http.ServeFile(w, r, TestGeoIPFile+".gz")
	})
	go http.ListenAndServe(":12534", nil)
	assert.Nil(t, getGeoIPFile("http://localhost:12534/geoip", TestGeoIPFile))
}

func TestCheckGeoIPFile(t *testing.T) {
	_, err := checkGeoIPFile(TestGeoIPFile)
	assert.Nil(t, err)
}

func TestGeoIPWhitelist(t *testing.T) {
	g, err := NewIPCheck(TestGeoIPFile, []string{})
	assert.Nil(t, err)

	gWhitelist, err := NewIPCheck(TestGeoIPFile, []string{"2.2.2.2/32", "1.1.1.0/24"})
	assert.Nil(t, err)

	assert.False(t, g.Check("1.1.1.1"))
	assert.False(t, g.Check("1.1.1.25"))

	assert.True(t, gWhitelist.Check("1.1.1.1"))
	assert.True(t, gWhitelist.Check("1.1.1.25"))

	assert.True(t, gWhitelist.Check("2.2.2.2"))
	assert.False(t, gWhitelist.Check("3.2.2.2"))
	assert.False(t, gWhitelist.Check("2.3.2.2"))
	assert.False(t, gWhitelist.Check("2.2.3.2"))
	assert.False(t, gWhitelist.Check("2.2.2.3"))

}
