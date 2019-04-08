// Package ws implements websocket client connection interface.
package ws

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
	"github.com/mnu5/svckit/log"
	"github.com/pkg/errors"
)

type listener struct {
	ln        net.Listener
	onNewConn func(*Conn)
	conns     sync.WaitGroup
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
	l.conns.Wait()
}

func (l *listener) loop() {
	for {
		tc, err := l.ln.Accept() // cekamo na novu tcp konekciju (tc)
		if err != nil {
			//log.Debug("listener end")
			return
		}
		go l.onConn(tc)
	}
}

func (l *listener) onConn(tc net.Conn) {
	setDeadline(tc)
	cc, err := l.upgrade(tc) // upgrade tcp konkekcije na websocket
	if err != nil {
		_ = tc.Close()
		return
	}
	l.conns.Add(1)
	defer l.conns.Done()
	c := newConn(tc, cc)
	l.onNewConn(c)
	c.wait() // ovdje blocka do prekida komunikacije
}

func (l *listener) upgrade(tc net.Conn) (connCap, error) {
	cc := connCap{}

	ug := ws.Upgrader{
		// podrzava li klijent websocket permessage-deflate
		ExtensionCustom: func(f []byte, os []httphead.Option) ([]httphead.Option, bool) {
			os = make([]httphead.Option, 0)
			field := string(f)
			if strings.Contains(field, "permessage-deflate") {
				params := map[string]string{
					"client_no_context_takeover": "",
					"server_no_context_takeover": "",
				}
				os = append(os, httphead.NewOption("permessage-deflate", params))
				cc.deflateSupported = true
			}
			// iPhone (WebKit) salje po starom standardu
			if strings.Contains(field, "x-webkit-deflate-frame") {
				params := map[string]string{
					"no_context_takeover": "",
				}
				os = append(os, httphead.NewOption("x-webkit-deflate-frame", params))
				cc.deflateSupported = true
			}
			return os, true
		},
		// ocitvamo ostale bitne http headere
		OnHeader: func(k, v []byte) error {
			key := strings.ToLower(string(k))
			value := string(v)
			switch key {
			case "user-agent":
				cc.userAgent = value
			case "x-forwarded-for":
				if cc.forwardedFor != "" {
					cc.forwardedFor += " "
				}
				cc.forwardedFor += value
			}
			return nil
		},
	}

	_, err := ug.Upgrade(tc)
	return cc, err
}
