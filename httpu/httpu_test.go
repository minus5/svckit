package httpu

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteIp(t *testing.T) {
	r, err := http.NewRequest("GET", "http://localhost/", nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	r.RemoteAddr = "2.3.4.5:3000"
	assert.Equal(t, "2.3.4.5", RemoteIp(r))
	r.Header.Add("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	assert.Equal(t, "1.2.3.4", RemoteIp(r))
}

func TestNewRequest(t *testing.T) {
	r, err := NewRequest("GET", "http://localhost/", nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, "http://localhost/", r.URL.String())
}

func TestNewRequestDcy(t *testing.T) {
	r, err := NewRequest("GET", "http://test2.service.sd/foo", nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, "http://10.11.12.13:1415/foo", r.URL.String())
}

func TestPing(t *testing.T) {
	t.Skip("zahtjeva zivi servis")
	assert.True(t, Ping("http://5-web-backend05.supersport.local:8091/ping"))

}
