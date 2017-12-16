package dcy

import (
	"fmt"
	"log"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testServiceAddresses = []string{
		"127.0.0.1:12345",
		"127.0.0.1:12348",
	}
)

func TestServiceName(t *testing.T) {
	s, d := serviceName("test.service.sd", "sd")
	assert.Equal(t, "test", s)
	assert.Equal(t, "", d)
	s, d = serviceName("test.service.s2.sd", "sd")
	assert.Equal(t, "test", s)
	assert.Equal(t, "s2", d)
	s, d = serviceName("test", "sd")
	assert.Equal(t, "test", s)
	assert.Equal(t, "", d)
}

func TestConsulSelf(t *testing.T) {
	assert.Equal(t, dc, "dev")
	assert.Equal(t, domain, "sd")
	assert.Equal(t, nodeName, "node01")
}

func TestServices(t *testing.T) {
	srvs, err := Services("test3.service.sd")
	assert.Nil(t, err)
	assert.Len(t, srvs, 2)

	srvs, err = Services("test1")
	assert.Nil(t, err)
	assert.Contains(t, testServiceAddresses, srvs[0].String())
	assert.Contains(t, testServiceAddresses, srvs[1].String())
	assert.Len(t, srvs, 2)
}

func TestURL(t *testing.T) {
	assert.Equal(t, "udp://127.0.0.1:9514/pero", URL("udp://syslog/pero"))
	assert.Equal(t, "http://google.com", URL("http://google.com"))

	urls := []string{
		"http://127.0.0.1:12345/pero",
		"http://127.0.0.1:12348/pero",
	}
	url0 := 0
	for i := 0; i < 10; i++ {
		u := URL("http://test1.service.sd/pero")
		assert.Contains(t, urls, u)
		if u == urls[0] {
			url0++
		}
	}
	t.Logf("url0: %d", url0)
	assert.True(t, url0 > 0 && url0 < 10)
}

func TestUnpackURLPackURL(t *testing.T) {
	data := []struct {
		url                      string
		scheme, host, port, path string
		query                    url.Values
	}{
		{"http://consul.service.sd/pero/zdero", "http", "consul.service.sd", "", "/pero/zdero", url.Values{}},
		// ovdje je bitno da ide prvo gljiva pa zdero, jer url.Values.Encode() vraca sortirano po keyu, pa kad bi gljiva bilo nakon zdero, mi expectamo isti string koji smo ubacili i dobili bi drukcije
		{"http://consul.service.sd/pero?gljiva=22&zdero=42%2F13&zdero=3.141592", "http", "consul.service.sd", "", "/pero", url.Values{"zdero": []string{"42/13", "3.141592"}, "gljiva": []string{"22"}}},
		{"http://consul.service.sd:123/pero/zdero", "http", "consul.service.sd", "123", "/pero/zdero", url.Values{}},
		{"wss://host:123/api", "wss", "host", "123", "/api", url.Values{}},
		{"wss://host:123", "wss", "host", "123", "", url.Values{}},
		{"wss://host", "wss", "host", "", "", url.Values{}},
		{"host", "", "host", "", "", nil},
		{"host:123", "", "host", "123", "", nil},
	}
	for _, d := range data {
		scheme, host, port, path, query := unpackURL(d.url)
		assert.Equal(t, d.scheme, scheme)
		assert.Equal(t, d.host, host)
		assert.Equal(t, d.port, port)
		assert.Equal(t, d.path, path)
		assert.Equal(t, d.query, query)
		assert.Equal(t, d.url, packURL(scheme, host, port, path, query))
	}
}

func TestShouldDiscoverHost(t *testing.T) {
	domain = "sd"
	assert.True(t, shouldDiscoverHost("host"))
	assert.True(t, shouldDiscoverHost("host.sd"))
	assert.False(t, shouldDiscoverHost("google.com"))
}

func TestCentrala(t *testing.T) {
	t.Skip("pokretao na produkciji")
	s, err := Service("centrala.service.sd")
	assert.Nil(t, err)
	fmt.Printf("%v\n", s)

	u := URL("http://centrala.service.s2.sd/HqStatusInfo")
	fmt.Printf("url: %v\n", u)

	u = URL("http://tecajna/ping")
	fmt.Printf("url: %v\n", u)
}

func TestNovi(t *testing.T) {
	t.Skip("pokretao na produkciji")
	u := URL("http://tecajna-beta.service.sd/tecajevi")
	log.Print(shouldDiscoverHost("tecajna-beta.service.sd"))
	log.Printf("%s", u)

	u = URL("http://tecajna-beta.service.sd/tecajevi")
	log.Print(shouldDiscoverHost("tecajna-beta.service.sd"))
	log.Printf("%s", u)
}

func TestMongo(t *testing.T) {
	c, err := MongoConnStr()
	assert.Nil(t, err)
	assert.Equal(t, "127.0.0.1:27017,192.168.10.123:27017", c)

	c, err = MongoConnStr("mongo")
	assert.Nil(t, err)
	assert.Equal(t, "127.0.0.1:27017,192.168.10.123:27017", c)
}

func TestSubscribe(t *testing.T) {
	assert.Len(t, subscribers, 0)
	h1 := func(Addresses) {}
	h2 := func(Addresses) {}
	Subscribe("svc", h1)
	assert.Len(t, subscribers, 1)
	assert.Len(t, subscribers["svc"], 1)
	Subscribe("svc", h2)
	assert.Len(t, subscribers, 1)
	assert.Len(t, subscribers["svc"], 2)

	Unsubscribe("svc", h1)
	assert.Len(t, subscribers, 1)
	assert.Len(t, subscribers["svc"], 1)

}
