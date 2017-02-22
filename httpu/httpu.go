package httpu

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"
)

const (
	StatusTooManyRequests = 429
)

func StartHttp(listen string) {
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Make a Ping to a service
func Ping(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := NewRequest("GET", url, nil)
	if err != nil {
		log.S("url", url).ErrorS("ping failed")
		return false
	}
	rsp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return false
	}
	return true
}

// Find remote_ip in various headers.
func RemoteIp(r *http.Request) string {
	proxyIp := r.Header.Get("X-Forwarded-For")
	if proxyIp == "" {
		return strings.Split(r.RemoteAddr, ":")[0]
	}
	// X-Forwarded-For is potentially a list of addresses separated with ","
	parts := strings.Split(proxyIp, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts[0]
}

// Isto kao NewRequest u http paketu no napravi service discovery preko dcy-ja
func NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, dcy.URL(url), body)
}

// Isto kao Get iz http uz service discovery
func Get(url string) (*http.Response, error) {
	return http.Get(dcy.URL(url))
}

// WsCompressionSupported returns true if the client supports permessage-deflat WebSocket Compression Extension
func WsCompressionSupported(r *http.Request) bool {
	wsExtHeader := r.Header.Get("Sec-WebSocket-Extensions")
	return strings.Contains(wsExtHeader, "permessage-deflate")
}

// NoneMatch returns true if the given etag doesn't match the request If-None-Match header
// If the header or the given etag is empty, it also returns true.
func NoneMatch(r *http.Request, etag string) bool {
	ifNoneMatch := r.Header.Get("If-None-Match")
	if ifNoneMatch == "" || etag == "" {
		return true
	}
	return ifNoneMatch != etag
}

// ModifiedSince returns true if the given lastModified time is before the request If-Modified-Since header.
// If the header is empty or not in RFC1123 format, it also returns true.
func ModifiedSince(r *http.Request, lastModified time.Time) bool {
	ifModifiedSince := r.Header.Get("If-Modified-Since")
	if ifModifiedSince == "" {
		return true
	}
	ifModifiedSinceTime, err := time.Parse(time.RFC1123, ifModifiedSince)
	if err != nil {
		log.Errorf("invalid If-Modified-Since header: %v", err)
		return true
	}
	return lastModified.Before(ifModifiedSinceTime)
}
