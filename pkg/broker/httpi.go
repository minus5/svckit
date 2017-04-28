package broker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/minus5/svckit/log"
)

// StreamingSSE sse helper
func StreamingSSE(w http.ResponseWriter, r *http.Request, b *Broker, closeSignal <-chan struct{}, extraWork func(*Message, error)) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	closeCh := w.(http.CloseNotifier).CloseNotify()
	//header-i potrebni za sse
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	msgsCh := b.Subscribe()

	send := func(event, data string) error {
		msg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data))
		_, err := w.Write(msg)
		if err != nil {
			return err
		}
		f.Flush()
		return nil
	}

	sendChan := make(chan *Message, 1024)
	go func() {
		for m := range sendChan {
			err := send(m.Event, string(m.Data))
			if extraWork != nil {
				extraWork(m, err)
			}
		}
	}()

	unsubscribe := func() {
		go b.Unsubscribe(msgsCh)
	}
	for {
		select {
		case <-closeCh:
			unsubscribe()
		case m := <-msgsCh:
			if m == nil { //kanal je closan
				close(sendChan)
				return
			}
			select {
			case sendChan <- m:
			default:
				log.Errorf("unable to send message")
			}
			if m.Event == "status" && string(m.Data) == "done" {
				unsubscribe()
			}
		case <-closeSignal:
			unsubscribe()
		case <-time.After(20 * time.Second):
			send("heartbeat", time.Now().Format(time.RFC3339))
		}
	}
}
