package jsonreq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWithVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versionIn := r.Header.Get(versionHeaderKey)
		assert.Equal(t, "123", versionIn)
		w.Header().Set(versionHeaderKey, "456")
	}))
	defer ts.Close()

	_, v, err := GetWithVersion(ts.URL, "123")
	assert.Nil(t, err)
	assert.Equal(t, "456", v)

}

func TestCalcRetryInterval(t *testing.T) {
	r := New("")
	r.retrySleep = 1000
	assert.Equal(t, 1, r.calcRetryInterval(0))
	assert.Equal(t, 2, r.calcRetryInterval(1))
	assert.Equal(t, 7, r.calcRetryInterval(2))
	assert.Equal(t, 20, r.calcRetryInterval(3))
	assert.Equal(t, 54, r.calcRetryInterval(4))
	assert.Equal(t, 148, r.calcRetryInterval(5))
	assert.Equal(t, 403, r.calcRetryInterval(6))
	assert.Equal(t, 1000, r.calcRetryInterval(7))
}

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	}))
	defer ts.Close()

	j := New(ts.URL, Retries(2, 1), Header("key", "value"))
	assert.Equal(t, 2, j.retries)
	assert.Equal(t, 1, j.retrySleep)
	assert.NotNil(t, j.headers)

	buf, err := j.Get()
	assert.Nil(t, err)
	assert.Equal(t, "hello\n", string(buf))
	assert.Equal(t, http.StatusOK, j.StatusCode())
}
