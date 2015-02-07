package jsonu

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/bitly/go-simplejson"
)

func ArrayToObject(arrayKey, idKey string, j *simplejson.Json) {
	o := make(map[string]interface{})
	for i, _ := range j.Get(arrayKey).MustArray() {
		v := j.Get(arrayKey).GetIndex(i)
		key := fmt.Sprintf("%d", i)
		if idKey != "" {
			key = ValueToString(v.Get(idKey))
		}
		o[key] = v.Interface()
	}
	if len(o) > 0 {
		j.Set(arrayKey, o)
	} else {
		j.Del(arrayKey)
	}
}

func ValueToString(j *simplejson.Json) string {
	switch v := j.Interface().(type) {
	case string:
		return v
	case json.Number:
		return fmt.Sprintf("%d", j.MustInt64())
	default:
		log.Fatalf("nepoznti tip type: %T value: %v", v, v)
		return ""
	}
}

//unused
func delIfEmptyObject(e *simplejson.Json, key string) {
	if _, err := e.Map(); err != nil {
		return
	}
	if len(e.MustMap()) == 0 {
		e.Del(key)
	}
}

func Rename(e *simplejson.Json, key, newKey string) {
	if v, ok := e.CheckGet(key); ok {
		e.Set(newKey, v)
		e.Del(key)
	}
}

func StrToInt(s *simplejson.Json, key string) {
	if v, err := strconv.Atoi(s.Get(key).MustString()); err == nil {
		s.Set(key, v)
	} else {
		log.Printf("error strToInt %s key %s %s %#v", err, key, s.Get(key).MustString(), s)
	}
}

func Empty(o *simplejson.Json) bool {
	return len(o.MustMap()) == 0
}

func MapToSimplejson(m map[string]interface{}) *simplejson.Json {
	j := simplejson.New()
	j.Set("__key", m)
	return j.Get("__key")
}

func MarshalPretty(i interface{}) ([]byte, error) {
	return json.MarshalIndent(i, "  ", "  ")
}
