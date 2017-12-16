package api

import (
	"reflect"

	"github.com/minus5/svckit/types/ers"
	"github.com/minus5/svckit/types/registry"
)

type addReq struct {
	X int
	Y int
}

type addRsp struct {
	Z int
}

var (
	typeRegistry = registry.New()
	appErrors    = ers.New()
	ErrOverflow  = appErrors.New("owerflow")
	ErrTransport = appErrors.New("transport")
)

func init() {
	typeRegistry.Add([]interface{}{
		addReq{},
	})
}

func NameFor(i interface{}) string {
	return typeRegistry.NameFor(i)
}

func TypeFor(typ string) reflect.Type {
	return typeRegistry.TypeFor(typ)
}

func ParseError(text string) error {
	return appErrors.Parse(text)
}
