/*package registry implements type reqistry.
 */
package registry

import (
	"reflect"
	"strings"
	"sync"
)

type Registry struct {
	r map[string]reflect.Type
	sync.Mutex
}

func New() *Registry {
	return &Registry{
		r: make(map[string]reflect.Type),
	}
}

func (r *Registry) AddKey(k string, v interface{}) {
	r.Lock()
	defer r.Unlock()
	r.r[k] = reflect.TypeOf(v)
}

func (r *Registry) Add(types []interface{}) {
	r.Lock()
	defer r.Unlock()
	for _, typ := range types {
		t := reflect.TypeOf(typ)
		r.r[typeToString(typ)] = t
	}
}

func (r *Registry) Find(k string) (reflect.Type, bool) {
	r.Lock()
	defer r.Unlock()
	i, ok := r.r[k]
	return i, ok
}

func typeToString(i interface{}) string {
	typ := reflect.TypeOf(i).String()
	if strings.HasPrefix(typ, "*") {
		typ = typ[1:]
	}
	return typ
}

func (r *Registry) NameFor(i interface{}) string {
	return typeToString(i)
}

func (r *Registry) TypeFor(k string) reflect.Type {
	t, _ := r.Find(k)
	return t
}
