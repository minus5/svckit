package httpu

import (
	"net/http"
	"testing"
	"time"

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

// TestNewRequestDcy must be run with the following command: SVCKIT_DCY_CONSUL=- bash -c 'go test -v --run=TestNewRequestDcy'
// SVCKIT_DCY_CONSUL=- env variable will trigger 'noConsulTestMode' in the init() function of dcy package,
// while calling go test with bash -c ensures that both env variable setup and test will be run in the same shell process.
func TestNewRequestDcy(t *testing.T) {
	t.Skip("requires dcy setup via SVCKIT_DCY_CONSUL=- env variable")
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

func TestNoneMatch(t *testing.T) {
	r, err := http.NewRequest("GET", "http://localhost/", nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, NoneMatch(r, "perozdero42"))
	assert.True(t, NoneMatch(r, "perozdero108"))
	assert.True(t, NoneMatch(r, ""))
	r.Header.Set("If-None-Match", "perozdero42")
	assert.True(t, NoneMatch(r, "perozdero108"))
	assert.False(t, NoneMatch(r, "perozdero42"))
	assert.True(t, NoneMatch(r, ""))
}

func TestModifiedSince(t *testing.T) {
	r, err := http.NewRequest("GET", "http://localhost/", nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	lastModified := time.Date(1977, 6, 21, 12, 34, 56, 0, time.UTC)
	assert.True(t, ModifiedSince(r, lastModified))
	r.Header.Set("If-Modified-Since", "garblegarble")
	assert.True(t, ModifiedSince(r, lastModified))
	r.Header.Set("If-Modified-Since", "Tue, 21 Jun 1977 12:34:55 UTC")
	assert.False(t, ModifiedSince(r, lastModified))
	r.Header.Set("If-Modified-Since", "Tue, 21 Jun 1977 12:34:57 UTC")
	assert.True(t, ModifiedSince(r, lastModified))
}
