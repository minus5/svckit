// Package ws implements websocket client connection interface.
package ws

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
	"github.com/minus5/svckit/log"
	"github.com/pkg/errors"
)

type listener struct {
	ln        net.Listener
	onNewConn func(*Conn)
}

// Open opens new tcp port.
// Returns net.Listener for call to the Listen method below.
// Fails if port is already open.
func Open(port int) (net.Listener, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	log.I("port", port).Info("ws listener started")
	return ln, nil
}

// MustOpen raises fatal if unsuccessful
func MustOpen(port int) net.Listener {
	ln, err := Open(port)
	if err != nil {
		log.Fatal(err)
	}
	return ln
}

// Listen starts listening for new connections, blocks until ctx closed.
// Then stops listening for new connections, and waits for current to finish.
func Listen(ctx context.Context, ln net.Listener, h func(*Conn)) {
	l := &listener{
		ln:        ln,
		onNewConn: h,
	}
	go func() {
		<-ctx.Done()
		_ = l.ln.Close()
	}()
	l.loop()
}

func (l *listener) loop() {
	var wg sync.WaitGroup
	for {
		tc, err := l.ln.Accept() // cekamo na novu tcp konekciju (tc)
		if err != nil {
			//log.Debug("listener end")
			return
		}
		wg.Add(1)
		go func() {
			l.onConn(tc)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (l *listener) onConn(tc net.Conn) {
	setDeadline(tc)
	cc, err := l.upgrade(tc) // upgrade tcp konkekcije na websocket
	if err != nil {
		_ = tc.Close()
		return
	}
	c := newConn(tc, cc)
	l.onNewConn(c) // ovdje blocka do prekida komunikacije
}

func (l *listener) upgrade(tc net.Conn) (connCap, error) {
	cc := connCap{}

	ug := ws.Upgrader{
		// podrzava li klijent websocket permessage-deflate
		ExtensionCustom: func(f []byte, os []httphead.Option) ([]httphead.Option, bool) {
			os = make([]httphead.Option, 0)
			field := string(f)
			// skip deflating for kladomat, implementation in Chromium is buggy, constantly reconnects
			if cc.meta["klad"] != "" {
				return os, true
			}
			if strings.Contains(field, "permessage-deflate") && !cc.deflateSupported {
				params := map[string]string{
					"client_no_context_takeover": "",
					"server_no_context_takeover": "",
				}
				os = append(os, httphead.NewOption("permessage-deflate", params))
				cc.deflateSupported = true
			}
			// iPhone (WebKit) salje po starom standardu
			if strings.Contains(field, "x-webkit-deflate-frame") && !cc.deflateSupported {
				params := map[string]string{
					"no_context_takeover": "",
				}
				os = append(os, httphead.NewOption("x-webkit-deflate-frame", params))
				cc.deflateSupported = true
			}
			return os, true
		},
		OnRequest: func(uri []byte) error {
			cc.meta = parseQueryString(uri)
			for k, v := range cc.meta {
				log.S("key", k).S("value", v).Debug("client meta")
			}
			return nil
		},
		// ocitvamo ostale bitne http headere
		OnHeader: func(k, v []byte) error {
			key := strings.ToLower(string(k))
			value := string(v)
			switch key {
			case "user-agent":
				cc.userAgent = value
				// na iOS 15, 16, ... ne radi vise web kompresija
				if strings.Contains(value, "OS 1") || strings.Contains(value, "Mac OS X 10_1") {
					cc.deflateSupported = false
				}
			case "x-forwarded-for":
				if cc.forwardedFor != "" {
					cc.forwardedFor += " "
				}
				cc.forwardedFor += value
			// case "cookie":
			// 	cc.cookies = parseCookies(value)
			// 	for k, v := range cc.cookies {
			// 		log.S("key", k).S("value", v).Debug("cookie")
			// 	}
			default:
				log.S("key", key).S("value", value).Debug("header")
			}
			return nil
		},
	}

	_, err := ug.Upgrade(tc)
	return cc, err
}

func parseQueryString(uri []byte) map[string]string {
	u, err := url.Parse(string(uri))
	if err != nil {
		return nil
	}
	qs := make(map[string]string)
	for k, v := range u.Query() {
		qs[k] = strings.Join(v, ",")
	}
	return qs
}

func parseCookies(rawCookies string) map[string]string {
	if rawCookies == "" {
		return nil
	}
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	request := http.Request{
		Header: header,
	}
	cookies := make(map[string]string)
	for _, c := range request.Cookies() {
		cookies[c.Name] = c.Value
	}
	return cookies
}
