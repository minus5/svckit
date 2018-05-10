package broker

import (
	"fmt"
	"net/http"
	"runtime"
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
		defer func() {
			if r := recover(); r != nil {
				stackTrace := make([]byte, 20480)
				stackSize := runtime.Stack(stackTrace, true)
				log.S("panic", fmt.Sprintf("%v", r)).
					I("stack_size", stackSize).
					S("stack_trace", string(stackTrace)).
					ErrorS("recover from panic")
			}
		}()

		msg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data))
		lwritten, err := w.Write(msg)
		if err != nil {
			return err
		}
		if lwritten > 0 {
			f.Flush()
		}
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
		go b.Unsubscribe(msgsCh) // zatvara msgsCh nakon unsubscribe-a
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
				log.S("event", m.Event).J("data", m.Data).ErrorS("unable to send last 1024 messages")
				unsubscribe()
			}
			if m.Event == "status" && string(m.Data) == "done" {
				unsubscribe()
			}
		case <-closeSignal:
			unsubscribe()
		case <-time.After(20 * time.Second):
			send("heartbeat", time.Now().Format(time.RFC3339))
			//log.Info("heartbeat send")
		}
	}
}
