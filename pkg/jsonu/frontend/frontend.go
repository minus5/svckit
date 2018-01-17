package frontend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"pkg/util"
	"time"

	"github.com/minus5/svckit/log"

	"golang.org/x/net/websocket"
	"math"
)

const (
	msgSubscribe = "subscribe"
	msgRead      = "read"
	msgPing      = "ping"
	msgJezik     = "jezik"
	msgAppVersion= "app_version"
	msgLog       = "log"
	msgStat      = "stat"
)

type Frontend struct {
	url        string
	origin     string
	app        string
	ver        string
	ws         *websocket.Conn
	retry      int
	jezik      string
	counter    int64
	subs       map[string]bool
	closed     chan struct{}
	Output     chan *Envelope
	input      chan *inMsg
	retryStore map[string]*inMsg
	retryCount map[string]int
}

func New(url, origin string, lang string, app string, ver string) *Frontend {
	fmt.Printf("Frontend new, lang=%s.\n",lang)
	url = url + "?conn_id=" + util.Uuid()[0:4]
	f := &Frontend{
		url:        url,
		origin:     origin,
		app:        app,
		ver:        ver,
		retry:      0,
		jezik:      lang,
		subs:       make(map[string]bool),
		closed:     make(chan struct{}),
		Output:     make(chan *Envelope, 1024),
		input:      make(chan *inMsg, 1024),
		retryStore: make(map[string]*inMsg),
		retryCount: make(map[string]int),
	}
	go f.loop()
	return f
}

func (f *Frontend) Close() {
	close(f.closed)
}

func (f *Frontend) loop() {
	for {
		select {
		case <-f.closed:
			fmt.Println("loop <-f.closed")
			close(f.Output) //zatvorimo Output je izlazimo van
			return
		default:
			fmt.Println("loop default, ide connect")
			if f.retry != 0 {
				fmt.Println("loop delay")
				delay := time.Duration(math.Pow(2, math.Min(float64(f.retry),13)))
				time.Sleep(delay*time.Millisecond)
			}
			f.retry++
			err := f.connect()
			if err == nil {
				fmt.Println("connect ok.")
				//f.Output = make(chan *Envelope, 1024)
				f.retry = 0
				f.Jezik(f.jezik)
				f.AppVersion()
				f.sendSubs()
				go f.sendLoop()
				f.listen()
			} else {
				fmt.Println(err)
			}
		}
	}
}

func (f *Frontend) connect() error {
	var err error
	config, _ := websocket.NewConfig(f.url, f.origin)
	ws, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}
	f.ws = ws
	return nil
}

func (f *Frontend) sendLoop() {
	for m := range f.input {
		if err := websocket.Message.Send(f.ws, m.Data()); err != nil {
			log.Error(err)
			f.ws.Close()
			return
		}
		// TODO izbaciti logiranje
		fmt.Printf("send: %s\n", m.Data())
	}
}

func (f *Frontend) getBody(path string) []byte {
	for i := 0; i < 10; i++ {
		url := f.origin + "/msgs/" + path
		client := http.Client{
			Timeout: time.Duration(5 * time.Second),
		}
		rsp, err := client.Get(url)
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second)
			continue
		}
		defer rsp.Body.Close()
		if rsp.StatusCode != 200 {
			log.S("url", url).I("status", rsp.StatusCode).
				ErrorS("getBody failed")
			continue
		}
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Error(err)
			continue
		}
		log.S("url", url).I("length", len(body)).Info("getBody")
		return body
	}
	return nil
}

func (f *Frontend) listen() {
	in := make(chan *Envelope)
	go func() {
		for {
			var frame []byte
again:
			var buf []byte
			if err := websocket.Message.Receive(f.ws, &buf); err != nil {
				log.Error(err)
				f.ws.Close()
				close(in)
				return
			}
			frame = append(frame, buf...)
			//FIXME: UGLY HACK
			//frameovi veci od 4096 se lome na blokove od 4096 pa ih skupljam prije obrade. ne kuzim zasto.
			if len(buf) == 4096 {
				fmt.Printf("truncated frame %d bytes, idem po ostatak.\n", len(buf))
				goto again
			}
			//fmt.Printf("frame: len %d\n", len(frame))
			//fmt.Printf("data: %s\n", string(frame))
			m, err := NewEnvelope(frame)
			if err != nil {
				continue
			}
			//fmt.Printf("header: len %d, \"%s\"\n", len(m.Header), m.Header)
			//fmt.Printf("body: len %d\n", len(m.Body))
			if m.BodyPath == "" {
				in <- m
			} else {
				go func() {
					if b := f.getBody(m.BodyPath); b != nil {
						m.Body = b
						in <- m
					}
				}()
			}
		}
	}()

	for {
		select {
		case m := <-in:
			if m == nil {
				fmt.Println("listen - in kanal zatvoren.")
				//ne zatvaramo output kanal jer odmah ide reconnect
				//close(f.Output)
				return
			}
			if _, ok := f.retryStore[m.Id]; ok {
				if m.Status != "0" { //poruka je failala a zatrazen je retry
					if f.retryCount[m.Id] < 5 {
						fmt.Printf("Msg %s failed, will retry.\n", m.Id)
						delay := time.Duration(math.Pow(2, math.Min(float64(f.retryCount[m.Id]), 13)))
						f.retryCount[m.Id]=f.retryCount[m.Id]+1
						go func(d time.Duration, msgId string) {
							time.Sleep(d * time.Millisecond) //sleepamo u goroutini da ne blokiramo
							f.input <- f.retryStore[msgId]
						}(delay, m.Id)
						continue //ne saljemo poruku na output ako je failala
					} else {
						fmt.Printf("Msg %s retried %d times, giving up.\n", m.Id, f.retryCount[m.Id])
					}
				}
				delete(f.retryCount,m.Id) //uspjesno obradjena ili odustajem, brisem retry...
				delete(f.retryStore, m.Id) //...i brisem poruku iz storea
			}
			f.Output <- m
		case <-time.After(50 * time.Second):
			f.Ping()
		case <-f.closed:
			f.ws.Close()
		}
	}
}

func (f *Frontend) Subscribe(key string) {
	f.subs[key] = true
	f.sendSubs()
}

func (f *Frontend) Unsubscribe(key string) {
	f.subs[key] = false
	f.sendSubs()
}

func (f *Frontend) Ping() {
	f.send(msgPing, nil, false)
}

func (f *Frontend) Jezik(jezik string) int64 {
	f.jezik = jezik
	msg := &struct {
		Jezik string `json:"jezik"`
	}{Jezik: jezik}
	counter := f.counter //treba nam da bi znali cekati odgovor na ovu poruku
	f.send(msgJezik, msg, true)
	return counter
}

func (f *Frontend) AppVersion() {
	msg := &struct {
		App string  `json:"app"`
		Version string `json:"version"`
	}{App: f.app, Version: f.ver}
	f.send(msgAppVersion, msg, true)
}

func (f *Frontend) Log(message interface{}) {
	f.send(msgLog, message, false)
}

func (f *Frontend) Stat(event string) {
	msg := &struct {
		Event string  `json:"event"`
	}{Event: event}
	f.send(msgStat, msg, false)
}

func (f *Frontend) Read(key string) {
	msg := &struct {
		Type string `json:"type"`
	}{Type: key}
	f.send(msgRead, msg, true)
}

func (f *Frontend) sendSubs() {
	msg := &struct {
		Subs []string `json:"subscriptions"`
	}{}
	for k, v := range f.subs {
		if v {
			msg.Subs = append(msg.Subs, k)
		}
	}
	f.send(msgSubscribe, msg, false)
}

func (f *Frontend) send(typ string, o interface{}, retry bool) error {
	f.counter++
	i := &inMsg{
		Type: typ,
		No:   f.counter,
	}
	if o != nil {
		buf, err := json.Marshal(o)
		if err != nil {
			return err
		}
		i.Body = string(buf)
	}
	if retry {
		f.retryStore[fmt.Sprintf("%d",f.counter)]=i
		f.retryCount[fmt.Sprintf("%d",f.counter)]=0
	}
	f.input <- i
	return nil
}

type inMsg struct {
	Type string
	No   int64
	Body string
}

func (i inMsg) Data() string {
	return fmt.Sprintf("%s:%d\n%s", i.Type, i.No, i.Body)
}
