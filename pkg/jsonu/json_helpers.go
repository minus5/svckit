package jsonu

import (
	"encoding/json"
	"log"

	"github.com/minus5/go-simplejson"
)

func Empty(o *simplejson.Json) bool {
	return len(o.MustMap()) == 0
}

func MapToSimplejson(m map[string]interface{}) *simplejson.Json {
	return simplejson.NewFromMap(m)
}

func Sprint(i interface{}) string {
	buf, err := MarshalPretty(i)
	if err != nil {
		return ""
	}
	return string(buf)
}

func MarshalPretty(i interface{}) ([]byte, error) {
	return json.MarshalIndent(i, "  ", "  ")
}

func MarshalPrettyBuf(buf []byte) ([]byte, error) {
	var data map[string]interface{}
	json.Unmarshal(buf, &data)
	return json.MarshalIndent(data, "", "  ")
}

func Marshal(i interface{}) []byte {
	if i != nil {
		j, err := json.Marshal(i)
		if err != nil {
			log.Println(err)
			return []byte{}
		}
		return j
	}
	return []byte{}
}
