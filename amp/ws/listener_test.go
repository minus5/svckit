package ws

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

const (
	uaMacChrome = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0.0.0 Safari/537.36"
	uaLinux     = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0.0.0 Safari/537.36"
	uaIPhone    = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"

	extPMD    = "permessage-deflate; client_max_window_bits"
	extWebKit = "x-webkit-deflate-frame"
)

// handshake runs one upgrade over an in-memory pipe. headers are written in
// the order given, because the upgrader sees them in arrival order.
func handshake(t *testing.T, l *listener, uri string, headers [][2]string) (cc connCap, status int, extensions string) {
	t.Helper()
	srv, cli := net.Pipe()
	defer func() { _ = srv.Close() }()
	defer func() { _ = cli.Close() }()
	_ = cli.SetDeadline(time.Now().Add(5 * time.Second))

	type result struct {
		cc  connCap
		err error
	}
	done := make(chan result, 1)
	go func() {
		cc, err := l.upgrade(srv)
		done <- result{cc, err}
	}()

	var req strings.Builder
	req.WriteString("GET " + uri + " HTTP/1.1\r\n")
	req.WriteString("Host: localhost\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n")
	req.WriteString("Sec-WebSocket-Version: 13\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n")
	for _, h := range headers {
		req.WriteString(h[0] + ": " + h[1] + "\r\n")
	}
	req.WriteString("\r\n")
	if _, err := cli.Write([]byte(req.String())); err != nil {
		t.Fatalf("write request: %v", err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(cli), nil)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	_ = resp.Body.Close()
	r := <-done
	if r.err != nil {
		t.Fatalf("upgrade: %v", r.err)
	}
	return r.cc, resp.StatusCode, resp.Header.Get("Sec-WebSocket-Extensions")
}

func TestUpgradeDeflateNegotiation(t *testing.T) {
	ua := func(v string) [2]string { return [2]string{"User-Agent", v} }
	ext := func(v string) [2]string { return [2]string{"Sec-WebSocket-Extensions", v} }

	tests := []struct {
		name     string
		opts     []ListenOption
		uri      string
		headers  [][2]string
		wantDefl bool
		wantExt  string // expected prefix of the response extension header; "" means none
	}{
		{"linux negotiates deflate", nil, "/", [][2]string{ua(uaLinux), ext(extPMD)}, true, "permessage-deflate"},
		{"mac is refused by default", nil, "/", [][2]string{ua(uaMacChrome), ext(extPMD)}, false, ""},
		{"iphone webkit frame is refused by default", nil, "/", [][2]string{ua(uaIPhone), ext(extWebKit)}, false, ""},
		{
			// extensions answered before the UA rule runs; harmless, deflate is never used
			"mac with extensions before user-agent: accepted in handshake, never used",
			nil, "/", [][2]string{ext(extPMD), ua(uaMacChrome)}, false, "permessage-deflate",
		},
		{"opt-in: mac negotiates deflate", []ListenOption{DeflateIgnoringUserAgent()}, "/",
			[][2]string{ua(uaMacChrome), ext(extPMD)}, true, "permessage-deflate"},
		{"opt-in: header order does not matter", []ListenOption{DeflateIgnoringUserAgent()}, "/",
			[][2]string{ext(extPMD), ua(uaMacChrome)}, true, "permessage-deflate"},
		{"opt-in: iphone webkit frame negotiates", []ListenOption{DeflateIgnoringUserAgent()}, "/",
			[][2]string{ua(uaIPhone), ext(extWebKit)}, true, "x-webkit-deflate-frame"},
		{"opt-in: kladomat still skips deflate", []ListenOption{DeflateIgnoringUserAgent()}, "/?klad=1",
			[][2]string{ua(uaLinux), ext(extPMD)}, false, ""},
		{"no extension requested", []ListenOption{DeflateIgnoringUserAgent()}, "/",
			[][2]string{ua(uaMacChrome)}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newListener(nil, nil, tt.opts...)
			cc, status, got := handshake(t, l, tt.uri, tt.headers)
			if status != http.StatusSwitchingProtocols {
				t.Fatalf("status = %d, want 101", status)
			}
			if cc.deflateSupported != tt.wantDefl {
				t.Errorf("deflateSupported = %v, want %v", cc.deflateSupported, tt.wantDefl)
			}
			if tt.wantExt == "" && got != "" || !strings.HasPrefix(got, tt.wantExt) {
				t.Errorf("response extensions = %q, want prefix %q", got, tt.wantExt)
			}
		})
	}
}
