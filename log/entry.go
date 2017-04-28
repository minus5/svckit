package log

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	layout = `2006-01-02T15:04:05.999999-07:00`
)

// Entry represents a parsed log line
type Entry struct {
	Time     time.Time
	Host     string
	App      string
	File     string
	Level    string
	Msg      string
	Original []byte
	attr     map[string]interface{}
}

// NewEntry constructs an Entry from a  raw log line
func NewEntry(b []byte) (*Entry, error) {
	e := &Entry{
		attr:     map[string]interface{}{},
		Original: b,
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(b, &m); err != nil {
		if strings.Contains(err.Error(), "hexadecimal character escape") {
			// popravi parsing kod ove greske
			b = []byte(strings.Replace(string(b), `\u0`, "", -1))
			err := json.Unmarshal(b, &m)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			switch k {
			case "time":
				tm, err := time.Parse(layout, vv)
				if err != nil {
					return nil, err
				}
				e.Time = tm
			case "host":
				e.Host = vv
			case "app":
				e.App = vv
			case "file":
				e.File = vv
			case "level":
				e.Level = vv
			case "msg":
				e.Msg = vv
			default:
				e.attr[k] = vv
			}
		default:
			e.attr[k] = vv
		}
	}
	return e, nil
}

// I retrieves an integer attribute by key
func (e *Entry) I(key string) (int, bool) {
	if val, ok := e.attr[key]; !ok {
		return 0, false
	} else {
		switch t := val.(type) {
		case int:
			return t, true
		case float64:
			return int(t), true
		default:
			return 0, false
		}
	}
}

// F retrieves an float64 attribute by key
func (e *Entry) F(key string) (float64, bool) {
	if val, ok := e.attr[key]; !ok {
		return 0, false
	} else {
		switch t := val.(type) {
		case int:
			return float64(t), true
		case float64:
			return t, true
		default:
			return 0, false
		}
	}
}

// S retrieves a string attribute by key
func (e *Entry) S(key string) (string, bool) {
	if val, ok := e.attr[key]; !ok {
		return "", false
	} else {
		valStr, ok := val.(string)
		return valStr, ok
	}
}
