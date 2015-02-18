package jsonu

import (
	"encoding/json"
	"log"

	"github.com/bitly/go-simplejson"
)

func Diff(left, right *simplejson.Json) *simplejson.Json {
	return diffObject(left, right)
}

func diff(bufL, bufR []byte) []byte {
	left, _ := simplejson.NewJson(bufL)
	right, _ := simplejson.NewJson(bufR)
	diff := diffObject(left, right)
	ret, _ := diff.Encode()
	return ret
}

//diffObject - vraca diff dvaju hash-eva
//lijevi je stari, desni je novi, diff nadogradjuje lijevi na desni
func diffObject(left, right *simplejson.Json) *simplejson.Json {
	diff := simplejson.New()
	//postoji left a ne postoji right, postavi na null u diff
	for k, _ := range left.MustMap() {
		if sameKey(k, left, right) == rightMissing {
			diff.Set(k, nil)
		}
	}
	//one koji u right imaju drugaciju vrijednost nego u left
	//ili ih nije bilo u left dodaj u diff
	for k, _ := range right.MustMap() {
		switch sameKey(k, left, right) {
		case areDifferent, leftMissing:
			diff.Set(k, right.Get(k).Interface())
		case isObject:
			o := diffObject(left.Get(k), right.Get(k))
			if !Empty(o) {
				diff.Set(k, o.MustMap())
			}
		}
	}
	return diff
}

const (
	leftMissing = iota
	rightMissing
	areSame
	areDifferent
	isObject
)

func sameKey(k string, l *simplejson.Json, r *simplejson.Json) int {
	if _, ok := l.CheckGet(k); !ok {
		return leftMissing
	}
	if _, ok := r.CheckGet(k); !ok {
		return rightMissing
	}
	vl := l.Get(k)
	vr := r.Get(k)
	return sameValue(k, vl, vr)
}

func sameValue(k string, vl *simplejson.Json, vr *simplejson.Json) int {
	switch tl := vl.Interface().(type) {
	case nil:
		if vr.Interface() == nil {
			return areSame
		}
	case string:
		if tr, err := vr.String(); err == nil && tr == tl {
			return areSame
		}
	case bool:
		if tr, err := vr.Bool(); err == nil && tr == tl {
			return areSame
		}
	case json.Number, float64:
		if vl.MustFloat64() == vr.MustFloat64() {
			return areSame
		}
	case int:
		if vl.MustInt() == vr.MustInt() {
			return areSame
		}
	case int64:
		if vl.MustInt64() == vr.MustInt64() {
			return areSame
		}
	case []interface{}:
		al := vl.MustArray()
		ar := vr.MustArray()
		if len(ar) != len(al) {
			return areDifferent
		}
		for i, _ := range al {
			if sameValue(k, vl.GetIndex(i), vr.GetIndex(i)) != areSame {
				return areDifferent
			}
		}
		return areSame
	case []int:
		al, okl := vl.Interface().([]int)
		ar, okr := vr.Interface().([]int)
		if !okl || !okr {
			return areDifferent
		}
		if len(ar) != len(al) {
			return areDifferent
		}
		for i, _ := range al {
			if al[i] != ar[i] {
				return areDifferent
			}
		}
		return areSame
		//uspredba serijalizacijom u json
	case []int64, [2][2]int, [][2]int:
		bl, _ := vl.Encode()
		br, _ := vr.Encode()
		if string(bl) == string(br) {
			return areSame
		}
	case map[string]interface{}, *simplejson.Json:
		return isObject
	default:
		log.Fatalf("nepoznti tip k: %s type: %T value: %v", k, tl, tl)
	}
	return areDifferent
}

//JsonMerge - merge keys from diff map into m map
func JsonMerge(m map[string]interface{}, d map[string]interface{}) {
	for k, v := range d {
		switch v.(type) {
		case map[string]interface{}:
			if m[k] == nil {
				m[k] = d[k]
			} else {
				if im, ok := m[k].(map[string]interface{}); ok {
					JsonMerge(im, d[k].(map[string]interface{}))
				}
			}
		case nil:
			delete(m, k)
		default:
			m[k] = d[k]
		}
	}
}

func DeepCopyMap(m map[string]interface{}) map[string]interface{} {
	r := make(map[string]interface{})
	for k, v := range m {
		switch v.(type) {
		case map[string]interface{}:
			r[k] = DeepCopyMap(v.(map[string]interface{}))
		default:
			r[k] = m[k]
		}
	}
	return r
}
