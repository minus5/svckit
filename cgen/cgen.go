//package cgen is code generation library
package cgen

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/fatih/structtag"
)

type Data struct {
	Package string
	Structs []Struct
}
type Struct struct {
	Type         string
	StructFields []Field
	Fields       []Field
	Maps         []Map
}
type Field struct {
	Name string
	Type string
	Tag  string
}
type Map struct {
	Field string
	Key   string
	Value string
	Tag   string
}

// NilConditions is helper method for building value.empty() method in template.
func (s Struct) NilConditions() []string {
	var c []string
	for _, f := range s.Fields {
		c = append(c, fmt.Sprintf(" i.%s == nil ", f.Name))
	}
	for _, s := range s.StructFields {
		c = append(c, fmt.Sprintf(" i.%s == nil ", s.Name))
	}
	for _, m := range s.Maps {
		c = append(c, fmt.Sprintf(" (i.%s == nil || len(i.%s) == 0) ", m.Field, m.Field))
	}
	return c
}

type structs map[string]Struct

func (s structs) sort() []Struct {
	var names []string
	for k := range s {
		names = append(names, k)
	}
	sort.Strings(names)
	var a []Struct
	for _, k := range names {
		a = append(a, s[k])
	}
	return a
}

// AnalyzeStruct returns information for all struct types
// reachable from o.
func AnalyzeStruct(o interface{}) Data {
	v := reflect.ValueOf(o).Elem()
	return Data{
		Package: packageName(v),
		Structs: analyzeStructs(deepFindValues(v)).sort(),
	}
}

func packageName(v reflect.Value) string {
	p := strings.Split(v.Type().PkgPath(), "/")
	return p[len(p)-1]
}

// deepFindValues finds all values recursively,
// staring with v and going into inner fields, maps.
func deepFindValues(v reflect.Value) map[string]reflect.Value {
	values := make(map[string]reflect.Value)
	findValues(v, values)
	return values
}

func findValues(v reflect.Value, values map[string]reflect.Value) {
	t := v.Type()
	values[t.Name()] = v

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() { // skip unexported
			continue
		}
		ft := t.Field(i)
		switch ft.Type.Kind() {
		case reflect.Map:
			mv := reflect.New(ft.Type.Elem()).Elem() // map value value
			findValues(mv, values)
		case reflect.Struct:
			ts := ft.Type.String()
			if ts == "sync.Mutex" {
				continue
			}
			if ts == "time.Time" {
				continue
			}
			sv := reflect.New(ft.Type).Elem()
			findValues(sv, values)
		}
	}
}

func analyzeStructs(values map[string]reflect.Value) structs {
	stcs := make(map[string]Struct)
	for k, v := range values {
		stcs[k] = analyzeStruct(v)
	}
	return stcs
}

func analyzeStruct(v reflect.Value) Struct {
	t := v.Type()
	stc := Struct{
		Type: t.Name(),
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() { // skip unexported
			continue
		}
		ft := t.Field(i)
		switch ft.Type.Kind() {
		case reflect.Map:
			vt := ft.Type.Elem() // map value
			if vt.Kind() == reflect.Ptr {
				vt = vt.Elem()
			}
			kt := ft.Type.Key() // map key
			stc.Maps = append(stc.Maps,
				Map{
					Field: ft.Name,
					Key:   kt.Name(),
					Value: vt.Name(),
					Tag:   parseTag(ft),
				})
			continue
		case reflect.Invalid, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
			// skip this types
			continue
		case reflect.Array, reflect.Slice:
			// slice !!!
			continue
		case reflect.Struct:
			ts := ft.Type.String()
			if ts == "sync.Mutex" {
				continue
			}
			if ts != "time.Time" {
				stc.StructFields = append(stc.StructFields, Field{
					Name: ft.Name,
					Type: ft.Type.Name(),
					Tag:  parseTag(ft),
				})
				continue
			}
		}
		stc.Fields = append(stc.Fields,
			Field{
				Name: ft.Name,
				Type: f.Type().String(),
				Tag:  parseTag(ft),
			})
	}
	return stc
}

func parseTag(ft reflect.StructField) string {
	tags, err := structtag.Parse(string(ft.Tag))
	if err != nil {
		panic(err)
	}
	jn := nonExported(ft.Name)
	if t, err := tags.Get("json"); err == nil {
		jn = t.Name
	}
	return "`" + `json:"` + jn + `,omitempty"` + "`"
}

// nonExported name; first letter lower
func nonExported(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}
