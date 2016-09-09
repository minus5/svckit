package httpu

import (
	"io"
	"net/http"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"
	"strings"
	"time"
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
