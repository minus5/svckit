package sse

import (
	"fmt"
	"net/http"
)

type Sse struct {
	w      http.ResponseWriter
	r      *http.Request
	f      http.Flusher
	closed chan struct{}
	read   bool
}

func (sse *Sse) Read() ([]byte, error) {
	if !sse.read {
		sse.read = true
		return []byte(`{"t":1,"u":[{"s":"m","n":0}]}`), nil
	}
	<-sse.closed
	return nil, fmt.Errorf("connection closed")
}

func (sse *Sse) Write(payload []byte, deflated bool) error {
	lwritten, err := sse.w.Write(payload)
	if err != nil {
		return err
	}
	if lwritten > 0 {
		sse.f.Flush()
	}
	return nil
}

func (sse *Sse) DeflateSupported() bool     { return false }
func (sse *Sse) Headers() map[string]string { return nil }
func (sse *Sse) No() uint64                 { return 12345 }
func (sse *Sse) Close() error {
	close(sse.closed)
	return nil
}
func (sse *Sse) Meta() map[string]string { return nil }

func New(w http.ResponseWriter, r *http.Request) (*Sse, error) {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming unsupported")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	return &Sse{
		w:      w,
		r:      r,
		f:      f,
		closed: make(chan struct{}),
	}, nil
}
