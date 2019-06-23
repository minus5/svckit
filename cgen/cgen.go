//package cgen is code generation library
package cgen

import (
	"fmt"
	"reflect"
	"sort"
	"unicode"

	"github.com/fatih/structtag"
)

type Struct struct {
	Type    string
	Structs []Field
	Fields  []Field
	Maps    []Map
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

func (s Struct) NilConditions() []string {
	var c []string
	for _, f := range s.Fields {
		c = append(c, fmt.Sprintf(" i.%s == nil ", f.Name))
	}
	for _, s := range s.Structs {
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
func AnalyzeStruct(o interface{}) []Struct {
	v := reflect.ValueOf(o).Elem()
	return analyzeValues(deepFindValues(v)).sort()
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

func analyzeValues(values map[string]reflect.Value) structs {
	stcs := make(map[string]Struct)
	for k, v := range values {
		stcs[k] = analyzeValue(v)
	}
	return stcs
}

func analyzeValue(v reflect.Value) Struct {
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
				stc.Structs = append(stc.Structs, Field{
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

/*
type translatorInterface interface {
	Lang(string) string
}

func save(file string) error {
	err := ioutil.WriteFile(file, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	err = exec.Command("go", "fmt", file).Run()
	if err != nil {
		return fmt.Errorf("go fmt failed with error: %s", err)
	}
	err = exec.Command("goimports", "-w", file).Run()
	if err != nil {
		return fmt.Errorf("go imports failed with error: %s", err)
	}
	fmt.Printf("generated %s\n", file)
	return nil
}

// nonExported name; first letter lower
func nonExported(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

func removePackagePrefix(typ string) string {
	p := strings.Split(typ, "/")
	return p[len(p)-1]
}

type templateData struct {
	Type             string
	TypeLower        string
	Fields           []fieldData
	Maps             []mapData
	Structs          []structData
	EmptyParts       []string
	HasChangedMethod bool
	HasTsMethod      bool
	HasParent        bool
	ParentName       string
	ParentType       string
}

type fieldData struct {
	Name             string
	Type             string
	ClientType       string
	IsClientType     bool
	NameLower        string
	Tag              string
	HasChangedMethod bool
	IsStruct         bool
	IsSlice          bool
	TagName          string
}

func (d fieldData) NeedUnpack() bool {
	return d.TagName != d.NameLower
}

type mapData struct {
	Field         string
	FieldLower    string
	Key           string
	Value         string
	KeyType       reflect.Type
	ValueType     reflect.Type
	ValueValue    reflect.Value
	HasHiddenAttr bool
	HasParent     bool
	HasKeyAttr    bool
	ParentAttr    string
	Tag           string
	TagName       string
}

func (d mapData) NeedUnpack() bool {
	return d.TagName != d.FieldLower
}

type structData struct {
	Field      string
	FieldLower string
	Type       string
	FieldType  reflect.Type
	FieldValue reflect.Value
	HasParent  bool
	HasKeyAttr bool
	ParentAttr string
	Tag        string
	TagName    string
}

func (d structData) NeedUnpack() bool {
	return d.TagName != d.FieldLower
}

func parseTag(tag string, jn string, tagName string) (bool, string, string) {
	tags, err := structtag.Parse(string(tag))
	if err != nil {
		panic(err)
	}
	if t, err := tags.Get(tagName); err == nil && t.Name == "-" {
		return true, "", ""
	}
	if t, err := tags.Get("json"); err == nil {
		jn = t.Name
	}
	mn := jn
	if t, err := tags.Get("msg"); err == nil {
		mn = t.Name
	}
	return false, "`" + `json:"` + jn + `,omitempty" bson:"` + jn + `,omitempty" msg:"` + mn + `"` + "`", jn
}

func parentName(t reflect.Type) string {
	n := nonExported(t.Name())
	n = strings.TrimSuffix(n, "Diff")
	n = strings.TrimSuffix(n, "Client")
	return n
}

func getTemplateData(t reflect.Type, v reflect.Value, tagName string,
	pt reflect.Type) templateData {
	translatorType := reflect.TypeOf((*translatorInterface)(nil)).Elem()
	exported, maps, isStruct, hasChangedMethod, hasTsMethod, structs, isSlice := findFields(t, v)
	d := templateData{
		Type:             t.Name(),
		TypeLower:        nonExported(t.Name()),
		HasChangedMethod: hasMethod(t, "Changed"),
		HasTsMethod:      hasTsMethod,
	}
	if pt != nil {
		d.HasParent = true
		d.ParentName = parentName(pt)
		d.ParentType = pt.Name()
	}
	for _, i := range exported {
		f := v.Field(i)
		ft := t.Field(i)
		nl := nonExported(ft.Name)

		skip, tag, jn := parseTag(string(ft.Tag), nl, tagName)
		if skip {
			continue
		}
		typ := strings.TrimPrefix(f.Type().String(), pkg+".")
		ltyp := typ
		if f.Type().Implements(translatorType) {
			ltyp = "string"
		}
		d.Fields = append(d.Fields, fieldData{
			Name:             ft.Name,
			Type:             typ,
			ClientType:       ltyp,
			IsClientType:     typ != ltyp,
			NameLower:        nl,
			HasChangedMethod: hasChangedMethod[i],
			IsStruct:         isStruct[i],
			IsSlice:          isSlice[i],
			Tag:              tag,
			TagName:          jn,
		})
		p := fmt.Sprintf(" o.%s == nil ", ft.Name)
		d.EmptyParts = append(d.EmptyParts, p)
	}
	for _, i := range maps {
		ft := t.Field(i)
		et := ft.Type.Elem() // map value
		if et.Kind() == reflect.Ptr {
			et = et.Elem()
		}
		kt := ft.Type.Key() // map key
		_, hasHiddenAttr := et.FieldByName("hidden")
		pn := parentName(t)
		_, hasParent := et.FieldByName(pn)
		_, hasKeyAttr := et.FieldByName("key")
		nl := nonExported(ft.Name)
		if !hasParent {
			pn = ""
		}
		skip, tag, jn := parseTag(string(ft.Tag), nl, tagName)
		if skip {
			continue
		}
		d.Maps = append(d.Maps, mapData{
			Field:         ft.Name,
			FieldLower:    nl,
			Key:           kt.Name(),
			Value:         et.Name(),
			ValueType:     et,
			KeyType:       kt,
			ValueValue:    reflect.New(et).Elem(),
			HasHiddenAttr: hasHiddenAttr,
			HasParent:     hasParent,
			HasKeyAttr:    hasKeyAttr,
			ParentAttr:    pn,
			Tag:           tag,
			TagName:       jn,
		})
		p := fmt.Sprintf(" (o.%s == nil || len(o.%s) == 0) ", ft.Name, ft.Name)
		d.EmptyParts = append(d.EmptyParts, p)
	}
	for _, i := range structs {
		ft := t.Field(i)
		et := ft.Type // map value
		//_, hasHiddenAttr := et.FieldByName("hidden")
		pn := parentName(t)
		_, hasParent := et.FieldByName(pn)
		_, hasKeyAttr := et.FieldByName("key")
		nl := nonExported(ft.Name)
		if !hasParent {
			pn = ""
		}
		skip, tag, jn := parseTag(string(ft.Tag), nl, tagName)
		if skip {
			continue
		}
		//typ := strings.TrimPrefix(et.String(), pkg+".")
		d.Structs = append(d.Structs, structData{
			Field:      ft.Name,
			FieldLower: nl,
			Type:       et.Name(),
			FieldType:  et,
			FieldValue: reflect.New(et).Elem(),
			HasParent:  hasParent,
			HasKeyAttr: hasKeyAttr,
			ParentAttr: pn,
			Tag:        tag,
			TagName:    jn,
		})
		p := fmt.Sprintf(" o.%s == nil ", ft.Name)
		d.EmptyParts = append(d.EmptyParts, p)
	}
	return d
}

func hasMethod(t reflect.Type, name string) bool {
	nt := reflect.New(t)
	if v := nt.MethodByName(name); v.IsValid() {
		return true
	}
	if _, ok := t.MethodByName(name); ok {
		return true
	}
	return false
}
*/
