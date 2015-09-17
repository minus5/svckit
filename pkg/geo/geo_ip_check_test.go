package geo

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	TEST_GEO_IP_FILE = "testGeoIP.dat"
)

func TestGeoIpCheck(t *testing.T) {
	//	file := os.Getenv("GEOIP_DB_PATH")
	//	if file == "" {
	//		t.Skip("skipping geo ip test, set GEOIP_DB_PATH env var to run this test")
	//	}
	g, err := NewIpCheck(TEST_GEO_IP_FILE)
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
	Init(TEST_GEO_IP_FILE)
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

func TestGetGeoIpFile(t *testing.T) {
	fi, err := os.Stat(TEST_GEO_IP_FILE + ".gz")
	assert.Nil(t, err)
	http.HandleFunc("/geoip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", fi.ModTime().Format(time.RFC1123))
		http.ServeFile(w, r, TEST_GEO_IP_FILE+".gz")
	})
	go http.ListenAndServe(":12534", nil)
	assert.Nil(t, getGeoIpFile("http://localhost:12534/geoip", TEST_GEO_IP_FILE))
}

func TestCheckGeoIpFile(t *testing.T) {
	_, err := checkGeoIpFile(TEST_GEO_IP_FILE)
	assert.Nil(t, err)
}
