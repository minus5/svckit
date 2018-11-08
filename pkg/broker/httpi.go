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
	closing := false // flag da zaatvaramo konekciju
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

	sendChan := make(chan *Message, 1024)
	go func() {
		for m := range sendChan {
			if closing {
				continue
			}
			err := send(m.Event, string(m.Data))
			if extraWork != nil {
				extraWork(m, err)
			}
		}
	}()

	unsubscribe := func() {
		closing = true
		go b.Unsubscribe(msgsCh) // zatvara msgsCh nakon unsubscribe-a
	}

	sendToCh := func(m *Message) {
		select {
		case sendChan <- m:
		default:
			log.S("client_id", clientID).I("send_len", len(sendChan)).S("event", m.Event).J("data", m.Data).ErrorS("unable to send last message")
			unsubscribe()

			if m.Event == "status" && string(m.Data) == "done" {
				unsubscribe()
			}
		}
	}

	for {
		select {
		case <-closeCh:
			log.S("client_id", clientID).Info("Client disconnected")
			unsubscribe()
		case <-closeSignal:
			log.S("client_id", clientID).Info("Server close")
			unsubscribe()
		case m := <-msgsCh:
			if m == nil {
				close(sendChan) //msgsCh closan, nema sto za slati
				return
			}
			sendToCh(m)
		case <-time.After(20 * time.Second):
			sendToCh(NewMessage("heartbeat", []byte(time.Now().Format(time.RFC3339))))
		}
	}
}
