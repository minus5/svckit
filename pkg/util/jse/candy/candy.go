package candy

import (
	"fmt"

	cjs "github.com/mcuadros/go-candyjs"
)

type Candy struct {
	c   *cjs.Context
	src string
}

func New(src []byte) (*Candy, error) {
	return &Candy{
		c:   cjs.NewContext(),
		src: string(src),
	}, nil
}

func (t *Candy) Eval(src string) (interface{}, error) {
	if err := t.c.PevalString(t.src); err != nil {
		return nil, err
	}
	if err := t.c.PevalString(src); err != nil {
		return nil, err
	}
	s := t.c.GetString(-1)
	if s != "" {
		return s, nil
	} else {
		return fmt.Sprint(t.c.GetNumber(-1)), nil
	}
}
