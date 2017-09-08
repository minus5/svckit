package keys

import (
	"sync"

	simplejson "github.com/minus5/go-simplejson"
)

func NewExtender(data *simplejson.Json) *Extender {
	return &Extender{
		data: data,
	}
}

type Extender struct {
	data *simplejson.Json
	sync.Mutex
}

func (e *Extender) ExtendWith(km map[string]string) *simplejson.Json {
	e.Lock()
	defer e.Unlock()
	return extend(e.data, km)
}

func extend(orig *simplejson.Json, km map[string]string) *simplejson.Json {
	data := simplejson.New()
	for key, _ := range orig.MustMap() {
		v := orig.Get(key)
		if k, ok := km[key]; ok {
			key = k
		}
		switch v.Interface().(type) {
		case *map[string]interface{}, map[string]interface{}, *simplejson.Json:
			data.Set(key, extend(v, km))
		default:
			data.Set(key, v)
		}
	}
	if len(data.MustMap()) == 0 {
		return orig
	}
	return data
}
