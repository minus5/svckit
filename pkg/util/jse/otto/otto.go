package otto

import (
	"github.com/robertkrimen/otto"
)

type Otto struct {
	o *otto.Otto
	s *otto.Script
}

func New(src []byte) (*Otto, error) {
	o := otto.New()
	s, err := o.Compile("", src)
	if err != nil {
		return nil, err
	}
	return &Otto{
		o: o,
		s: s,
	}, nil
}

func (t *Otto) Eval(src string) (interface{}, error) {
	if _, err := t.o.Run(t.s); err != nil {
		return nil, err
	}
	val, err := t.o.Run(src)
	if err != nil {
		return nil, err
	}
	return val.ToString()
}
