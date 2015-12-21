package nsqu

import (
	"sync"

	"github.com/bitly/go-nsq"
)

var (
	NsqReaderDefaultConcurency = 256
)

//NsqReader cita poruke iz nsq-a i pise ih na Messages kanal
//Omogucava jednostavno koristenje u select ili range petlji.
//implementira clean stop (pozivo na Stop).
//Svaku procitanu poruku potrebno je potvrditi da je ok obradjena ili ne (Ok metoda na NsqReaderMessage).
//Ako nije obradjena vratit ce se u nsq.
type NsqReader struct {
	closeCh          chan struct{}
	Messages         chan *NsqReaderMessage
	onceClose        sync.Once
	topic            string
	channel          string
	lookupdHTTPAddrs []string
	concurency       int
}

type NsqReaderMessage struct {
	Body []byte
	done chan error
}

func NewNsqReaderMessage(body []byte) *NsqReaderMessage {
	return &NsqReaderMessage{
		Body: body,
		done: make(chan error),
	}
}

//Ok - poruka je uspjesno obradjena
func (m *NsqReaderMessage) Ok() {
	m.done <- nil
}

//Done - ako je err == nil poruka je uspjesno obradjena Ok, inace Fail
func (m *NsqReaderMessage) Done(err error) {
	m.done <- err
}

//Fail - poruka nije uspjesno obradjena vrati ju u nsq
func (m *NsqReaderMessage) Fail(err error) {
	m.done <- err
}

func (m *NsqReaderMessage) waitDone() error {
	return <-m.done
}

func NewNsqReader(topic, channel string, lookupdHTTPAddrs []string, concurency int) *NsqReader {
	r := &NsqReader{
		topic:            topic,
		channel:          channel,
		lookupdHTTPAddrs: lookupdHTTPAddrs,
		Messages:         make(chan *NsqReaderMessage),
		closeCh:          make(chan struct{}),
		concurency:       concurency,
	}
	go r.loop()
	return r
}

func (r *NsqReader) loop() {
	conn := NewConnection(r.lookupdHTTPAddrs, "")
	go func() {
		var c *nsq.Consumer
		c = conn.MustNewConsumer(r.topic, r.channel, r.concurency, func(buf []byte) error {
			msg := NewNsqReaderMessage(buf)
			r.Messages <- msg
			return msg.waitDone()
		})
		go func() {
			<-r.closeCh
			c.Stop()
		}()
		<-c.StopChan
		close(r.Messages)
	}()

}

func (r *NsqReader) Stop() {
	r.onceClose.Do(func() {
		close(r.closeCh)
	})
}
