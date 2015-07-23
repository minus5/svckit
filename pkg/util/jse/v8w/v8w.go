package v8w

import (
	"fmt"

	"github.com/ry/v8worker"
)

type V8W struct {
	out    chan interface{}
	worker *v8worker.Worker
}

func New(buf []byte) (*V8W, error) {
	out := make(chan interface{})
	v := &V8W{
		out: out,
		worker: v8worker.New(func(msg string) {
			go func() { out <- msg }()
		}),
	}
	if err := v.worker.Load("script.js", string(buf)); err != nil {
		return nil, err
	}
	return v, nil
}

func (t *V8W) Eval(src string) (interface{}, error) {
	if err := t.worker.Load("src.js", fmt.Sprintf("var res = %s; $send(String(res))", src)); err != nil {
		return nil, err
	}
	res := <-t.out
	return res, nil
}

func V8Version() string {
	return v8worker.Version()
}
