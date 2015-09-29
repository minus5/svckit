package jsonu

import (
	"encoding/json"

	"github.com/bitly/go-simplejson"
)

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

func MarshalPrettyBuf(buf []byte) ([]byte, error) {
	var data map[string]interface{}
	json.Unmarshal(buf, &data)
	return json.MarshalIndent(data, "", "  ")

}
