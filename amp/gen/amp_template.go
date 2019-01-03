package gen

import (
	"text/template"
)

var ampTemplate = template.Must(template.New("").Funcs(fns).Parse(`
import (
	"encoding/json"
	"pkg/amp"

	"github.com/minus5/svckit/log"
)

// satisfy amp.BodyMarshaler and amp.BodyUnmarshaler interfaces
// from pkg/amp package
func (i *{{.Type}}) ToJSON() ([]byte, error) {
  i.Lock()
  defer i.Unlock()
	return json.Marshal(i)
}

func (i *{{.Type}}) ToMsgp() ([]byte, error) {
	w := new(bytes.Buffer)
	mw := msgp.NewWriter(w)
	if err := i.EncodeMsg(mw); err != nil {
		return nil, err
	}
	mw.Flush()
	return w.Bytes(), nil
}

func (i *{{.Type}}) FromJSON(buf []byte) error {
	return json.Unmarshal(buf, i)
}

func (i *{{.Type}}) FromMsgp(buf []byte) error {
	return i.DecodeMsg(msgp.NewReader(bytes.NewReader(buf)))
}

func (i *{{.Type}}) ToLang(lang string) amp.BodyMarshaler {
	return i
}

func (i *{{.Type}}Diff) ToJSON() ([]byte, error) {
	return json.Marshal(i)
}

func (i *{{.Type}}Diff) ToMsgp() ([]byte, error) {
	w := new(bytes.Buffer)
	mw := msgp.NewWriter(w)
	if err := i.EncodeMsg(mw); err != nil {
		return nil, err
	}
	mw.Flush()
	return w.Bytes(), nil
}

func (i *{{.Type}}Diff) FromJSON(buf []byte) error {
	return json.Unmarshal(buf, i)
}

func (i *{{.Type}}Diff) FromMsgp(buf []byte) error {
	return i.DecodeMsg(msgp.NewReader(bytes.NewReader(buf)))
}

func (i *{{.Type}}Diff) ToLang(lang string) amp.BodyMarshaler {
	return i
}

func (full *{{.Type}}) Pipe(in <-chan amp.PublishMessage) <-chan amp.FullDiffMessage {
	out := make(chan amp.FullDiffMessage)
	go func() {
		for m := range in {
			diff := &{{.Type}}Diff{}
			if err := m.BodyTo(diff); err != nil {
				log.Error(err)
				continue
			}
			if m.Full() {
				full = &{{.Type}}{}
			}
      full.Lock()
			full.MergeDiff(diff)
      full.Unlock()
			out <- amp.NewFullDiffMessage(m.Stream(), m.No(), full, diff, m.Full())
		}
		close(out)
	}()
	return out
}

`))
