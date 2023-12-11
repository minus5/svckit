package broker

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/satori/go.uuid"

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

	// dozvoljavamo da client posalje svoj id, sluzi za bug tracking
	clientID := r.FormValue("clientid")
	if "" == clientID {
		clientID = uuid.NewV4().String()
	}

	send := func(event, data string) error {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := make([]byte, 10240)
				stackSize := runtime.Stack(stackTrace, true)
				log.S("client_id", clientID).
					S("panic", fmt.Sprintf("%v", r)).
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

	unsubscribe := func() {
		b.Unsubscribe(msgsCh) // Unsubscribe sa brokera i zatvara channel
	}

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-closeCh:
			log.S("client_id", clientID).Info("Client disconnected")
			unsubscribe()
			return
		case <-closeSignal:
			log.S("client_id", clientID).Info("Server close")
			unsubscribe()
			return
		case m := <-msgsCh:
			if m == nil {
				log.S("client_id", clientID).Info("Broker closed channel")
				return
			}
			err := send(m.Event, string(m.GetData()))
			if extraWork != nil {
				extraWork(m, err)
			}
			if m.Event == "status" && string(m.GetData()) == "done" {
				unsubscribe()
				return
			}
		case <-heartbeat.C:
			m := NewMessage("heartbeat", []byte(time.Now().Format(time.RFC3339)), nil)
			err := send(m.Event, string(m.GetData()))
			if extraWork != nil {
				extraWork(m, err)
			}
		}
	}
}
