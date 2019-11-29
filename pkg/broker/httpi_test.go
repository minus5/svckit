package broker

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gorilla/mux"
)

var (
	closeHTTP chan struct{}
	wg        sync.WaitGroup
	Client    = &http.Client{}
)

func testSSEHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	b, ok := FindBroker(topic)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	StreamingSSE(w, r, b, closeHTTP, nil)
}

func testHTTPServer(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/{topic}", testSSEHandler)
	srv := &http.Server{
		Addr: "0.0.0.0:6969",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}
	closeHTTP = make(chan struct{})
	go func() {
		wg.Add(1)
		srv.ListenAndServe()
		wg.Done()
	}()
	go func() {
		<-closeHTTP
		srv.Close()
	}()
}

func testSSEReq(t *testing.T, topic string) (chan *Message, chan struct{}, error) {
	req, err := http.NewRequest("GET", "http://localhost:6969/"+topic, nil)
	if err != nil {
		t.Log(err)
		return nil, nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	rsp, err := Client.Do(req)
	if err != nil {
		t.Log(err)
		return nil, nil, err
	}

	msgCh := make(chan *Message)  // kanal za sse poruke
	doneCh := make(chan struct{}) // signalni kanal za preikdanje citanja

	// wait done loop
	go func() {
		<-doneCh
		rsp.Body.Close()
		t.Log("SSE request stop")
	}()

	// read loop
	go func() {
		wg.Add(1)
		defer wg.Done()
		defer close(msgCh)
		var msg *Message
		delim := []byte{':', ' '}
		br := bufio.NewReader(rsp.Body)
		for {
			bs, err := br.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					t.Log(err)
				}
				break
			}
			if len(bs) < 2 {
				continue
			}
			spl := bytes.Split(bs, delim)
			if len(spl) < 2 {
				continue
			}
			switch string(spl[0]) {
			case "event":
				msg = &Message{}
				msg.Event = string(bytes.TrimSpace(spl[1]))
			case "data":
				msg.data = bytes.TrimSpace(spl[1])
				t.Log(msg)
				select {
				case msgCh <- msg:
				case <-doneCh:
					break
				}
			}
		}
		t.Log("SSE read done")
	}()

	return msgCh, doneCh, nil
}

func TestStreamingSSE(t *testing.T) {
	topic := "httpi_test"
	b := GetBufferedBroker(topic)
	assert.NotNil(t, b)
	assert.Len(t, b.subscribers, 0) // nema subascribera

	// Postavi podatke u brokera
	Stream(topic, "event-q", []byte("1"))
	Stream(topic, "event-q", []byte("2"))

	testHTTPServer(t) // sse server (ceka da dobije prvi response)

	// sse klijent
	msgCh, doneCh, err := testSSEReq(t, topic)
	assert.NoError(t, err)
	assert.NotNil(t, msgCh)
	assert.NotNil(t, doneCh)

	var msg *Message
	msg = <-msgCh
	assert.Equal(t, "event-q", msg.Event)
	assert.Equal(t, []byte("1"), msg.GetData())
	assert.Len(t, b.subscribers, 1) // ima 1 subscriber jer smo primili 1 poruku

	msg = <-msgCh
	assert.Equal(t, "event-q", msg.Event)
	assert.Equal(t, []byte("2"), msg.GetData())

	close(doneCh)    // Prestani slusati SSE
	close(closeHTTP) // Stop HTTP SSE server

	wg.Wait()                          // pricekaj se se server i klijent zagase
	time.Sleep(100 * time.Millisecond) // pricekaj malo da se ociste konekcije
	assert.Len(t, b.subscribers, 0)    // nema subscribera client se odspojio
}
